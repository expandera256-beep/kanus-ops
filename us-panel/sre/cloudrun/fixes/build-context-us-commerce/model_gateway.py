"""Synchronous in-process adapter. HTTP streaming uses commerce_http instead."""
import hashlib
import json

class MeteredGateway:
    def __init__(self, store, upstream):
        self.store = store
        self.upstream = upstream

    def complete(self, session, request_id, payload, now=None):
        fingerprint=hashlib.sha256(json.dumps(payload,sort_keys=True,separators=(',',':'),ensure_ascii=False).encode()).hexdigest()
        reservation,remaining = self.store.reserve_prompt(session,request_id,fingerprint,now=now)
        try:
            response = self.upstream(payload)
            if response is None or response == '':raise ValueError('empty upstream response')
        except BaseException:
            self.store.finish_reservation(reservation,False,now=now)
            raise
        self.store.finish_reservation(reservation,True,now=now)
        return {'response': response, 'remaining': remaining}
