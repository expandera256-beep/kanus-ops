"""Validate completion framing without retaining prompt or answer bodies."""
import json

MAX_RESPONSE = 8 * 1024 * 1024


class ClosingBody:
    """WSGI close must settle even when iteration never starts."""
    def __init__(self, iterator, cleanup):
        self.iterator=iterator
        self.cleanup=cleanup
        self.closed=False

    def __iter__(self): return self

    def __next__(self):
        try:return next(self.iterator)
        except BaseException:
            self.close()
            raise

    def close(self):
        if self.closed:return
        self.closed=True
        try:self.iterator.close()
        finally:self.cleanup()


class CompletionReceipt:
    def __init__(self, content_type):
        self.sse = 'text/event-stream' in content_type.lower()
        self.buffer = b''
        self.total = 0
        self.content = False
        self.finished = False
        self.done = False
        self.invalid = False

    def _event(self, data):
        if data == b'[DONE]':
            self.done = True
            return
        if self.done:
            self.invalid = True
            return
        try:
            value=json.loads(data)
            if not isinstance(value,dict) or 'error' in value:
                raise ValueError('upstream error')
            for choice in value.get('choices',[]):
                message=choice.get('delta' if self.sse else 'message',{})
                content=message.get('content')
                if isinstance(content,str) and content:
                    self.content=True
                reason=choice.get('finish_reason')
                if reason=='stop': self.finished=True
                elif reason is not None: self.invalid=True
        except (ValueError,TypeError,AttributeError):
            self.invalid=True

    def feed(self, block):
        self.total+=len(block)
        if self.total>MAX_RESPONSE: raise ValueError('upstream response too large')
        self.buffer+=block
        if not self.sse: return
        while b'\n' in self.buffer:
            line,self.buffer=self.buffer.split(b'\n',1)
            line=line.rstrip(b'\r')
            if line.startswith(b'data:'):self._event(line[5:].strip())
            elif line and not line.startswith(b':') and not line.startswith((b'event:',b'id:',b'retry:')):self.invalid=True
        if len(self.buffer)>1024*1024:raise ValueError('upstream event too large')

    def complete(self):
        if not self.sse:
            self._event(self.buffer)
            self.buffer=b''
            self.done=True
        elif self.buffer.strip():
            # A disconnected stream must not be mistaken for a terminal event.
            self.invalid=True
        return self.done and self.finished and self.content and not self.invalid
