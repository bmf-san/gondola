[English](https://github.com/bmf-san/gondola) [日本語](https://github.com/bmf-san/gondola/blob/master/README-ja.md)

# gondola
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![GitHub release](https://img.shields.io/github/release/bmf-san/gondola.svg)](https://github.com/bmf-san/gondola/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/bmf-san/gondola)](https://goreportcard.com/report/github.com/bmf-san/gondola)
[![codecov](https://codecov.io/gh/bmf-san/gondola/branch/main/graph/badge.svg?token=ZLOLQKUD39)](https://codecov.io/gh/bmf-san/gondola)
[![GitHub license](https://img.shields.io/github/license/bmf-san/gondola)](https://github.com/bmf-san/gondola/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmf-san/gondola.svg)](https://pkg.go.dev/github.com/bmf-san/gondola)
[![Sourcegraph](https://sourcegraph.com/github.com/bmf-san/gondola/-/badge.svg)](https://sourcegraph.com/github.com/bmf-san/gondola?badge)

Golangのリバースプロキシです。

<img src="https://storage.googleapis.com/gopherizeme.appspot.com/gophers/22fd9b7a49eac4101fc9819578641c2e71706f6f.png" alt="gondola" title="gondola" width="250px">

This log was created by [gopherize.me](https://gopherize.me/gopher/22fd9b7a49eac4101fc9819578641c2e71706f6f)

# 目次
- [gondola](#gondola)
- [目次](#目次)
- [特徴](#特徴)
- [インストール](#インストール)
  - [Go](#go)
  - [Binary](#binary)
  - [Docker](#docker)
- [例](#例)
- [使い方](#使い方)
- [Projects](#projects)
- [ADR](#adr)
- [Wiki](#wiki)
- [コントリビューション](#コントリビューション)
- [スポンサー](#スポンサー)
- [ライセンス](#ライセンス)
- [Stargazers](#stargazers)
- [Forkers](#forkers)
- [作者](#作者)

# 特徴
- バーチャルホスト
  - アップストリームサーバーに複数のホストを設定できます。
- 設定ファイルローダー
  - YAML形式の設定ファイルを使用できます。
- TLS
  - TLS証明書を用意することでTLSを利用できます。
- 静的ファイルの配信
  - 静的ファイルを配信できます。
- アクセスログ
  - Proxyのアクセスログとアップストリームのサーバーアクセスログを出力します。
- バイナリ配布
  - クロスコンパイルされたバイナリを配布しています。

# インストール
## Go
```
go get -u github.com/bmf-san/gondola
```

## Binary
[release page](https://github.com/bmf-san/gondola/releases)からバイナリをダウンロードして利用できます。

## Docker
[bmfsan/gondola](https://hub.docker.com/r/bmfsan/gondola)

# 例
以下のリンクからgondolaの使い方を参照できます。

- [_examples](https://github.com/bmf-san/gondola/tree/main/_examples)

# 使い方
オプションを指定してバイナリを実行します。

```sh
gondola -config config.yaml
```

# Projects
- [The gondola's board](https://github.com/users/bmf-san/projects/1/views/1)

# ADR
- [ADR](https://github.com/bmf-san/gondola/discussions?discussions_q=is%3Aopen+label%3AADR)

# Wiki
- [wiki](https://github.com/bmf-san/gondola/wiki)

# コントリビューション
IssueやPull Requestはいつでもお待ちしています。

気軽にコントリビュートしてもらえると嬉しいです。

コントリビュートする際は、以下の資料を事前にご確認ください。

- [CODE_OF_CONDUCT](https://github.com/bmf-san/godra/blob/main/.github/CODE_OF_CONDUCT.md)
- [CONTRIBUTING](https://github.com/bmf-san/godra/blob/main/.github/CONTRIBUTING.md)

# スポンサー
もし気に入って頂けたのならスポンサーしてもらえると嬉しいです！

[Github Sponsors - bmf-san](https://github.com/sponsors/bmf-san)

あるいはstarを貰えると嬉しいです！

継続的にメンテナンスしていく上でのモチベーションになります :D

# ライセンス
MITライセンスに基づいています。

[LICENSE](https://github.com/bmf-san/gondola/blob/main/LICENSE)

# Stargazers
[![Stargazers repo roster for @bmf-san/gondola](https://reporoster.com/stars/bmf-san/gondola)](https://github.com/bmf-san/gondola/stargazers)

# Forkers
[![Forkers repo roster for @bmf-san/gondola](https://reporoster.com/forks/bmf-san/gondola)](https://github.com/bmf-san/gondola/network/members)

# 作者
[bmf-san](https://github.com/bmf-san)

- Email
  - bmf.infomation@gmail.com
- Blog
  - [bmf-tech.com](http://bmf-tech.com)
- Twitter
  - [bmf-san](https://twitter.com/bmf-san)