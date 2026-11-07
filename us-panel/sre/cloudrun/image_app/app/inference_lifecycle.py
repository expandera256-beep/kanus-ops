"""Durable, entitlement-bound start/stop controller for explicitly owned workers.

Metadata and audit reads never start workers. Provider calls happen outside SQL
transactions; persisted operation IDs make ambiguous starts/stops idempotent.
"""
import contextlib
import copy
import dataclasses
import json
import re
import threading
import time
import uuid
from storage import connect


class LifecycleError(RuntimeError):
    pass


@dataclasses.dataclass(frozen=True)
class WorkerPolicy:
    key: str
    model: str
    project: str
    zone: str
    instance: str
    instance_id: str
    upstream: str
    idle_seconds: int = 120
    boot_seconds: int = 480
    request_seconds: int = 600
    lease_seconds: int = 60
    capacity: int = 10
    max_uptime_seconds: int = 1800
    daily_runtime_seconds: int = 3600
    daily_starts: int = 6
    preserve_local_ssd: bool = True
    adopt_running: bool = False
    require_watchdog: bool = True

    def __post_init__(self):
        if not re.fullmatch(r'[a-z][a-z0-9-]{0,62}', self.key) or not self.model or len(self.model)>128:
            raise ValueError('invalid worker identity')
        if not isinstance(self.instance_id,str) or not self.instance_id.isdecimal() or not isinstance(self.upstream,str) or not self.upstream:
            raise ValueError('verified instance identity and upstream required')
        if any(type(v) is not bool for v in (self.preserve_local_ssd,self.adopt_running,self.require_watchdog)):raise ValueError('invalid lifecycle guard switch')
        for value,low,high in ((self.idle_seconds,0,1800),(self.boot_seconds,1,600),(self.request_seconds,1,600),(self.lease_seconds,5,120),(self.capacity,1,10),(self.max_uptime_seconds,60,86400),(self.daily_runtime_seconds,60,86400),(self.daily_starts,1,100)):
            if type(value) is not int or not low<=value<=high:raise ValueError('invalid lifecycle limit')
        if self.max_uptime_seconds < self.request_seconds+30:
            raise ValueError('uptime guard must leave room for an admitted request')


class LifecycleStore:
    def __init__(self, database):
        self.database=database
        with self.db() as db:
            db.executescript('''
                CREATE TABLE IF NOT EXISTS inference_workers (worker TEXT PRIMARY KEY, state TEXT NOT NULL);
                CREATE TABLE IF NOT EXISTS inference_events (event_id TEXT PRIMARY KEY,worker TEXT NOT NULL,created INTEGER NOT NULL,kind TEXT NOT NULL,detail TEXT NOT NULL);
                CREATE TABLE IF NOT EXISTS inference_watchdog (name TEXT PRIMARY KEY,heartbeat INTEGER NOT NULL);
            ''')

    @contextlib.contextmanager
    def db(self):
        db=connect(self.database)
        try:yield db
        finally:db.close()

    def change(self,key,fn):
        with self.db() as db:
            db.execute('BEGIN IMMEDIATE')
            try:
                row=db.execute('SELECT state FROM inference_workers WHERE worker=?',(key,)).fetchone()
                state=json.loads(row[0]) if row else {'phase':'unknown','leases':{},'op':None,'last_activity':0,'run_since':0,'billed_at':0,'day':0,'day_seconds':0,'starts':0,'error':''}
                result=fn(state)
                raw=json.dumps(state,separators=(',',':'),sort_keys=True)
                if len(raw)>65536:raise LifecycleError('worker state exceeds bound')
                if row:db.execute('UPDATE inference_workers SET state=? WHERE worker=?',(raw,key))
                else:db.execute('INSERT INTO inference_workers VALUES(?,?)',(key,raw))
                db.execute('COMMIT');return result
            except BaseException:
                db.execute('ROLLBACK');raise

    def view(self,key):
        with self.db() as db:
            row=db.execute('SELECT state FROM inference_workers WHERE worker=?',(key,)).fetchone()
            return json.loads(row[0]) if row else None

    def event(self,key,kind,detail,now):
        # No customer identity, prompt, response, token or provider exception.
        with self.db() as db:db.execute('INSERT INTO inference_events VALUES(?,?,?,?,?)',(uuid.uuid4().hex,key,int(now),kind,str(detail)[:300]))

    def watchdog_beat(self,now):
        with self.db() as db:
            db.execute('BEGIN IMMEDIATE')
            try:
                exists=db.execute('SELECT 1 FROM inference_watchdog WHERE name=?',('external',)).fetchone()
                if exists:db.execute('UPDATE inference_watchdog SET heartbeat=? WHERE name=?',(int(now),'external'))
                else:db.execute('INSERT INTO inference_watchdog VALUES(?,?)',('external',int(now)))
                db.execute('COMMIT')
            except BaseException:db.execute('ROLLBACK');raise

    def watchdog_fresh(self,now):
        with self.db() as db:
            row=db.execute('SELECT heartbeat FROM inference_watchdog WHERE name=?',('external',)).fetchone()
            return bool(row and 0<=now-row[0]<=120)


class LifecycleController:
    def __init__(self,store,provider,policies,authorize,clock=time.time,sleeper=time.sleep):
        self.store=store;self.provider=provider;self.policies={p.key:p for p in policies};self.authorize=authorize;self.clock=clock;self.sleep=sleeper
        if len(self.policies)!=len(policies) or len({p.model for p in policies})!=len(policies):raise ValueError('duplicate worker/model')

    def policy_for_model(self,model):
        for p in self.policies.values():
            if p.model==model:return p
        raise LifecycleError('model has no authorized worker')

    @staticmethod
    def _account(state,now):
        day=int(now)//86400
        elapsed=max(0,now-state.get('billed_at',now)) if state.get('billed_at') else 0
        if state.get('day')!=day:
            state['day']=day;state['day_seconds']=0;state['starts']=0
            elapsed=min(elapsed,now-day*86400)
        if state.get('run_since'):state['day_seconds']+=elapsed
        state['billed_at']=now
        state['leases']={k:v for k,v in state['leases'].items() if v['expires']>now and v['deadline']>now}

    def acquire(self,reservation,model,deadline=None):
        if not self.authorize(reservation):raise LifecycleError('unreserved request cannot wake a worker')
        p=self.policy_for_model(model);now=self.clock();deadline=deadline or now+p.request_seconds
        if not now<deadline<=now+p.request_seconds:raise LifecycleError('invalid request deadline')
        if p.require_watchdog and not self.store.watchdog_fresh(now):raise LifecycleError('external watchdog not verified')
        def add(s):
            self._account(s,now)
            if reservation in s['leases']:return copy.deepcopy(s['leases'][reservation])
            if len(s['leases'])>=p.capacity:raise LifecycleError('worker queue is full')
            if s['day_seconds']+deadline-now>p.daily_runtime_seconds:raise LifecycleError('daily runtime budget exhausted')
            if s['run_since'] and now-s['run_since']+deadline-now+30>p.max_uptime_seconds:raise LifecycleError('worker draining for uptime limit')
            s['leases'][reservation]={'deadline':deadline,'expires':min(deadline,now+p.lease_seconds),'created':now}
            s['last_activity']=now
            return copy.deepcopy(s['leases'][reservation])
        lease=self.store.change(p.key,add)
        self.store.event(p.key,'request_admitted','reserved model request',now)
        return p,lease

    def heartbeat(self,p,reservation):
        if not self.authorize(reservation):raise LifecycleError('request reservation is no longer active')
        now=self.clock()
        def beat(s):
            self._account(s,now)
            if reservation not in s['leases']:raise LifecycleError('request lease expired')
            lease=s['leases'][reservation];lease['expires']=min(lease['deadline'],now+p.lease_seconds)
        self.store.change(p.key,beat)

    def release(self,p,reservation):
        now=self.clock()
        def finish(s):
            self._account(s,now)
            if s['leases'].pop(reservation,None) is not None:s['last_activity']=now
        self.store.change(p.key,finish)

    def _observe(self,p,vm,expected_epoch=None):
        now=self.clock()
        if str(vm.get('id'))!=p.instance_id:raise LifecycleError('cloud instance identity changed')
        def update(s):
            # An HTTP describe begun before another controller claimed or
            # completed a power operation must not overwrite its newer state.
            if expected_epoch is not None and s.get('epoch',0)!=expected_epoch:
                result=copy.deepcopy(s);result['observation_stale']=True;return result
            self._account(s,now)
            status=vm['status'];s['provider_status']=status
            if status=='TERMINATED':
                if not s['op'] or s['op']['kind']!='start':s['run_since']=0
                if s['op'] and s['op']['kind']=='stop':s['op']=None;s['epoch']=s.get('epoch',0)+1
                if not s['op']:s['phase']='stopped'
            elif status=='RUNNING':
                if not s['run_since']:s['run_since']=vm.get('started_at') or now
                if s['op'] and s['op']['kind']=='start':s['op']=None;s['epoch']=s.get('epoch',0)+1
                if not s['op']:s['phase']='running'
            else:s['phase']=status.lower()
            return copy.deepcopy(s)
        return self.store.change(p.key,update)

    def _claim(self,p,kind,expected_epoch=None):
        now=self.clock()
        def claim(s):
            if expected_epoch is not None and s.get('epoch',0)!=expected_epoch:return None
            self._account(s,now)
            op=s['op']
            if op:
                if op['kind']!=kind or op['until']>now:return None
                op['until']=now+30;return dict(op)
            if kind=='start':
                if s.get('provider_status')!='TERMINATED':return None
                if not s['leases']:return None
                if s['starts']>=p.daily_starts:raise LifecycleError('daily start budget exhausted')
                if p.require_watchdog and not self.store.watchdog_fresh(now):raise LifecycleError('external watchdog not verified')
                s['starts']+=1;s['run_since']=now;s['managed']=True
            else:
                if s.get('provider_status') not in ('RUNNING','STOPPING'):return None
                if not s.get('managed') and not p.adopt_running:return None
                limit_due=bool(s.get('run_since') and now-s['run_since']>=p.max_uptime_seconds) or s['day_seconds']>=p.daily_runtime_seconds
                if s['leases'] or (not limit_due and now-s['last_activity']<p.idle_seconds):return None
                if not s['last_activity'] and not p.adopt_running:return None
            op={'kind':kind,'id':str(uuid.uuid4()),'until':now+30};s['op']=op;s['epoch']=s.get('epoch',0)+1;s['phase']='starting' if kind=='start' else 'stopping'
            return dict(op)
        return self.store.change(p.key,claim)

    def ensure_ready(self,p,reservation,cancel=None):
        started=self.clock();cancel=cancel or threading.Event()
        while True:
            if cancel.is_set():raise LifecycleError('request cancelled')
            self.heartbeat(p,reservation)
            view=self.store.view(p.key);lease=view['leases'].get(reservation)
            if not lease or self.clock()>=min(lease['deadline'],started+p.boot_seconds):raise LifecycleError('worker startup deadline exceeded')
            epoch=view.get('epoch',0)
            vm=self.provider.describe(p);view=self._observe(p,vm,epoch)
            if view.get('observation_stale'):
                self.sleep(1);continue
            if view.get('run_since') and view['run_since']+p.max_uptime_seconds<lease['deadline']+30:
                raise LifecycleError('request cannot fit before worker uptime limit')
            if vm['status']=='RUNNING' and not view.get('managed') and not p.adopt_running:
                raise LifecycleError('running worker has not been explicitly migrated to managed service')
            if view['op'] and view['op']['kind']=='stop':
                self.sleep(1);continue
            if vm['status']=='TERMINATED':
                op=self._claim(p,'start',view.get('epoch',0))
                if op:
                    self.store.event(p.key,'start_requested',op['id'],self.clock())
                    try:self.provider.start(p,op['id'])
                    except TimeoutError:pass # Persist intent; inspect before retry.
            elif vm['status']=='RUNNING' and self.provider.model_ready(p):
                self.store.event(p.key,'model_ready','verified model metadata',self.clock());return
            self.sleep(1)

    def reconcile(self,external=False):
        results=[]
        for p in self.policies.values():
            try:
                epoch=(self.store.view(p.key) or {}).get('epoch',0)
                vm=self.provider.describe(p);view=self._observe(p,vm,epoch)
                if view.get('observation_stale'):continue
                if vm['status'] not in ('RUNNING','STOPPING'):continue
                if view['op'] and view['op']['kind']=='stop':
                    op=self._claim(p,'stop',view.get('epoch',0))
                elif vm['status']=='RUNNING':op=self._claim(p,'stop',view.get('epoch',0))
                else:op=None
                if op:
                    self.provider.stop(p,op['id'])
                    self.store.event(p.key,'stop_requested',op['id'],self.clock())
                results.append({'worker':p.key,'status':vm['status'],'active':len(view['leases']),'stop_requested':bool(op)})
            except Exception:
                self.store.event(p.key,'reconcile_error','provider/guard unavailable; no destructive fallback',self.clock())
                results.append({'worker':p.key,'error':'reconciliation failed'})
        if external and not any('error' in r for r in results):self.store.watchdog_beat(self.clock())
        return results

    def snapshot(self,live=False):
        # Local persisted metadata only. A poll cannot refresh idle activity.
        out=[];now=self.clock()
        for p in self.policies.values():
            s=self.store.view(p.key) or {}
            leases=[x for x in s.get('leases',{}).values() if x['expires']>now and x['deadline']>now]
            item={'worker':p.key,'model':p.model,'phase':s.get('phase','unknown'),'provider_status':s.get('provider_status','unknown'),'active':len(leases),'idle_seconds':p.idle_seconds,'idle_remaining':None if leases or not s.get('last_activity') else max(0,int(p.idle_seconds-now+s['last_activity'])),'daily_starts':s.get('starts',0),'daily_runtime_seconds':int(s.get('day_seconds',0)),'observed_at':int(now),'live':False}
            if live:
                vm=self.provider.describe(p)
                item['provider_status']=vm['status'];item['model_ready']=vm['status']=='RUNNING' and self.provider.model_ready(p);item['live']=True
            out.append(item)
        return out
