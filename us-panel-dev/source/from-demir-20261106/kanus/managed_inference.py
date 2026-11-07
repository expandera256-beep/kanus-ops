"""Bounded request/worker lease around a lazy model connection and response.

The SSE caller receives comments during cold start. They do not count as model
output and do not complete metering. Disconnect closes the upstream and releases
the lease, including before the first WSGI iteration.
"""
import json
import queue
import threading
import time
from inference_lifecycle import LifecycleError


class ManagedResponse:
    status=200

    def __init__(self,controller,reservation,model,body,upstream,stream):
        self.controller=controller;self.reservation=reservation;self.policy,self.lease=controller.acquire(reservation,model)
        self.headers={'Content-Type':'text/event-stream' if stream else 'application/json'}
        self.stream=stream;self.body=body;self.upstream=upstream
        self.queue=queue.Queue(maxsize=8);self.stop=threading.Event();self.response=None
        self.closed=False;self.lock=threading.Lock();self.release_lock=threading.Lock();self.released=False;self.buffer=b'';self.ended=False
        self.worker=threading.Thread(target=self._run,daemon=True);self.worker.start()

    def _put(self,item):
        while not self.stop.is_set():
            try:self.queue.put(item,timeout=.2);return
            except queue.Full:pass

    def _run(self):
        try:
            self.controller.ensure_ready(self.policy,self.reservation,self.stop)
            if self.stop.is_set():return
            response=self.upstream(self.body)
            with self.lock:
                if self.closed:response.close();return
                self.response=response
            if getattr(response,'status',200)!=200:raise LifecycleError('model rejected request')
            expected='text/event-stream' if self.stream else 'application/json'
            if expected not in response.headers.get('Content-Type','application/json').lower():raise LifecycleError('model content type mismatch')
            while not self.stop.is_set():
                self.controller.heartbeat(self.policy,self.reservation)
                reader=getattr(response,'read1',None) or response.read
                block=reader(8192)
                if not block:break
                self._put(block)
            self._put(None)
        except Exception:
            self._put(LifecycleError('worker startup/inference failed; credit will be refunded'))
        finally:
            self._release_once()

    def _release_once(self):
        with self.release_lock:
            if self.released:return
            try:self.controller.release(self.policy,self.reservation)
            except Exception:
                # The durable, bounded lease still expires if the state store
                # is unavailable. Never emit provider/DB secrets on a thread.
                return
            self.released=True

    def read1(self,n=8192):
        if self.buffer:
            block,self.buffer=self.buffer[:n],self.buffer[n:];return block
        if self.ended:return b''
        while True:
            if self.stop.is_set():raise LifecycleError('request cancelled')
            if self.controller.clock()>=self.lease['deadline']:
                self.close();raise LifecycleError('absolute request deadline exceeded')
            try:item=self.queue.get(timeout=min(2,max(.01,self.lease['deadline']-self.controller.clock())))
            except queue.Empty:
                self.controller.heartbeat(self.policy,self.reservation)
                if self.stream:return b': us worker preparing or processing\n\n'
                continue
            if isinstance(item,BaseException):self.close();raise item
            if item is None:self.ended=True;return b''
            self.buffer=item[n:];return item[:n]

    read=read1

    def close(self):
        with self.lock:
            if self.closed:return
            self.closed=True;self.stop.set();response=self.response
        # Release first: close() on a network reader can block until its socket
        # timeout. Never keep a customer lease alive because of a blocked close.
        self._release_once()
        if response is not None:
            threading.Thread(target=response.close,daemon=True).start()


class ManagedModelClient:
    def __init__(self,controller,clients):self.controller=controller;self.clients=clients

    def metered(self,body,reservation):
        obj=json.loads(body);policy=self.controller.policy_for_model(obj['model'])
        return ManagedResponse(self.controller,reservation,obj['model'],body,self.clients[policy.key],obj.get('stream',False))

    def __call__(self,body):raise LifecycleError('unreserved model call rejected')

    def models(self):
        # Probe only already-running workers. Neither metadata nor health reads
        # acquire a lease, update idle activity or call provider.start.
        result=[]
        for p in self.controller.policies.values():
            vm=self.controller.provider.describe(p)
            ready=vm['status']=='RUNNING' and self.controller.provider.model_ready(p)
            item={'id':p.model,'status':{'value':'loaded' if ready else 'unloaded'},'worker_state':vm['status'].lower()}
            if ready:
                fetch=getattr(self.clients[p.key],'models',None)
                if callable(fetch):
                    response=fetch()
                    try:
                        raw=response.read(1_000_001)
                        if len(raw)>1_000_000:raise LifecycleError('model metadata exceeds bound')
                        listing=json.loads(raw)
                    finally:response.close()
                    for entry in listing.get('data',[]):
                        if entry.get('id')==p.model:item.update(entry);break
                    item['status']={'value':'loaded'}
            result.append(item)
        import io
        response=io.BytesIO(json.dumps({'data':result}).encode());response.status=200;return response
