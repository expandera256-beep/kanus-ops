from pathlib import PurePosixPath
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
        if PurePosixPath(name).name!=name or name in ('.','..') or '\\' in name:raise ValueError('unsafe path')
        if len(content)>1024:raise ValueError('too large')
        self.files[(n,name)]=bytes(content)
    def download(self,user,n,name):self.project(user,n);return self.files[(n,name)]
    def admin(self,user):
        tenant,role=self.who(user)
        if role!='admin':raise PermissionError('403')
        return [p for p in self.projects.values() if p['tenant']==tenant]
