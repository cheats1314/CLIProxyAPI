# CLIProxyAPI fastMode 分支

[English](README.md) | 中文 | [日本語](README_JA.md)

这个 fork 为 CLIProxyAPI 增加了一个 **CPA-only** 的桥接能力：把 Claude Code 的 `fastMode` 映射为 Codex 的 `service_tier: "priority"`。

## 这个 fork 做了什么

- 读取 Claude Code 的 `settings.json`
- 检测 `fastMode: true`
- 只对来源为 Claude 且最终解析到基座模型 `gpt-5.4` 的请求生效
- 在最终发往 Codex 的 `/responses` 请求体中注入 `service_tier: "priority"`
- 用内存缓存配置，并在 `settings.json` 文件变化时自动刷新
- 不需要修改 CCS

## 为什么需要这个 fork

静态的 CPA payload 配置在当前活跃的 Codex `/responses` 执行路径上，不能稳定地强制写入 `service_tier`。
因此这个 fork 把逻辑放到了 `internal/runtime/executor/codex_executor.go`，也就是最终组装 Codex 请求体的位置。

## 这个 fork 改了哪些文件

- `internal/runtime/executor/codex_executor.go`
- `internal/runtime/executor/codex_executor_fastmode_test.go`
- `README.md`
- `README_CN.md`
- `README_JA.md`

## 构建

Linux/macOS:

```bash
go build -o cli-proxy-api ./cmd/server
```

Windows PowerShell:

```powershell
go build -o cli-proxy-api.exe .\cmd\server
```

## 使用方法

### 默认行为

默认情况下，CPA 会读取：

- Linux/macOS：`~/.claude/settings.json`
- Windows：`%USERPROFILE%\\.claude\\settings.json`

如果该文件包含：

```json
{
  "fastMode": true
}
```

那么来源为 Claude、最终走到 Codex 的 `gpt-5.4` 请求，会携带：

```json
{
  "service_tier": "priority"
}
```

如果 `fastMode` 不存在或为 `false`，CPA 会发送正常请求体。

### 可选的配置目录覆盖

你也可以用 `CLAUDE_CONFIG_DIR` 指向另一个 Claude 配置目录。

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

## 测试

运行聚焦测试：

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

## 之后如何在 upstream CPA 更新后维护这个 fork

推荐流程：

1. 先添加 upstream 远端：

```bash
git remote add upstream https://github.com/router-for-me/CLIProxyAPI.git
```

2. 拉取 upstream 更新：

```bash
git fetch upstream
```

3. 切到你的功能分支并 rebase：

```bash
git checkout feat/claude-fastmode-codex-priority
git rebase upstream/main
```

4. 重新跑测试：

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

5. 如果 upstream 修改了 Codex 请求组装逻辑，优先检查：

- `internal/runtime/executor/codex_executor.go`
- `internal/translator/codex/claude/codex_claude_request.go`
- 所有在最终发往 Codex `/responses` 之前会改 body 的 helper

维护经验：

- 如果 upstream 只是更新别的 provider 或文档，这个 fork 通常可以直接 rebase。
- 如果 upstream 改了 Codex 请求组装逻辑，就把注入逻辑重新落在“最终写出 payload”附近。
- 保持补丁足够小，后续维护会轻松很多。

## License

本 fork 仍然沿用 upstream 的 MIT License。详见 [LICENSE](LICENSE)。
