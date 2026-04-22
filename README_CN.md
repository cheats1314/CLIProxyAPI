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

### 环境要求

- 已安装 Go 工具链
- 已拉取本 fork 源码

### 从源码构建

Linux/macOS:

```bash
go build -o cli-proxy-api ./cmd/server
```

Windows PowerShell:

```powershell
go build -o cli-proxy-api.exe .\cmd\server
```

### 以后 upstream CPA 更新时，真正需要维护哪些文件

这个 fork 的功能补丁刻意保持得很小。
正常的小版本更新里，最关键的是：

- 已修改：`internal/runtime/executor/codex_executor.go`
- 新增：`internal/runtime/executor/codex_executor_fastmode_test.go`

README 系列文件只是文档。

很多情况下，你确实只需要把上面这份代码补丁继续带上，然后重新编译即可。
但不要机械地盲目覆盖，还是要先确认 upstream 有没有改动最终的 Codex payload 组装路径。
如果 `codex_executor.go` 依然负责最终 `/responses` 请求体落地，那么通常替换这几个代码文件再重编译就够了。

### 重编译后的快速验证

```bash
go test ./internal/runtime/executor -run 'TestApplyClaudeFastServiceTier|TestClaudeFastModeEnabledReloadsWhenSettingsFileChanges'
```

## 使用方法

### 如何安装到你自己的电脑上

如果你当前实际运行的 CPA 二进制是：

- `/home/cheat/cliproxyapi/cli-proxy-api`

那么最实用的安装流程是：

1. 先备份现有二进制
2. 用本 fork 编译出的新二进制覆盖它
3. 重启 CPA
4. 验证 Claude `fastMode` 是否真的变成 Codex `service_tier: "priority"`

Linux 示例：

```bash
cp /home/cheat/cliproxyapi/cli-proxy-api /home/cheat/cliproxyapi/cli-proxy-api.bak
cp ./cli-proxy-api /home/cheat/cliproxyapi/cli-proxy-api
```

如果你的 CPA 由 systemd 或其他 supervisor 管理，就按你平时的方式重启服务。
如果你是手工启动，就先停掉旧进程，再启动新二进制。

如果你想更稳一点，也可以先复制一份配置，在另一个端口启动新二进制做验证，确认没问题后再替换现网。

### 以后升级这个 fork 的最小流程

对于小更新，你的常规流程可以是：

1. 同步 upstream
2. 保留或重新应用 `codex_executor.go` 里的补丁
3. 保留 `codex_executor_fastmode_test.go`
4. 重新编译
5. 重新运行聚焦测试
6. 替换你本机安装的 CPA 二进制

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
