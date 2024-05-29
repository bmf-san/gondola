[English](https://github.com/bmf-san/gondola) [日本語](https://github.com/bmf-san/gondola/blob/master/README-ja.md)

# gondola
[![GitHub release](https://img.shields.io/github/release/bmf-san/gondola.svg)](https://github.com/bmf-san/gondola/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/bmf-san/gondola)](https://goreportcard.com/report/github.com/bmf-san/gondola)
[![codecov](https://codecov.io/gh/bmf-san/gondola/branch/main/graph/badge.svg?token=ZLOLQKUD39)](https://codecov.io/gh/bmf-san/gondola)
[![GitHub license](https://img.shields.io/github/license/bmf-san/gondola)](https://github.com/bmf-san/gondola/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmf-san/gondola.svg)](https://pkg.go.dev/github.com/bmf-san/gondola)
[![Sourcegraph](https://sourcegraph.com/github.com/bmf-san/gondola/-/badge.svg)](https://sourcegraph.com/github.com/bmf-san/gondola?badge)

A golang reverse proxy.

<img src="https://storage.googleapis.com/gopherizeme.appspot.com/gophers/22fd9b7a49eac4101fc9819578641c2e71706f6f.png" alt="gondola" title="gondola" width="250px">

This log was created by [gopherize.me](https://gopherize.me/gopher/22fd9b7a49eac4101fc9819578641c2e71706f6f)

# Table of contents
- [gondola](#gondola)
- [Table of contents](#table-of-contents)
- [Features](#features)
- [Install](#install)
  - [Go](#go)
  - [Binary](#binary)
  - [Docker](#docker)
- [Example](#example)
- [Usage](#usage)
- [Projects](#projects)
- [ADR](#adr)
- [Wiki](#wiki)
- [Contribution](#contribution)
- [Sponsor](#sponsor)
- [License](#license)
- [Stargazers](#stargazers)
- [Forkers](#forkers)
- [Author](#author)

# Features
- Virtual host
  - You can set up multiple hosts on upstream servers.
- Configuration file loader
  - You can use configuration files in YAML format.
- TLS
  - You can use TLS by preparing a TLS certificate.
- Serve static files
  - You can serve static files.
- Access log
  - Outputs Proxy access logs and Upstream servers access logs.
- Binary distribution
  - Distributing cross-compiled binaries.

# Install
## Go
```
go get -u github.com/bmf-san/gondola
```

## Binary
You can download the binary from the [release page](https://github.com/bmf-san/gondola/releases), and you can use it.

## Docker
<!-- TODO: -->

# Example
See below for how to use gondola.

- [_examples](https://github.com/bmf-san/gondola/tree/main/_examples)

# Usage
Run a binary with the option.

```sh
gondola -config config.yaml
```

# Projects
- [The gondola's board](https://github.com/users/bmf-san/projects/1/views/1)

# ADR
- [ADR](https://github.com/bmf-san/gondola/discussions?discussions_q=is%3Aopen+label%3AADR)

# Wiki
- [wiki](https://github.com/bmf-san/gondola/wiki)

# Contribution
Issues and Pull Requests are always welcome.

We would be happy to receive your contributions.

Please review the following documents before making a contribution.

- [CODE_OF_CONDUCT](https://github.com/bmf-san/godra/blob/main/.github/CODE_OF_CONDUCT.md)
- [CONTRIBUTING](https://github.com/bmf-san/godra/blob/main/.github/CONTRIBUTING.md)

# Sponsor
If you like it, I would be happy to have you sponsor it!

[Github Sponsors - bmf-san](https://github.com/sponsors/bmf-san)

Or I would be happy to get a STAR.

It motivates me to keep up with ongoing maintenance. :D

# License
Based on the MIT License.

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