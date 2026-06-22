# Getting Started

## Installation

### Go

```bash
go get github.com/bmf-san/gondola
```

### Binary

Download the latest binary from the [Releases](https://github.com/bmf-san/gondola/releases) page.

### Docker

```bash
docker pull bmfsan/gondola
docker run -v $(pwd)/config.yaml:/etc/gondola/config.yaml -p 8080:8080 bmfsan/gondola
```

## Quickstart (no Docker)

The repository ships a self-contained example that starts two in-process
backends behind a gondola proxy:

```bash
cd _examples/standalone
go run .
```

Then, in another terminal:

```bash
# virtual-host routing by Host header
curl -H "Host: backend1.local" http://localhost:8080/
curl -H "Host: backend2.local" http://localhost:8080/

# static files + fallback
curl http://localhost:8080/public/index.html
```

## Running with a config file

```bash
gondola -config config.yaml
```

See [Configuration](configuration.md) for the full file reference.
