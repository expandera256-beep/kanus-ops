import json, threading, unittest
from urllib.request import Request, urlopen
from urllib.error import HTTPError
from http.server import ThreadingHTTPServer
from server import Handler, TASKS
class ServiceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        TASKS.clear()
        cls.srv = ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        cls.thread = threading.Thread(target=cls.srv.serve_forever, daemon=True)
        cls.thread.start()
        cls.base = 'http://127.0.0.1:' + str(cls.srv.server_port)
    @classmethod
    def tearDownClass(cls):
        cls.srv.shutdown(); cls.srv.server_close(); cls.thread.join(timeout=2)
    def call(self, path, body=None, user=None):
        headers = {'Content-Type': 'application/json'}
        if user: headers['Authorization'] = 'Bearer ' + user
        req = Request(self.base + path, data=json.dumps(body).encode() if body is not None else None, headers=headers)
        try:
            with urlopen(req, timeout=3) as r: return r.status, json.load(r)
        except HTTPError as e: return e.code, json.load(e)
    def test_health(self): self.assertEqual(self.call('/healthz'), (200, {'ok': True}))
    def test_auth(self): self.assertEqual(self.call('/tasks')[0], 401)
    def test_cross_user_isolation(self):
        self.assertEqual(self.call('/tasks', {'title': 'alice-private'}, 'alice')[0], 201)
        self.assertEqual(self.call('/tasks', user='bob'), (200, {'tasks': []}))
        self.assertEqual(self.call('/tasks', user='alice')[1]['tasks'][0]['title'], 'alice-private')
    def test_validation(self): self.assertEqual(self.call('/tasks', {'title': ''}, 'bob')[0], 400)
if __name__ == '__main__': unittest.main()
