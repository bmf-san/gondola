# Gondola

A simple and flexible reverse proxy written in Go

## Features
- Lightweight and fast execution
- Simple configuration file (YAML)
- Fallback support
- Static file hosting
- Virtual host support
- Graceful shutdown
- Detailed access logs (nginx-compatible)
- Structured access logs (JSON)
- TLS/SSL support
- Trace ID support
- Timeout control

## Installation

### Go
```bash
go get github.com/bmf-san/gondola
```

### Binary
Download the latest binary from the [Releases](https://github.com/bmf-san/gondola/releases) page.

### Docker
```bash
docker pull bmf-san/gondola
docker run -v $(pwd)/config.yaml:/etc/gondola/config.yaml bmf-san/gondola
```

## Usage

### Command Line Options
```bash
Usage: gondola [options]

Options:
  -config string    Path to configuration file (default: /etc/gondola/config.yaml)
  -version         Display version information
  -help           Show help
```

### Environment Variables
The following environment variables can override command line options:

- `GONDOLA_CONFIG`: Path to the configuration file
- `GONDOLA_LOG_LEVEL`: Log level (debug, info, warn, error)

### Signal Handling
Gondola handles the following signals:

- `SIGTERM`, `SIGINT`: Graceful shutdown
- `SIGHUP`: Reload configuration file
- `SIGUSR1`: Reopen access logs

### Configuration File
```yaml
proxy:
  port: "8080"
  read_header_timeout: 2000  # milliseconds
  shutdown_timeout: 3000     # milliseconds
  log_level: "info"         # debug, info, warn, error
  static_files:
    - path: /public/
      dir: /path/to/public

upstreams:
  - host_name: api.example.com
    target: http://localhost:3000
    read_timeout: 5000      # milliseconds
    write_timeout: 5000     # milliseconds
  - host_name: web.example.com
    target: http://localhost:8000
```

### Startup Examples

Basic startup:
```bash
gondola -config config.yaml
```

Start in debug mode:
```bash
GONDOLA_LOG_LEVEL=debug gondola -config config.yaml
```

Start with Docker:
```bash
docker run -v $(pwd)/config.yaml:/etc/gondola/config.yaml \
          -p 8080:8080 \
          bmf-san/gondola
```

Start with systemd:
```ini
[Unit]
Description=Gondola Reverse Proxy
After=network.target

[Service]
ExecStart=/usr/local/bin/gondola -config /etc/gondola/config.yaml
Restart=always
Environment=GONDOLA_LOG_LEVEL=info

[Install]
WantedBy=multi-user.target
```

## Access Logs

Gondola outputs detailed access logs compatible with nginx.

### Log Fields

1. Client Information
- `remote_addr`: Client IP address
- `remote_port`: Client port number
- `x_forwarded_for`: X-Forwarded-For header

2. Request Information
- `method`: HTTP method (GET, POST, etc.)
- `request_uri`: Full request URI
- `query_string`: Query parameters
- `host`: Host header
- `request_size`: Request size

3. Response Information
- `status`: HTTP status
- `body_bytes_sent`: Response body size
- `bytes_sent`: Total response size
- `request_time`: Request processing time (seconds)

4. Upstream Information
- `upstream_addr`: Backend address
- `upstream_status`: Backend status
- `upstream_size`: Backend response size
- `upstream_response_time`: Backend response time (seconds)

5. Other Headers
- `referer`: Referer header
- `user_agent`: User-Agent header

### Log Output Example
```json
{
  "level": "INFO",
  "msg": "access_log",
  "remote_addr": "192.168.1.100",
  "remote_port": "54321",
  "x_forwarded_for": "",
  "method": "GET",
  "request_uri": "/api/users",
  "query_string": "page=1",
  "host": "api.example.com",
  "request_size": 243,
  "status": "OK",
  "body_bytes_sent": 1532,
  "bytes_sent": 1843,
  "request_time": 0.153,
  "upstream_addr": "localhost:3000",
  "upstream_status": "200 OK",
  "upstream_size": 1532,
  "upstream_response_time": 0.142,
  "referer": "https://example.com",
  "user_agent": "Mozilla/5.0 ...",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

# Projects
- [The gondola's board](https://github.com/users/bmf-san/projects/1/views/1)

# ADR
- [ADR](https://github.com/bmf-san/gondola/discussions?discussions_q=is%3Aopen+label%3AADR)

# Contribution
We welcome issues and pull requests at any time.

Feel free to contribute!

Before contributing, please check the following documents:

- [CODE_OF_CONDUCT](https://github.com/bmf-san/godra/blob/main/.github/CODE_OF_CONDUCT.md)
- [CONTRIBUTING](https://github.com/bmf-san/godra/blob/main/.github/CONTRIBUTING.md)

# Sponsors
If you like this project, consider sponsoring us!

[Github Sponsors - bmf-san](https://github.com/sponsors/bmf-san)

Alternatively, giving us a star would be appreciated!

It helps motivate us to continue maintaining this project. :D

# License
This project is licensed under the MIT License.

[LICENSE](https://github.com/bmf-san/gondola/blob/main/LICENSE)

# Stargazers
[![Stargazers repo roster for @bmf-san/gondola](https://reporoster.com/stars/bmf-san/gondola)](https://github.com/bmf-san/gondola/stargazers)

# Forkers
[![Forkers repo roster for @bmf-san/gondola](https://reporoster.com/forks/bmf-san/gondola)](https://github.com/bmf-san/gondola/network/members)

# Author
[bmf-san](https://github.com/bmf-san)

- Email
  - bmf.infomation@gmail.com
- Blog
  - [bmf-tech.com](http://bmf-tech.com)
- Twitter
  - [bmf-san](https://twitter.com/bmf-san)

