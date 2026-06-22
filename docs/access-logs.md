# Access Logs

Gondola emits detailed access logs compatible with nginx.

## Formats

Selectable via `log_format`:

- `json` (default) — structured JSON
- `common` — Apache Common Log Format
- `combined` — Apache Combined Log Format
- `custom` — a template via `log_custom_format`

### Custom variables

`${timestamp}`, `${remote_ip}`, `${method}`, `${uri}`, `${protocol}`,
`${status}`, `${user_agent}`, `${referer}`, `${request_time}`, `${trace_id}`

```yaml
log_format: "custom"
log_custom_format: '${remote_ip} ${method} ${uri} ${status} ${request_time}'
```

## JSON fields

Client, request, response, and upstream details are logged, including
`remote_addr`, `method`, `request_uri`, `status`, `status_code`,
`body_bytes_sent`, `request_time`, `upstream_addr`, `upstream_status`,
`upstream_response_time`, and `trace_id`.

## Output & rotation

Logs are written to stdout by default, or to `log_file` when set. Send
`SIGUSR1` to reopen the file after rotation (e.g. with logrotate).
