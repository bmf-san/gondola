---
title: "Signals & Reload"
description: "Graceful shutdown, configuration reload, and log rotation."
slug: "signals"
categories:
  - features
---

Gondola handles the following signals:

| Signal | Behavior |
|---|---|
| `SIGINT` / `SIGTERM` | Graceful shutdown (waits up to `shutdown_timeout`) |
| `SIGHUP` | Reload the configuration (routing) without dropping connections |
| `SIGUSR1` | Reopen the access log file (for rotation) |

On `SIGHUP`, gondola re-reads the config file, validates it, and atomically
swaps the router. The running configuration is only replaced when the new one
builds successfully (otherwise the current one is kept). Changes to the
listening port and server-level timeouts require a restart.

> On Windows, only `SIGINT` / `SIGTERM` are supported.
