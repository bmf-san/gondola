---
title: "Static Files & Error Pages"
description: "Serve static assets with fallback and custom error pages."
slug: "static-files"
categories:
  - features
---

## Static files

`static_files` serve files under a URL path prefix. When a file is missing or a
directory has no `index.html`, the configured `fallback_path` (relative to
`dir`) is served. Paths are validated to stay within the configured directory.

```yaml
proxy:
  static_files:
    - path: /public/
      dir: /var/www/public
      fallback_path: 404.html
```

## Custom error pages

`error_pages` maps status codes — a single code, a space-separated group, or
`default` — to custom pages (relative to the first `static_files` directory).
The original status code is preserved, and gondola falls back to the built-in
response when a page file is missing.

```yaml
proxy:
  error_pages:
    404: /404.html
    "500 502 503 504": /50x.html
    default: /error.html
```
