import json
import tempfile
import unittest
import urllib.error
from pathlib import Path

from account_store import AccountStore
from delivery import DeliveryUncertain, deliver_one, send_hostinger


class Response:
    status = 204
    def __enter__(self): return self
    def __exit__(self, *args): pass


class Opener:
    def __init__(self, error=None): self.error, self.request = error, None
    def open(self, request, timeout):
        self.request = request
        if self.error: raise self.error
        return Response()


class HostingerDeliveryTest(unittest.TestCase):
    def test_send_request_contract(self):
        opener = Opener()
        send_hostinger('example-token', 'AC123', 'buyer@example.com', ['test-license'], opener)
        self.assertEqual(opener.request.full_url, 'https://api.mail.hostinger.com/api/v1/mailboxes/AC123/send')
        self.assertEqual(opener.request.get_header('Authorization'), 'Bearer example-token')
        self.assertEqual(json.loads(opener.request.data)['to'], ['buyer@example.com'])
        self.assertIn('test-license', json.loads(opener.request.data)['text'])

    def test_timeout_marks_review_instead_of_retry(self):
        with tempfile.TemporaryDirectory() as tmp:
            store = AccountStore(Path(tmp)/'db')
            with store._db() as db:
                db.execute("INSERT INTO delivery_jobs(order_id,email) VALUES('1','buyer@example.com')")
                db.execute("INSERT INTO issued_licenses VALUES('hash','1','test-license','buyer@example.com','active')")
            with self.assertRaises(DeliveryUncertain):
                deliver_one(store, lambda recipient, keys: send_hostinger('token', 'AC123', recipient, keys, Opener(TimeoutError())))
            with store._db() as db:
                self.assertEqual(db.execute('SELECT state FROM delivery_jobs').fetchone()[0], 'review')
            self.assertFalse(deliver_one(store, lambda *_: self.fail('ambiguous send was retried')))

    def test_reject_invalid_mailbox_identifier(self):
        with self.assertRaises(ValueError):
            send_hostinger('token', '../bad', 'a@example.com', ['license'], Opener())


if __name__ == '__main__': unittest.main()
