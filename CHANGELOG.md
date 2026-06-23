# Changelog

## [2.0.1](https://github.com/bmf-san/gondola/compare/2.0.0...2.0.1) (2026-06-23)


### Continuous Integration

* drop unsupported linux/arm/v6 from docker release platforms ([#86](https://github.com/bmf-san/gondola/issues/86)) ([fe46c64](https://github.com/bmf-san/gondola/commit/fe46c645cb0671db598a26e074c838f52df6b637))

## [2.0.0](https://github.com/bmf-san/gondola/compare/1.2.0...2.0.0) (2026-06-23)


### ⚠ BREAKING CHANGES

* log_level is now a string (debug/info/warn/error) instead of an int. NewGondola is variadic (Option), NewLogger takes (slog.Level, path), and the exported Server/NewProxyServer type was removed.

### Features

* custom error pages by status code ([#81](https://github.com/bmf-san/gondola/issues/81)) ([bc5054d](https://github.com/bmf-san/gondola/commit/bc5054df7e135abbce3ad395293380d4d28a0cfd))
* customizable access log format (json/common/combined/custom) ([#80](https://github.com/bmf-san/gondola/issues/80)) ([2bcdec4](https://github.com/bmf-san/gondola/commit/2bcdec45513bc6864267028be2d3e98c5c033aed))
* redesign config, logging, routing and lifecycle ([45f56a4](https://github.com/bmf-san/gondola/commit/45f56a43991a0fdc3c493ff4ac0b66f54a25d346))


### Documentation

* add a no-Docker standalone example ([ceb67c2](https://github.com/bmf-san/gondola/commit/ceb67c2edc033fc49d9abd9aeb899dae035a33a6))
* align README and examples with the v2 design ([68416e3](https://github.com/bmf-san/gondola/commit/68416e3a2f439cd2c6fb41bfd28203242b8143a4))
* build the documentation site with gohan ([#82](https://github.com/bmf-san/gondola/issues/82)) ([42ea740](https://github.com/bmf-san/gondola/commit/42ea740822937823a560573389e3a77136203726))
* fix stale master-branch links in READMEs and CONTRIBUTING ([#83](https://github.com/bmf-san/gondola/issues/83)) ([285fe62](https://github.com/bmf-san/gondola/commit/285fe62179fc19a50e8baa247ca3603ba6e8aed9))
* replace the Docker Compose example with standalone ([9872663](https://github.com/bmf-san/gondola/commit/987266394e1c22580fefbcd475078be1d6560a67))


### Tests

* add benchmark tests for proxy, static, and router ([#79](https://github.com/bmf-san/gondola/issues/79)) ([1bc5147](https://github.com/bmf-san/gondola/commit/1bc5147988fef604aa69aed259c6a1313641d7cc))


### Continuous Integration

* add operational automation workflows ([1833caf](https://github.com/bmf-san/gondola/commit/1833caf2556f210ab9ef36f4811349318d2a2c3b))
* automate releases with release-please ([#84](https://github.com/bmf-san/gondola/issues/84)) ([876b4ac](https://github.com/bmf-san/gondola/commit/876b4ac99df1088ca4397fe893dd0a062ce73a0b))
* harden workflows and migrate goreleaser to v2 ([ffb9c89](https://github.com/bmf-san/gondola/commit/ffb9c89131abee0bf6491657d02e98c6a9d4a08e))
