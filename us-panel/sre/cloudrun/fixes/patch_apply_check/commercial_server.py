"""Local-only WSGI runner for integration testing, not a public production server."""
import os
import json
import threading
from pathlib import Path
from socketserver import ThreadingMixIn
from wsgiref.simple_server import make_server, WSGIServer, WSGIRequestHandler
from account_store import AccountStore
from commerce_http import local_model_client, make_app


class BoundedWSGIServer(ThreadingMixIn, WSGIServer):
    """Local HTTP concurrency; the shared database caps model work at ten."""
    daemon_threads = True
    block_on_close = False

    def __init__(self, *args, **kwargs):
        self._slots = threading.BoundedSemaphore(32)
        super().__init__(*args, **kwargs)

    def process_request(self, request, address):
        if not self._slots.acquire(blocking=False):
            try:
                request.sendall(b'HTTP/1.0 503 Service Unavailable\r\nContent-Length: 0\r\nConnection: close\r\n\r\n')
            except OSError:
                pass
            finally:
                self.shutdown_request(request)
            return
        try:
            super().process_request(request, address)
        except BaseException:
            self._slots.release()
            raise

    def process_request_thread(self, request, address):
        try:
            super().process_request_thread(request, address)
        finally:
            self._slots.release()


class LocalRequestHandler(WSGIRequestHandler):
    timeout = 15

    def log_message(self, format, *args):
        # Request URLs, headers and bodies do not belong in account logs.
        return


def require(name):
    value=os.environ.get(name,'').strip()
    if not value:raise RuntimeError(name+' is required')
    return value


def main():
    database=require('US_COMMERCE_DB')
    postgres=database.startswith(('postgresql://','postgres://'))
    db=None if postgres else Path(database)
    if db is not None and not db.is_absolute():raise RuntimeError('US_COMMERCE_DB must be absolute')
    os.umask(0o077)
    store=AccountStore(database)
    if db is not None and os.name!='nt':os.chmod(db,0o600)
    # Explicit allowlist of verified Shopify variant IDs and package prices.
    products=json.loads(os.environ.get('US_CREDIT_PRODUCTS','{}'))
    production=os.environ.get('US_ENV')=='production'
    if production and not os.environ.get('US_WORKERS_CONFIG'):raise RuntimeError('production requires request-driven worker policy')
    if os.environ.get('US_WORKERS_CONFIG'):
        from lifecycle_config import configure_managed
        upstream,controller=configure_managed(store,require('US_WORKERS_CONFIG'),production)
    else:upstream=local_model_client(require('US_MODEL_UPSTREAM'))
    app=make_app(store,require('SHOPIFY_WEBHOOK_SECRET').encode(),require('US_LICENSE_SECRET').encode(),require('SHOPIFY_SHOP'),upstream,products)
    port=int(os.environ.get('US_COMMERCE_PORT','8888'))
    with make_server('127.0.0.1',port,app,server_class=BoundedWSGIServer,handler_class=LocalRequestHandler) as server:server.serve_forever()


if __name__=='__main__':main()
