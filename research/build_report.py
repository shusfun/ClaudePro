from pathlib import Path
import json

ROOT=Path(__file__).parent
groups=[]
def group(name,rows):
    groups.append((name,[dict(zip(['title','url','note'],r)) for r in rows]))

group('官方文档：判断政策、数据采集与申诉边界',[
('Anthropic Privacy Policy','https://www.anthropic.com/legal/privacy','已读。明确列出 IP、时区、设备/浏览器信息、设备或广告标识符、概率标识符、使用日志，并说明反欺诈和反滥用用途；没有公布封号权重或阈值。'),
('Safeguards warnings and appeals','https://support.claude.com/en/articles/8241253-why-was-my-account-banned','已读。封禁原因包括反复违反使用政策、不支持地区注册、违反服务条款；申诉入口以该页面当前链接为准。旧 URL 目前显示新标题。'),
('Supported Regions Policy','https://www.anthropic.com/supported-countries','已读。当前规则覆盖应用、API、云平台和授权转售；区分实际所在地区和实体所属地区。不能用“出口 IP 看起来在哪”替代支持资格判断。'),
('Consumer Terms of Service','https://www.anthropic.com/legal/consumer-terms','已读。账号使用、自动化访问限制及终止条款；允许的 API 或其他明确许可场景有例外。'),
('Claude Code — Legal and compliance','https://code.claude.com/docs/en/legal-and-compliance','已读。特别关注 Authentication and credential use：个人订阅 OAuth 与开发者 API key 的用途边界；第三方收集、转接订阅凭据存在明确限制。'),
('Claude Code — Data usage','https://code.claude.com/docs/en/data-usage','已读。说明运行遥测、错误报告、数据保留和官方退出选项；关闭遥测不代表取消正常推理请求及账号认证。'),
('Claude Code — Monitoring','https://code.claude.com/docs/en/monitoring-usage','已读。企业自行配置的 OpenTelemetry 导出；文档明确区分它与 Anthropic 自身运行遥测，不应混为一谈。'),
('Anthropic Cookie Policy','https://www.anthropic.com/legal/cookies','已读。Cookie、脚本及类似技术用于识别、个性化、分析等。没有给出“删 Cookie 即解除封禁”的依据。'),
('Anthropic Subprocessors','https://trust.anthropic.com/subprocessors','官方隐私政策链接到此处。页面动态加载，本轮仅取得页面壳，未核验当前完整供应商名单。'),
])

group('直接相关的逆向研究与技术文章：优先阅读',[
('wtfwhs/tengu-decoded','https://github.com/wtfwhs/tengu-decoded','已读 README。按 Claude Code 版本整理设备标识、遥测事件、功能开关、接口和环境变量，提供提取产物与代码位置。第三方逆向研究，不能据此看到服务端封号规则。'),
('Device Fingerprinting — Claude Code v2.1.300','https://github.com/wtfwhs/tengu-decoded/blob/main/versions/2.1.300/device-fingerprinting.md','已读对应 raw 文档。标注 2026-10-14 构建、linux-x86_64。分析 userID、machineID、summonSidKey、remoteControlMachineId 和设备密钥；区分随机安装标识、系统 machine-id 的本地用途与实际发送字段。'),
('Device Fingerprinting — Claude Code v2.1.169','https://github.com/wtfwhs/tengu-decoded/blob/main/versions/2.1.169/device-fingerprinting.md','已读对应 raw 文档。标注 2026-06-08 构建。作者追踪 device_id 到随机安装 token，提供函数和行号；适合作为版本对照，不能直接外推全部平台和版本。'),
('B1anYu/claude-telemetry-audit：Claude 三端遥测与隐私审计','https://github.com/B1anYu/claude-telemetry-audit','已读 README。覆盖 Desktop macOS 2.9939.2、Code darwin-arm64 2.1.283、Android 1.260930.20。讨论 ant-did、userID、模型请求元数据、第三方端点上报、Android Sift/ANDROID_ID；本轮未独立重跑反编译或抓包。'),
('Claude Code Is Quietly Fingerprinting China-Linked API Routers','https://www.vincentschmalbach.com/claude-code-china-router-fingerprint/','已读。Vincent Schmalbach，2026-06-30。比较 2.1.90/2.1.91 的日期上下文，分析自定义 ANTHROPIC_BASE_URL、主机名分类和时区如何影响日期/标点标记。作者明确未在此次分析中抓取实时流量；不能证明所有中转都把标记传回 Anthropic，更不能直接证明封号因果。'),
])

group('直接相关项目：检测、清理、浏览器与请求特征',[
('yacuo/check-cc','https://github.com/yacuo/check-cc','已读 README。40+ 项环境检查：浏览器语言、时区、Intl、UA、运行容器、网络出口等。项目自己的风险评分，不是 Anthropic 评分；在线入口 checkcc.org 来自仓库。'),
('Nooobhh/claude-analysis','https://github.com/Nooobhh/claude-analysis','已读 README。环境检测和匿名封号问卷，列明规则、数据源及低样本标记；适合查看可观察信号，但自愿问卷存在选择偏差，不能单凭占比推断因果。'),
('cso1z/claude-ip-guard','https://github.com/cso1z/claude-ip-guard','已读 README。Claude Code Hook，检查 IP 地区和变化。检测失败默认放行；检测站出口不一定等于代理分流后所有 Claude 请求的出口。'),
('YOURSJEYO/claude-switcher','https://github.com/YOURSJEYO/claude-switcher','已读 README。环境审计、代理中断处理、状态快照；macOS 为主，Windows/Linux 标注 Beta。“物理上绝不泄漏”等是作者主张，本轮未运行验证。'),
('dai-chao/Agent-Guard','https://github.com/dai-chao/Agent-Guard','已读 README。本机代理、时区、端点、配置残留和密钥扫描，免费扫描/Pro 修复，含商业产品推广；风险分不代表官方规则。'),
('BellyBook/claude-local-cleanup','https://github.com/BellyBook/claude-local-cleanup','已读 README。macOS 本地字段、OAuth Keychain、遥测和历史数据清理脚本；会改配置并删除数据，清理动作不等于服务端解封。'),
('l2461090/claude-post-ban-cleanup','https://github.com/l2461090/claude-post-ban-cleanup','已读 README。Windows/macOS 本地资料清理，强调检查、确认、备份和恢复；作者明确写明不能解除服务端封禁，不保证避免再封。'),
('jaysu66/claude-anti-ban','https://github.com/jaysu66/claude-anti-ban','已读 README。Windows 诊断与清理项目。声称 device_id=SHA256(MachineGuid)、再封概率>90%，但本轮未找到足够支撑；与有明确版本和数据流引用的随机安装 ID 研究冲突，不能直接采信或据此改系统标识。'),
('AnonUsAl/claude-fingerprint-shield','https://github.com/AnonUsAl/claude-fingerprint-shield','已读 README。Chromium MV3 扩展，改变时区、语言、字体、WebRTC 等暴露值。作者明确目标包括降低第三方检测站的加权分；检测分下降不证明 Claude 风控认可。'),
('danger-dream/cc-proxy','https://github.com/danger-dream/cc-proxy','已读 README；GitHub 元数据标为 archived。研究 Claude Code 请求特征注入的反向代理；不是官方兼容或防封保证。'),
('kobie3717/claude-oauth-proxy','https://github.com/kobie3717/claude-oauth-proxy','已读 README。订阅 OAuth 转 API，声称模拟 headers、工具定义、系统提示词与请求节奏。“不可区分/Anti-Ban”未经本轮验证，须结合官方凭据使用限制理解。'),
('rinchao0721/Web-Claude-spoof','https://github.com/rinchao0721/Web-Claude-spoof','仅核验 GitHub 搜索元数据。简介为隐藏/伪装部分浏览器指纹；尚未审计实现和效果。'),
('Hype-yolo/claude-ban-check','https://github.com/Hype-yolo/claude-ban-check','仅核验 GitHub 搜索元数据。简介涵盖 IP/时区/语言、IPv6、浏览器指纹、撇号标记和终端检查；规则依据尚未审计。'),
('zzusec/claude-residue-cleaner','https://github.com/zzusec/claude-residue-cleaner','仅核验 GitHub 搜索元数据。macOS 设备标识、Keychain、浏览器和其他残留审计/清理；清理范围较广，未执行。'),
('fourierwang66666/claude-ban-risk-check','https://github.com/fourierwang66666/claude-ban-risk-check','仅核验 GitHub 搜索元数据。本地 Claude Code 环境风险检查 Skill；未审计评分和规则依据。'),
])

group('封号文章与经验汇总：当作线索，不当作已证实机制',[
('Claude 防封号 & KYC 避坑完整指南（2026最新实测版）','https://www.panewslab.com/zh/articles/019db34c-eb93-707b-9c25-dfe946dc11b1','已读中文页，署名 MyToken。讨论环境、指纹浏览器、IP、支付、KYC。诸如欺诈分<15、申诉成功率<20%未见可审计样本或官方依据；英文页本轮 502。'),
('Claude 防封指南 — 腾讯云开发者社区','https://cloud.tencent.com/developer/news/3854714','已读。2026-04-19，来源为企鹅号“G人工智能”的转载，并非腾讯或 Anthropic 官方技术结论。主要是 IP、设备切换与账号使用经验。'),
('Fishason/Claude-anti-ban','https://github.com/Fishason/Claude-anti-ban','已读 README。封号原因、支付、KYC、VPS 和社区链接汇总；含订阅/VPS 服务导流。固定调用频率、观察期和“极低风险”结论不宜照单全收。'),
('MFLS2023/claude-anti-ban-guide','https://github.com/MFLS2023/claude-anti-ban-guide','已读 README。收录多篇中文手册及 Linux.do 线索；虽声明黑盒局限，仍使用“VPS 必死”“毫秒级识别”等绝对说法，证据不足。适合作为索引。'),
('I was banned from Claude for scaffolding a CLAUDE.md file','https://hugodaniel.com/posts/claude-code-banned-me/','已读。Hugo Daniel 个人封号叙述；HN 于 2026-01-22 收录。作者明确称原因只是其假设，因此不是指纹封号实证。'),
('Why 157,000 developers are hedging against Anthropic with OpenCode','https://thenewstack.io/anthropic-claudecode-opencode-split/','已读。The New Stack 报道第三方客户端、OAuth 限制及开发者反应；适合作为背景，具体时间线与规则以官方文档及仓库提交为准。'),
])

group('GitHub 讨论与变更：保留原始问题上下文',[
('sub2api #2185：怎么抹去 Claude Code 的指纹信息呢','https://github.com/Wei-Shaw/sub2api/issues/2185','已读。2026-05-04 用户称开启 TLS 指纹模拟后仍被封；只能说明该用户遇到问题，未证明设备指纹是原因。'),
('sub2api #580：OAuth 模拟模式 User-Agent 版本不一致','https://github.com/Wei-Shaw/sub2api/issues/580','已读。2026-02-14，有具体 UA 覆盖路径与代码引用。请求一致性问题可检查；是否导致封号仍是发帖者推测。'),
('claude-relay-service #587：这两天封号严重','https://github.com/Wei-Shaw/claude-relay-service/issues/587','已读可见讨论。始于 2025-10-17，含成功和失败案例以及后续营销评论；相似部署结局不同，不支持单因子保证。'),
('claude-code #8327：API key 覆盖订阅造成 organization disabled','https://github.com/anthropics/claude-code/issues/8327','已读。2025-09-29 用户报告凭据/组织选择问题；提醒“组织被禁用”报错不一定意味着当前个人订阅被设备指纹封禁。'),
('claude-code #17118：Support for OpenCode and Max plan','https://github.com/anthropics/claude-code/issues/17118','已读。2026-01-09 社区功能请求及讨论；在官方仓库发帖不等于内容由官方确认。'),
('OpenCode 提交：anthropic legal requests','https://github.com/anomalyco/opencode/commit/973715f3da1839ef2eba62d4140fe7441d539411','已读提交 diff。可查看订阅认证相关代码/提示变更，是第三方客户端支持变化的原始证据，但不披露 Anthropic 风控模型。'),
])

group('底层指纹文档与通用开源项目：帮助理解术语',[
('Cloudflare — JA3/JA4 fingerprint','https://developers.cloudflare.com/bots/additional-configurations/ja3-ja4-fingerprint/','已读官方文档。TLS 客户端类型和握手特征，多个用户可共享相同值。不能把 JA4 直接当成唯一硬件 ID，也不能由 Cloudflare 的能力推断 Claude 配置了哪些规则。'),
('FoxIO-LLC/ja4','https://github.com/FoxIO-LLC/ja4','已读仓库页面。JA4+ 网络指纹规范与实现，覆盖 TLS、HTTP、TCP 等多个维度。非 Claude 专用。'),
('fingerprintjs/fingerprintjs','https://github.com/fingerprintjs/fingerprintjs','已读仓库页面。浏览器指纹基础库。用于理解客户端可见信号；本轮没有证据证明 Claude 使用此特定库。'),
('abrahamjuliot/creepjs','https://github.com/abrahamjuliot/creepjs','已读仓库页面。浏览器指纹、环境一致性与伪装痕迹检测；非 Claude 风控模拟器。'),
('jason5ng32/MyIP','https://github.com/jason5ng32/MyIP','已读仓库页面。IP、网络出口、WebRTC/DNS 等诊断；检测结果描述可见网络环境，不能直接预测账号封禁。'),
('lexiforest/curl_cffi','https://github.com/lexiforest/curl_cffi','已读仓库页面。研究 TLS/HTTP 请求特征和浏览器模拟；不等于复制浏览器 JavaScript 环境或账号历史。'),
('daijro/camoufox','https://github.com/daijro/camoufox','已读仓库页面。Firefox 衍生浏览器与环境指纹一致性工程；没有由此得出 Claude 防封保证。'),
('Stripe Radar — Transaction risk prevention','https://docs.stripe.com/radar/risk-evaluation','已读官方文档。支付风险评估、卡指纹与欺诈信号。支付拒绝和 Claude 账号禁用是不同事件，不能混为一谈。'),
('Android Settings.Secure — ANDROID_ID','https://developer.android.com/reference/android/provider/Settings.Secure#ANDROID_ID','已读 Android 官方参考页面。用于校正 Android 审计中的术语：ANDROID_ID 是具有应用签名/用户/设备作用域的系统标识，不等于不可变的物理硬件序列号。'),
])

group('未能读取正文的文章线索：不据此作结论',[
('Linux.do topic/1797342','https://linux.do/t/topic/1797342','由 MFLS2023 汇总仓库引用，声称是封号/存活案例汇总。本轮返回 403 验证页面；真实标题、内容和样本量未核验。'),
('Linux.do topic/2997319','https://linux.do/t/topic/2997319','由 MFLS2023 汇总仓库引用，声称讨论账号生命周期。本轮 403，未核验原文。'),
('Linux.do topic/2959915','https://linux.do/t/topic/2959915','由 MFLS2023 汇总仓库引用，声称讨论家宽与分流。本轮 403，未核验原文。'),
('博客园 notlate-cn / 20002343','https://www.cnblogs.com/notlate-cn/p/20002343','由 Fishason 汇总仓库引用。访问落到用户中心，未读到文章；不视为正文已验证。'),
('知乎专栏 / 2019847407014293704','https://zhuanlan.zhihu.com/p/2019847407014293704','由 Fishason 汇总仓库引用。本轮 403，未核验标题和正文。'),
])

intro='''# Claude 封号、指纹与遥测资料索引

检索日期：2026-10-18（Asia/Shanghai）。范围：Claude 网页、Desktop、Code、Android，及第三方客户端、浏览器/TLS 指纹、网络出口、支付与申诉。

这是一份尽可能广泛的公开资料筛选，不是“全网所有资料”或对 Anthropic 内部风控的完整还原。主要通过 GitHub 搜索/API、Hacker News Algolia 发现文章，并直接读取官方与作者页面。Google/部分搜索站遇到验证，Bing 多次返回与精确查询无关的宽泛结果，因此未将它们作为有效证据。Linux.do、知乎等未读到正文的线索单独标出。

## 最值得先看的五项

1. [官方 Privacy Policy](https://www.anthropic.com/legal/privacy)：确认公开披露了哪些数据类别。
2. [tengu-decoded 2.1.300 设备标识分析](https://github.com/wtfwhs/tengu-decoded/blob/main/versions/2.1.300/device-fingerprinting.md)：分辨随机安装 ID、硬件信息、设备密钥及发送路径。
3. [三端遥测与隐私审计](https://github.com/B1anYu/claude-telemetry-audit)：比较 Desktop、Code、Android，而不是混称“Claude 指纹”。
4. [自定义 API 路由标记文章](https://www.vincentschmalbach.com/claude-code-china-router-fingerprint/)：了解请求上下文中的另一类指纹。
5. [官方 Legal and compliance](https://code.claude.com/docs/en/legal-and-compliance)：区分客户端识别问题与订阅 OAuth 使用限制。

## 可以得出与不能得出的结论

- **官方确认的是数据类别与政策边界。** 隐私政策列出设备、连接、IP、时区、各类标识符与使用信息，并列出反欺诈/滥用用途；未公布“哪个字段多少分就封号”。
- **“指纹”至少包含四类。** 浏览器环境特征；TLS/HTTP 协议特征；安装/设备/账号标识；请求头、提示词、工具定义或路由标记。它们的采集位置和作用不同。
- **第三方逆向比纯经验文章更接近可检查的证据，但仍有版本和平台边界。** 本轮读取研究正文，没有独立重跑客户端提取和抓包，也看不到服务端如何使用这些信号。
- **不能把 device_id 一概认作硬件 ID。** 两份带版本说明的研究把多个桌面/CLI 标识追踪到随机值；另有 Windows 项目声称是 MachineGuid 哈希，存在冲突。需要对同一版本、同一平台、同一字段追踪，不能凭字段名作结论。
- **Android 需要单独研究。** 三端审计声称特定 Android 版本向 Sift 发送 ANDROID_ID 与网卡信息；这是研究者发现，尚未由本轮独立复验。ANDROID_ID 也不是永不变化的硬件序列号。
- **第三方“封号风险分”不等于 Anthropic 的内部判定。** 测出中文字体、时区或 IP 属性，只证明能观察到这些值；权重、阈值和封号概率需要额外证据。
- **本地清理不等于服务端解封；关闭运行遥测不等于匿名使用。** 推理请求、认证、账号和服务器端记录仍是独立数据路径。
- **报错应分层。** Cloudflare 验证/403、API 429、支付失败、组织禁用、账号暂停，不是同一个事件。Issue #8327 就展示了凭据选择造成“组织被禁用”报错的另一条线索。

## 阅读时最需要排除的误导

| 常见断言 | 本轮判断 |
|---|---|
| “中文字体/中文系统必封” | 未找到官方或可复现实验证据支持该绝对结论。 |
| “JA4 是唯一设备指纹” | 不正确；多名用户/设备可共享 TLS 客户端特征。 |
| “住宅 IP / VPS 可以保证不封” | 不成立；社区指南彼此矛盾，且支持地区资格不只由出口 IP 决定。 |
| “养号 24 小时、每分钟低于 10 次就安全” | 属于未充分论证的经验阈值，非公开官方规则。 |
| “改系统 MachineGuid 就解决” | 字段来源论证有冲突；也不能改变服务端账户处置。 |
| “检测站得分下降就是防封成功” | 只说明满足该检测站规则，没有证明与实际风控一致。 |
| “GitHub 官方仓库里的 issue 就是官方承认” | Issue 多为用户报告，须看作者、维护者回应与可复现证据。 |

以下“已读”指读到了页面或 README 的相关正文，不表示已审计全部代码、全部评论或实测效果。项目介绍中的防封承诺不代表本报告认可。

'''
lines=[intro]
idx=0
for name,rows in groups:
    lines += ['## '+name+'\n','| 编号 | 资料 | 内容与证据边界 |\n|---|---|---|\n']
    for r in rows:
        idx+=1;r['id']=idx;r['category']=name
        lines.append(f"| {idx} | [{r['title']}]({r['url']}) | {r['note']} |\n")
    lines.append('\n')
lines.append('''## 后续检索词

中文：`Claude 封号 指纹`、`Claude Code 遥测 device_id`、`Claude 浏览器指纹`、`Claude MachineGuid`、`Claude ant-did`、`Claude TLS JA4`、`Claude 中转 日期 撇号`、`Claude Android Sift`。

英文：`Claude device fingerprint account suspended`、`Claude Code telemetry userID machineID`、`Claude Code metadata.user_id`、`Claude Code apiBaseUrlHost`、`Claude router fingerprint`、`Claude OAuth third-party client ban`。

检索时排除“模型真伪指纹”“AI 文风识别”以及只是由 Claude 驱动的其他网站自动化项目：这些经常命中 fingerprint，但不是 Claude 账号风控研究。

## 本地资料说明

- `claude-ban-fingerprint-sources.json`：本报告的结构化目录。
- `sources/`：本轮抓取的文本快照及 URL、HTTP 状态；包括验证失败页面，不把 HTTP 200 自动视为正文成功。
- `github_log.jsonl`：GitHub 检索候选及元数据，包含无关结果，未全部纳入目录。
- `search_log.jsonl`：通用搜索引擎尝试，相关性不足，仅保留过程记录。

来源会更新或失效；本轮保留的是检索时的文本。没有安装或运行上述检测、清理、代理或扩展项目。
''')
(ROOT/'claude-ban-fingerprint-research.md').write_text(''.join(lines),encoding='utf8')
(ROOT/'claude-ban-fingerprint-sources.json').write_text(json.dumps([r for _,rows in groups for r in rows],ensure_ascii=False,indent=2),encoding='utf8')
print(json.dumps({'entries':idx,'categories':len(groups),'report':str(ROOT/'claude-ban-fingerprint-research.md')},ensure_ascii=False))
