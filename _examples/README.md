# _examples
This is a sample code for using gondola as a proxy server and a backend server to verify their operation.

There are two examples:
- **`standalone/`** — the quickest way to try gondola (`go run .`, no Docker).
- **Docker Compose** — a production-like multi-host setup with TLS.

# Quickstart (no Docker)
Starts two in-process backends and a gondola reverse proxy in front of them.
No Docker and no `/etc/hosts` changes required.

```sh
cd standalone
go run .
```

In another terminal:
```sh
# virtual-host routing by Host header
curl -H "Host: backend1.local" http://localhost:8080/
curl -H "Host: backend2.local" http://localhost:8080/

# static files + fallback
curl http://localhost:8080/public/index.html
curl http://localhost:8080/public/missing.html   # -> 404.html
```

# Docker Compose example (TLS, virtual hosts)
## Edit a /etc/hosts
```sh
sudo vim /etc/hosts
```

```sh
# gondola
127.0.0.1 backend1.local
127.0.0.1 backend2.local
```

## Start a gondola
```sh
make create-certs
make up
```

To stop the stack:
```sh
make down
```

# Demonstration
## Access to a backend server
`https://backend1.local` and `https://backend2.local` are available.

## Static Files and Fallback
You can test static file hosting and fallback functionality:

1. Normal static file access
```
https://backend1.local/public/index.html   # Successfully displays index.html
https://backend1.local/public/example.html # Successfully displays example.html
```

2. Fallback demonstration
```
https://backend1.local/public/not-exist.html  # Access to a non-existent file
```
When accessing a non-existent file, the configured fallback file is served (`fallback_path: 404.html`, relative to the static `dir`).
