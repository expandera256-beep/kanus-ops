import unittest
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
