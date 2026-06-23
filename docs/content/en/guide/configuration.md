---
title: "Configuration"
description: "The gondola YAML configuration reference."
slug: "configuration"
categories:
  - guide
---

Gondola is configured with a single YAML file. Environment variables of the
form `$VAR` / `${VAR}` are expanded before parsing.

## Example

```yaml
proxy:
  port: "8080"
  read_header_timeout: 2000  # milliseconds
  read_timeout: 30000
  write_timeout: 30000
  idle_timeout: 60000
  shutdown_timeout: 3000
  max_header_bytes: 1048576  # bytes
  # tls_cert_path: /path/to/cert.pem
  # tls_key_path: /path/to/key.pem
  # log_file: /var/log/gondola/access.log  # default: stdout
  static_files:
    - path: /public/
      dir: /path/to/public
      fallback_path: 404.html
  # error_pages:
  #   404: /404.html
  #   "500 502 503 504": /50x.html
  #   default: /error.html

upstreams:
  - host_name: api.example.com
    target: http://localhost:3000
    read_timeout: 5000
    write_timeout: 5000
  - host_name: web.example.com
    target: http://localhost:8000

log_level: "info"            # debug, info, warn, error
log_format: "json"           # json, common, combined, custom
# log_custom_format: '${remote_ip} ${method} ${uri} ${status}'
```

## proxy

| Key | Type | Description |
|---|---|---|
| `port` | string | Listen port |
| `read_header_timeout` | int (ms) | Header read timeout |
| `read_timeout` | int (ms) | Request read timeout |
| `write_timeout` | int (ms) | Response write timeout |
| `idle_timeout` | int (ms) | Keep-alive idle timeout |
| `shutdown_timeout` | int (ms) | Graceful shutdown timeout |
| `max_header_bytes` | int | Maximum request header size |
| `tls_cert_path` / `tls_key_path` | string | Enable TLS when both are set |
| `log_file` | string | Log file path (default: stdout) |
| `static_files` | list | Static file rules |
| `error_pages` | map | Custom error pages by status code |

## upstreams

| Key | Type | Description |
|---|---|---|
| `host_name` | string | Matched against the request Host header |
| `target` | string | Upstream URL to forward to |
| `read_timeout` | int (ms) | Upstream response header timeout |
| `write_timeout` | int (ms) | Upstream connection timeout |

## Logging

| Key | Type | Description |
|---|---|---|
| `log_level` | string | debug, info, warn, error |
| `log_format` | string | json, common, combined, custom |
| `log_custom_format` | string | Template used when `log_format` is `custom` |

## Environment variables

- `GONDOLA_CONFIG` — config file path (overrides `-config`)
- `GONDOLA_LOG_LEVEL` — log level override
