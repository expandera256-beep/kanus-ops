"""KAN.US thin HTTP/SSE client aligned with api-0009 OpenAPI (/v1)."""

from .client import KanusClient, KanusError, ErrorEnvelope, SSEEvent
from .auth import Auth

__all__ = ["KanusClient", "KanusError", "ErrorEnvelope", "SSEEvent", "Auth"]
__version__ = "0.1.0"
