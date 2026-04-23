# CLIProxyAPI fastMode フォーク

[English](README.md) | [中文](README_CN.md) | 日本語

この fork は CLIProxyAPI に **CPA-only** のブリッジ機能を追加し、Claude Code の `fastMode` を Codex の `service_tier: "priority"` に変換します。

## この fork が行うこと

- Claude Code の `settings.json` を読む
- `fastMode: true` を検出する
- Claude 由来で、最終的にベースモデル `gpt-5.4` に解決されるリクエストにのみ適用する
- 最終的に Codex へ送る request body に `service_tier: "priority"` を注入する
- 設定をメモリにキャッシュし、`settings.json` が変更されたら自動で再読込する
- CCS の変更は不要
- Codex HTTP executor と websocket executor の両方をカバーする

## なぜこの fork が必要か

Claude Code の `fastMode` はローカルの Claude 設定に保存されますが、現在 CPA に入ってくるリクエストには、それを直接再利用できるきれいな fast フィールドがありません。
そのため、この fork では CLIProxyAPI 内部でローカルブリッジを実装し、CCS を変更せずに Codex priority モードを有効にします。

前回版は Codex の一部経路しかカバーしていませんでした。今回は upstream `v6.9.34` をベースにし、実際に命中する 2 つの実行経路を両方修正しています。

## この fork で変更したファイル

- `internal/runtime/executor/codex_executor.go`
- `internal/runtime/executor/codex_websockets_executor.go`
- `internal/runtime/executor/codex_executor_fastmode_test.go`
- `README.md`
- `README_CN.md`
- `README_JA.md`

## ビルド

### 必要条件

- Go ツールチェーンがインストールされていること
- この fork のソースを取得していること

### ソースからビルド

Linux/macOS:

```bash
go build -o cli-proxy-api ./cmd/server
```

Windows PowerShell:

```powershell
go build -o cli-proxy-api.exe .\cmd\server
```

### upstream CPA 更新時に実際に持ち回るべきファイル

この fork の機能パッチは意図的に小さく保っています。
通常の小さな upstream 更新では、重要なのは次です。

- 変更済み: `internal/runtime/executor/codex_executor.go`
- 変更済み: `internal/runtime/executor/codex_websockets_executor.go`
- 追加済み: `internal/runtime/executor/codex_executor_fastmode_test.go`

README 系はドキュメントのみです。

多くのケースでは、このコードパッチを維持して再ビルドするだけで足ります。
ただし、機械的に上書きする前に upstream が最終 Codex payload 組み立て経路を変えていないか確認してください。
`codex_executor.go` と `codex_websockets_executor.go` が引き続き最終の Codex request body を担当しているなら、これらのコードファイルを反映して再ビルドすれば通常は十分です。

### 再ビルド後の簡易確認

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

## 使い方

### 自分のPCへインストールする方法

現在動かしている CPA バイナリが次なら：

- `/home/cheat/cliproxyapi/cli-proxy-api`

実用的な導入手順は次です。

1. 既存バイナリをバックアップする
2. この fork でビルドした新バイナリで置き換える
3. CPA を再起動する
4. Claude `fastMode` が Codex `service_tier: "priority"` になることを確認する

Linux 例：

```bash
cp /home/cheat/cliproxyapi/cli-proxy-api /home/cheat/cliproxyapi/cli-proxy-api.bak
cp ./cli-proxy-api /home/cheat/cliproxyapi/cli-proxy-api
```

CPA を systemd や他の supervisor で管理しているなら、普段どおりの方法で再起動してください。
手動起動なら旧プロセスを止めて新バイナリを起動します。

より安全に進めたい場合は、設定をコピーして別ポートで新バイナリを先に動かし、確認後に本番バイナリを置き換えてください。

### デフォルト動作

デフォルトでは CPA は次を読みます：

- Linux/macOS: `~/.claude/settings.json`
- Windows: `%USERPROFILE%\\.claude\\settings.json`

そのファイルに以下が含まれている場合：

```json
{
  "fastMode": true
}
```

Claude 由来で Codex の `gpt-5.4` に到達するリクエストには、次が入ります：

```json
{
  "service_tier": "priority"
}
```

`fastMode` が存在しない、または `false` の場合、CPA は通常の payload を送ります。

### 設定ディレクトリの上書き

`CLAUDE_CONFIG_DIR` を使って別の Claude 設定ディレクトリを指定できます。

Linux/macOS:

```bash
export CLAUDE_CONFIG_DIR="$HOME/.claude"
./cli-proxy-api
```

Windows PowerShell:

```powershell
$env:CLAUDE_CONFIG_DIR = "$HOME/.claude"
.\cli-proxy-api.exe
```

## テスト

対象テストだけを実行：

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

## upstream CPA 更新後にこの fork を保守する方法

推奨フロー：

1. upstream を追加：

```bash
git remote add upstream https://github.com/router-for-me/CLIProxyAPI.git
```

2. upstream を取得：

```bash
git fetch upstream
```

3. あなたの機能ブランチに切り替えて rebase：

```bash
git checkout feat/claude-fastmode-codex-priority
git rebase upstream/main
```

4. テストを再実行：

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

5. upstream が Codex payload 組み立てを変更していたら、まず次を確認：

- `internal/runtime/executor/codex_executor.go`
- `internal/runtime/executor/codex_websockets_executor.go`
- `internal/translator/codex/claude/codex_claude_request.go`
- 最終的な Codex dispatch の直前で body を変更する helper

保守のコツ：

- upstream が他の provider や docs だけを変えた場合、この fork はそのまま rebase できることが多いです。
- upstream が Codex request assembly を変えた場合は、注入ロジックを「最終 payload 書き込みの近く」に再配置してください。
- パッチを小さく保つほど、今後の rebase は楽になります。

## License

この fork は upstream の MIT License をそのまま使います。詳細は [LICENSE](LICENSE) を参照してください。
