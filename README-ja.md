# Gondola

シンプルで柔軟な Go 製のリバースプロキシ

## 特徴
- 軽量で高速な実行
- シンプルな設定ファイル（YAML）
- 静的ファイルのホスティング
- バーチャルホストサポート
- 詳細なアクセスログ（nginx互換）
- アクセスログの構造化出力（JSON）
- TLS/SSL対応
- トレースID対応
- タイムアウト制御

## インストール

### Go
```bash
go get github.com/bmf-san/gondola
```

### バイナリ
[Releases](https://github.com/bmf-san/gondola/releases) ページから最新のバイナリをダウンロード

### Docker
```bash
docker pull bmf-san/gondola
docker run -v $(pwd)/config.yaml:/etc/gondola/config.yaml bmf-san/gondola
```

## 使用方法

### コマンドラインオプション

```bash
Usage: gondola [options]

Options:
  -config string    設定ファイルのパス（デフォルト: /etc/gondola/config.yaml）
  -version         バージョン情報を表示
  -help           ヘルプを表示
```

### 環境変数

以下の環境変数でコマンドラインオプションを上書きできます：

- `GONDOLA_CONFIG`: 設定ファイルのパス
- `GONDOLA_LOG_LEVEL`: ログレベル（debug, info, warn, error）

### シグナルハンドリング

Gondolaは以下のシグナルを処理します：

- `SIGTERM`, `SIGINT`: グレースフルシャットダウン
- `SIGHUP`: 設定ファイルの再読み込み
- `SIGUSR1`: アクセスログの再オープン

### 設定ファイル

```yaml
proxy:
  port: "8080"
  read_header_timeout: 2000  # ミリ秒
  shutdown_timeout: 3000     # ミリ秒
  log_level: "info"         # debug, info, warn, error
  static_files:
    - path: /public/
      dir: /path/to/public

upstreams:
  - host_name: api.example.com
    target: http://localhost:3000
    read_timeout: 5000      # ミリ秒
    write_timeout: 5000     # ミリ秒
  - host_name: web.example.com
    target: http://localhost:8000
```

### 起動例

基本的な起動：
```bash
gondola -config config.yaml
```

デバッグモードで起動：
```bash
GONDOLA_LOG_LEVEL=debug gondola -config config.yaml
```

Docker での起動：
```bash
docker run -v $(pwd)/config.yaml:/etc/gondola/config.yaml \
          -p 8080:8080 \
          bmf-san/gondola
```

systemd での起動：
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

## アクセスログ

nginxと互換性のある詳細なアクセスログを出力します。

### ログフィールド

1. クライアント情報
- `remote_addr`: クライアントのIPアドレス
- `remote_port`: クライアントのポート番号
- `x_forwarded_for`: X-Forwarded-Forヘッダー

2. リクエスト情報
- `method`: HTTPメソッド（GET, POST等）
- `request_uri`: フルリクエストURI
- `query_string`: クエリパラメータ
- `host`: Hostヘッダー
- `request_size`: リクエストサイズ

3. レスポンス情報
- `status`: HTTPステータス
- `body_bytes_sent`: レスポンスボディのサイズ
- `bytes_sent`: レスポンス全体のサイズ
- `request_time`: リクエスト処理時間（秒）

4. アップストリーム情報
- `upstream_addr`: バックエンドのアドレス
- `upstream_status`: バックエンドのステータス
- `upstream_size`: バックエンドのレスポンスサイズ
- `upstream_response_time`: バックエンドの応答時間（秒）

5. その他のヘッダー
- `referer`: Refererヘッダー
- `user_agent`: User-Agentヘッダー

### ログ出力例

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
