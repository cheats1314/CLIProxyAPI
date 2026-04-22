# CLIProxyAPI fastMode フォーク

[English](README.md) | [中文](README_CN.md) | 日本語

この fork は CLIProxyAPI に **CPA-only** のブリッジ機能を追加し、Claude Code の `fastMode` を Codex の `service_tier: "priority"` に変換します。

## この fork が行うこと

- Claude Code の `settings.json` を読む
- `fastMode: true` を検出する
- Claude 由来で、最終的にベースモデル `gpt-5.4` に解決されるリクエストにのみ適用する
- 最終的に Codex へ送る `/responses` リクエスト body に `service_tier: "priority"` を注入する
- 設定をメモリにキャッシュし、`settings.json` が変更されたら自動で再読込する
- CCS の変更は不要

## なぜこの fork が必要か

静的な CPA payload ルールでは、現在の Codex `/responses` 実行経路で `service_tier` を安定して強制できませんでした。
そのため、この fork では最終的な Codex payload を組み立てる `internal/runtime/executor/codex_executor.go` にロジックを入れています。

## この fork で変更したファイル

- `internal/runtime/executor/codex_executor.go`
- `internal/runtime/executor/codex_executor_fastmode_test.go`
- `README.md`
- `README_CN.md`
- `README_JA.md`

## ビルド

Linux/macOS:

```bash
go build -o cli-proxy-api ./cmd/server
```

Windows PowerShell:

```powershell
go build -o cli-proxy-api.exe .\cmd\server
```

## 使い方

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
- `internal/translator/codex/claude/codex_claude_request.go`
- 最終的な Codex `/responses` dispatch の直前で body を変更する helper

保守のコツ：

- upstream が他の provider や docs だけを変えた場合、この fork はそのまま rebase できることが多いです。
- upstream が Codex request assembly を変えた場合は、注入ロジックを「最終 payload 書き込みの近く」に再配置してください。
- パッチを小さく保つほど、今後の rebase は楽になります。

## License

この fork は upstream の MIT License をそのまま使います。詳細は [LICENSE](LICENSE) を参照してください。
