# _examples
A self-contained demo of gondola: two in-process backends behind a gondola
reverse proxy, runnable with a single `go run .` — no Docker, no `/etc/hosts`
changes required.

## Run
```sh
cd standalone
go run .
```

## Try it
In another terminal:
```sh
# virtual-host routing by Host header
curl -H "Host: backend1.local" http://localhost:8080/
curl -H "Host: backend2.local" http://localhost:8080/

# static files + fallback
curl http://localhost:8080/public/index.html
curl http://localhost:8080/public/missing.html   # -> 404.html
```
