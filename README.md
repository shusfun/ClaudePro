# ClaudePro

中文 Windows / macOS 环境切换与精确还原工具。采用 Wails v3.0.0-alpha2.117、Go 1.25、React 19.3.1、TypeScript、Vite 6。前端使用从 MiGate 迁入的 UIAble 组件、Base UI 1.9.0 和 Tailwind CSS 4.3.3。产物为便携式 Windows x64 EXE（`bin/claudepro.exe`）和 macOS Universal 应用（Apple Silicon + Intel，`bin/claudepro.app`）。

只管理系统时区、当前用户区域格式和国家地区（macOS 对应 `/etc/localtime`、`AppleLocale` 与 `Country`）。修改前保存原始值，逐项回读验证，支持失败回滚、中断恢复和外部配置冲突选择。支持 Claude Code CLI、MSIX 和普通安装的 Desktop。不提供防封保证。

## 下载

在 [Releases](https://github.com/shusfun/ClaudePro/releases) 下载 Windows x64 的 `claudepro.exe` 和 SHA256 校验文件。需要 Microsoft Edge WebView2 Runtime。首版为预发布版本，系统修改与恢复建议先在可还原的测试环境验证。

## 使用

打开应用，首页顶部的状态卡显示当前处境和唯一的主操作。在下方选择地区模板（美国西部/东部、英国、德国、日本、新加坡），状态卡即时显示“当前 → 目标”差异。关闭 Claude 会话后点击“切换环境”，完成工作后点击“恢复原环境”。界面普通权限运行，修改时区时 Windows 请求 UAC、macOS 请求管理员密码；macOS 会同时关闭“自动设置时区”，恢复时还原原值。

自定义选择器支持名称、代码和时区 ID 搜索。客户端路径在标题栏齿轮打开的“客户端设置”中保存。有未恢复记录时锁定目标，先恢复再更换地区。

首页“检测并推荐地区”可查询出口城市和时区，并推荐对应模板。美西 / 美东按 IP 地理时区判断，位置不明确或其他美国时区则提示自定义。“采用推荐”只填预览，仍需点“切换环境”。

“环境检测”自动检查本机设置、客户端、会话和恢复记录；手动点击才使用 GeoJS、ipwho.is、ipinfo、ipapi 多源查询出口信息，自动跳过限流源，显示来源、冷却倒计时和地区分歧。网络结果仅在当前会话保留，不代表 Claude 专属路由或账号风险。`--demo` 检测使用固定数据，不联网。

窗口使用自绘标题栏：Windows 为无边框窗口（保留系统阴影与圆角），右上角为最小化 / 最大化 / 关闭；macOS 保留系统红绿灯。标题栏中间显示环境状态，右侧为刷新、主题与客户端设置。主题循环切换浅色、深色与跟随系统，默认跟随系统。窄窗口时左侧导航收成图标条。

- [使用、恢复和构建指南](docs/claudepro-guide.html)
- [验证记录和 Windows 虚拟机验收步骤](docs/validation.md)
- [已有指纹相关调研](docs/claude-ban-fingerprint-research.html)

```powershell
# 模拟模式：独立数据目录，不改系统，不启动真实 Claude
.\bin\claudepro.exe --demo

# 只读诊断
Start-Process .\bin\claudepro.exe -ArgumentList '--diagnose' -WindowStyle Hidden -Wait -RedirectStandardOutput .\bin\diagnose.json
```

正式备份在 `%LOCALAPPDATA%\ClaudePro`（macOS：`~/Library/Application Support/ClaudePro`），模拟数据在同级的 `ClaudePro-Demo`。存在未恢复操作时请保留整个备份目录。

## 构建与测试

Windows x64、Go 1.25.0+、Node.js 22.18.0+、npm、WebView2 Evergreen Runtime。依赖版本固定于 go.mod / go.sum 和 frontend/package-lock.json。

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha2.117
npm --prefix frontend ci
# 首次克隆先构建嵌入资源，再运行 Wails 绑定生成
npm --prefix frontend run build
wails3 build ARCH=amd64

go test ./...
go vet ./...
npm --prefix frontend run typecheck
npm --prefix frontend run build
Push-Location frontend
npx playwright test
Pop-Location
```

输出 `bin/claudepro.exe`。macOS 在本机执行 `wails3 task darwin:package:universal`，输出 Universal `bin/claudepro.app`；CI 的 macOS 任务会上传其 zip。UI 测试使用本机 Edge 和模拟 Wails 接口。普通测试不修改系统设置、不启动或终止真实 Claude。`wails3 dev` 默认使用真实适配器，安全演示请使用 `--demo`。

## 结构

- internal/core：事务状态机、原子 JSON、日志、冲突检查及模拟测试。
- internal/platform：Windows（固定 PowerShell 操作、UAC 时区辅助进程）与 macOS（`defaults`、`systemsetup` 管理员授权、终端 / `open` 启动）适配器，客户端发现和演示适配器。
- main.go / main_windows.go / main_darwin.go / service.go：Wails 接口、自绘标题栏窗口、事件、单实例、退出提醒（含 macOS ⌘Q）、WebView2 检查。
- frontend：中文界面、生成绑定和 Playwright 测试。
- frontend/src/components/ui：MiGate 组件的独立本地副本，保留 LICENSE 与 SOURCE.json 来源记录；构建不需要 MiGate 项目。

已有 docs、research 保留。生成器附带的 Linux / iOS / Android 资源不表示本工具支持这些系统。

当前 EXE 和 .app 均未签名、未公证（macOS 首次打开需在 Finder 中右键 › 打开）；macOS 的真实切换尚未在实机验收；真实系统设置切换、UAC 取消、断电恢复与客户端重启后的行为仍需在可还原的 Windows 测试环境验收。

## 许可证

原创代码采用 [MIT](LICENSE)。第三方组件保留各自许可，见 [第三方说明](THIRD_PARTY_NOTICES.md)。调研抓取快照、原始搜索日志、个人运行数据与本机构建缓存不上传。
