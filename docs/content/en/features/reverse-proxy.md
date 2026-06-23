---
title: "Reverse Proxy"
description: "Virtual-host routing to upstream servers."
slug: "reverse-proxy"
categories:
  - features
---

Requests are routed to an upstream by matching the request `Host` header
against each upstream's `host_name`. Reverse proxies are constructed once at
startup for efficiency, and keep-alive connections to upstreams are reused via
Go's standard transport.

```yaml
upstreams:
  - host_name: api.example.com
    target: http://localhost:3000
    read_timeout: 5000   # upstream response header timeout (ms)
    write_timeout: 5000  # upstream connection timeout (ms)
```

A request whose Host matches no upstream and whose path matches no static rule
returns `404`.
