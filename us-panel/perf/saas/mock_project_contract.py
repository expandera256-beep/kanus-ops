"""Fixed mock service contract, NOT model-produced SaaS/PostgreSQL acceptance."""
import copy,json,os,subprocess,sys,time
from pathlib import Path
OUT=Path('/workspace/us-panel/reports/r20261108')
FILES={
 'backend.py':"""from pathlib import PurePosixPath
class Service:
    def __init__(self):
        self.users={'alice':('A','member'),'bob':('B','member'),'admin':('A','admin')}
        self.projects={};self.tasks={};self.files={}
    def who(self,user):
        if user not in self.users: raise PermissionError('401')
        return self.users[user]
    def create(self,user,name):
        tenant,role=self.who(user);n=len(self.projects)+1
        self.projects[n]={'id':n,'tenant':tenant,'name':name};return n
    def project(self,user,n):
        tenant,role=self.who(user)
        if n not in self.projects or self.projects[n]['tenant']!=tenant:raise PermissionError('403')
        return self.projects[n]
    def task(self,user,n,title):
        self.project(user,n);i=len(self.tasks)+1;self.tasks[i]={'project':n,'title':title};return i
    def upload(self,user,n,name,content):
        self.project(user,n)
        if PurePosixPath(name).name!=name or name in ('.','..') or '\\\\' in name:raise ValueError('unsafe path')
        if len(content)>1024:raise ValueError('too large')
        self.files[(n,name)]=bytes(content)
    def download(self,user,n,name):self.project(user,n);return self.files[(n,name)]
    def admin(self,user):
        tenant,role=self.who(user)
        if role!='admin':raise PermissionError('403')
        return [p for p in self.projects.values() if p['tenant']==tenant]
""",
 'test_backend.py':"""import unittest
from backend import Service
class Contracts(unittest.TestCase):
    def setUp(self):self.s=Service();self.p=self.s.create('alice','Project A')
    def test_auth(self):
        with self.assertRaises(PermissionError):self.s.create('unknown','X')
    def test_project_task(self):self.assertEqual(self.s.task('alice',self.p,'T'),1)
    def test_cross_tenant(self):
        with self.assertRaises(PermissionError):self.s.project('bob',self.p)
    def test_admin(self):
        self.assertEqual(len(self.s.admin('admin')),1)
        with self.assertRaises(PermissionError):self.s.admin('alice')
    def test_upload_download(self):
        self.s.upload('alice',self.p,'a.txt',b'hello');self.assertEqual(self.s.download('alice',self.p,'a.txt'),b'hello')
    def test_upload_traversal(self):
        with self.assertRaises(ValueError):self.s.upload('alice',self.p,'../a',b'x')
    def test_upload_limit(self):
        with self.assertRaises(ValueError):self.s.upload('alice',self.p,'a',b'x'*1025)
    def test_upload_cross_tenant(self):
        self.s.upload('alice',self.p,'a',b'A')
        with self.assertRaises(PermissionError):self.s.download('bob',self.p,'a')
if __name__=='__main__':unittest.main()
""",
 'frontend/index.html':'<!doctype html><title>KANUS SaaS Project Manager — MOCK</title><h1>Mock contract fixture, not a complete SaaS</h1>\n',
 'README.md':'MOCK fixed fixture. python3 -m unittest discover -v. No actual login/session security, PostgreSQL, persistence, Docker, browser UI or model repair cycle verified.\n'}
def main():
    p=OUT/'saas_mock_fixture';p.mkdir(exist_ok=True)
    for name,text in FILES.items():q=p/name;q.parent.mkdir(parents=True,exist_ok=True);q.write_text(text)
    env={'PATH':os.environ.get('PATH',''),'HOME':str(p),'LANG':'C.UTF-8','PYTHONNOUSERSITE':'1'};steps=[]
    for cmd in ([sys.executable,'-m','compileall','-q','.'],[sys.executable,'-m','unittest','discover','-v']):
        r=subprocess.run(cmd,cwd=p,env=env,capture_output=True,text=True,timeout=20)
        steps.append({'command':cmd[1:],'exit_code':r.returncode,'stdout':r.stdout,'stderr':r.stderr})
    result={'data_source':'MOCK_FIXED_CONTRACT_NOT_MODEL_OUTPUT','pass':all(r['exit_code']==0 for r in steps),'steps':steps,'fixture_file_manifest':list(FILES),'mock_scope':'auth lookup/RBAC/project/task/tenant/upload contracts only; in-memory service, not real backend HTTP/UI/DB','real_model_acceptance':'BLOCKED: .11 deployment + GPU window pending','not_tested':['PostgreSQL','real frontend/browser','actual backend HTTP integration','Docker','model-generated code','agent repair cycle']}
    (OUT/'results/saas7_results.json').write_text(json.dumps(result,indent=2))
    (OUT/'results/saas7_results.md').write_text('# MOCK — §7 contract fixture\n\n'+json.dumps(result,indent=2)+'\n')
    print(json.dumps({'mock_pass':result['pass'],'real_model_acceptance':result['real_model_acceptance']}))
if __name__=='__main__':main()
