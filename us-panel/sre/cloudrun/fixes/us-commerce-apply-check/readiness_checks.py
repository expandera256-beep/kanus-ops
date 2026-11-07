"""Read-only readiness: never contacts a GPU/provider or touches activity."""
def readiness(store, configured=True):
    checks = {'config': bool(configured), 'store': False}
    try:
        factory = getattr(store, '_db', None) or getattr(store, 'db', None)
        if not callable(factory):
            raise RuntimeError('store unavailable')
        with factory() as db:
            db.execute('SELECT 1')
        checks['store'] = True
    except Exception:
        pass  # Never expose driver errors or a DSN to HTTP/log consumers.
    return all(checks.values()), checks
