# TODOs

## Multi-instance realtime monitor aggregation

**What:** Add multi-instance support for the realtime monitor after v1 ships.

**Why:** The v1 monitor uses an in-process ring buffer. That is correct for the trimmed first version, but in a multi-instance deployment each API process has its own buffer and its own log IDs. A load balancer can make the browser poll different instances, causing missing logs, duplicate-looking IDs, or misleading ordering.

**Context:** v1 should explicitly label itself as single-instance/current-process monitoring. When the backend is deployed behind multiple API instances, add one of: sticky routing plus instance display, an `instance_id` selector, Redis/pubsub fan-in, or a real centralized log backend. Start from `server/internal/monitoring/` and the `/api/admin/monitoring/logs/recent` endpoint.

**Depends on / blocked by:** Ship the single-instance v1 first. Revisit when deployment topology includes more than one API process or when operators need cross-instance incident visibility.
