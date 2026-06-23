---
title: "TLS"
description: "Serve over HTTPS with TLS 1.2+."
slug: "tls"
categories:
  - features
---

Set `tls_cert_path` and `tls_key_path` to enable TLS. The minimum negotiated
version is TLS 1.2.

```yaml
proxy:
  port: "443"
  tls_cert_path: /etc/gondola/cert.pem
  tls_key_path: /etc/gondola/key.pem
```

When both paths are set, gondola serves HTTPS; otherwise it serves plain HTTP.
