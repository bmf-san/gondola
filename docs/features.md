# Features

## Reverse proxy & virtual hosts

Requests are routed to an upstream by matching the request `Host` header
against each upstream's `host_name`. Reverse proxies are constructed once at
startup for efficiency.

## Static files & fallback

`static_files` serve files under a URL path prefix. When a file is missing or a
directory has no `index.html`, the configured `fallback_path` is served.

## Custom error pages

`error_pages` maps status codes (a single code, a space-separated group, or
`default`) to custom pages, preserving the original status code. See
[Configuration](configuration.md#error_pages).

## TLS

Set `tls_cert_path` and `tls_key_path` to enable TLS. The minimum negotiated
version is TLS 1.2.

## Graceful shutdown & signals

| Signal | Behavior |
|---|---|
| `SIGINT` / `SIGTERM` | Graceful shutdown (waits up to `shutdown_timeout`) |
| `SIGHUP` | Reload the configuration (routing) without dropping connections |
| `SIGUSR1` | Reopen the access log file (for rotation) |

!!! note
    `SIGHUP` reload applies routing/upstream/static changes. Changes to the
    listening port or server-level timeouts require a restart.

## Timeouts

Both server-level (`read_header`, `read`, `write`, `idle`) and per-upstream
(`read`, `write`) timeouts are configurable. See [Configuration](configuration.md).
