# CLIProxyAPI fastMode fork

English | [中文](README_CN.md) | [日本語](README_JA.md)

This fork adds a CPA-only bridge from Claude Code `fastMode` to Codex `service_tier: "priority"`.

## What this fork does

- Reads Claude Code `settings.json`
- Detects `fastMode: true`
- Applies only to Claude-originated requests that resolve to base model `gpt-5.4`
- Injects `service_tier: "priority"` into the final upstream Codex request body
- Caches the setting in memory and reloads automatically when `settings.json` changes on disk
- Requires no CCS changes
- Covers both Codex HTTP executor and Codex websocket executor paths

## Why this fork exists

Claude Code `fastMode` is stored in local Claude settings, but the incoming CPA request does not expose a clean fast field that upstream CPA can consume directly.
This fork implements a local bridge inside CLIProxyAPI so existing Claude Code sessions can drive Codex priority mode without changing CCS.

A previous version only patched part of the Codex path. This refreshed fork is based on upstream `v6.9.34` and patches both real execution paths used by Codex requests.

## Files changed in this fork

- `internal/runtime/executor/codex_executor.go`
- `internal/runtime/executor/codex_websockets_executor.go`
- `internal/runtime/executor/codex_executor_fastmode_test.go`
- `README.md`
- `README_CN.md`
- `README_JA.md`

## Build

### Requirements

- Go toolchain installed
- A working checkout of this fork

### Build from source

Linux/macOS:

```bash
go build -o cli-proxy-api ./cmd/server
```

Windows PowerShell:

```powershell
go build -o cli-proxy-api.exe .\cmd\server
```

### What to carry forward when upstream CPA updates

The functional patch in this fork is intentionally small.
For normal upstream updates, the important files are:

- Modified: `internal/runtime/executor/codex_executor.go`
- Modified: `internal/runtime/executor/codex_websockets_executor.go`
- Added: `internal/runtime/executor/codex_executor_fastmode_test.go`

The README files are documentation only.

In many upstream updates, you can re-apply just the code patch above and rebuild.
That said, always verify whether upstream changed the final Codex payload assembly path before blindly copying files.
If `codex_executor.go` and `codex_websockets_executor.go` still own the final outgoing Codex request body, replacing these code files and rebuilding is usually enough.

### Quick verification after rebuilding

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

## Usage

### Install on your own computer

If you already run CPA from:

- `/home/cheat/cliproxyapi/cli-proxy-api`

then the practical install flow is:

1. Back up the current binary
2. Replace it with the fork build
3. Restart CPA
4. Verify that Claude `fastMode` produces Codex `service_tier: "priority"`

Example on Linux:

```bash
cp /home/cheat/cliproxyapi/cli-proxy-api /home/cheat/cliproxyapi/cli-proxy-api.bak
cp ./cli-proxy-api /home/cheat/cliproxyapi/cli-proxy-api
```

If CPA is managed by systemd or another supervisor, restart it using your normal service command.
If you launch it manually, stop the old process and start the new binary.

If you want a safer rollout, run the new binary on a different port first with a copied config, verify behavior, then replace the live binary.

### Default behavior

By default CPA reads:

- Linux/macOS: `~/.claude/settings.json`
- Windows: `%USERPROFILE%\\.claude\\settings.json`

When that file contains:

```json
{
  "fastMode": true
}
```

Claude-originated `gpt-5.4` requests forwarded to Codex will carry:

```json
{
  "service_tier": "priority"
}
```

If `fastMode` is absent or `false`, CPA sends the normal payload.

### Optional config directory override

You can point CPA at a different Claude config directory with `CLAUDE_CONFIG_DIR`.

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

## Tests

Run the focused tests:

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

## Maintaining this fork after upstream CPA updates

Recommended workflow:

1. Add upstream remote once:

```bash
git remote add upstream https://github.com/router-for-me/CLIProxyAPI.git
```

2. Fetch upstream updates:

```bash
git fetch upstream
```

3. Rebase or merge your branch onto the new upstream release branch or commit:

```bash
git checkout feat/claude-fastmode-codex-priority
git rebase upstream/main
```

4. Re-run tests:

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

5. If upstream changed Codex assembly logic, review these locations first:

- `internal/runtime/executor/codex_executor.go`
- `internal/runtime/executor/codex_websockets_executor.go`
- `internal/translator/codex/claude/codex_claude_request.go`
- any helper called right before final Codex dispatch

Maintenance rule of thumb:

- If upstream only changes unrelated providers or docs, your fork should rebase cleanly.
- If upstream changes Codex request assembly, re-apply or adjust the injection near the final payload write step.
- Keep the patch minimal so future rebases stay easy.

## License

This project remains under the upstream MIT license. See [LICENSE](LICENSE).
