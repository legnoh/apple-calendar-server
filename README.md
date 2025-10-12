# apple-calendar-server

`apple-calendar-cli` コマンドを HTTP 経由で叩いた結果を配信するラッパーサーバ。

alecthomas/kong を利用したシンプルなCLI。

## ビルド

```
go build ./cmd/calendar-server
```

生成されたバイナリ例: `./calendar-server`

## 使い方

```
./calendar-server \
	--command apple-calendar \
	--listen :8080 \
	--endpoint /calendar \
	--timeout 15s
```

### 主なフラグ

| フラグ | 説明 | デフォルト |
|--------|------|-------------|
| --command | 実行するコマンド (必須) | なし |
| --command-arg | 常に付与する追加引数 (複数指定可) | なし |
| --listen | HTTPリッスンアドレス | :8080 |
| --endpoint | コマンドを叩くHTTPパス | /calendar |
| --timeout | コマンド実行タイムアウト | 10s |
| --allow-origin | CORS許可Origin | (空) |
| -v / --version | バージョン表示 |  | 

### リクエストでの追加引数

HTTPアクセス時にクエリ `?arg=foo&arg=bar` を付けると実行時引数に追加されます。

### 応答フォーマット

```
{
	"command": "apple-calendar",
	"args": ["--some", "option"],
	"stdout": { ... または [..] もしくは 文字列 },
	"stderr": "",
	"error": "",        // Go上でのエラー文字列 (終了コード!=0 等)
	"runtime": "12.345ms"
}
```

`stdout` が JSON オブジェクト/配列で始まらない場合は JSON 文字列としてラップされます。

### 例: curl で取得

```
curl -s http://localhost:8080/calendar | jq
```

追加引数:

```
curl -s 'http://localhost:8080/calendar?arg=--today'
```

### 簡易テスト用 (echo)

`echo '{"ok":true}'` をラップする場合:

```
./calendar-server --command echo --command-arg '{"ok":true}'
curl -s http://localhost:8080/calendar | jq
```

## ライセンス

MIT
