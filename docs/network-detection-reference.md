# 网络检测实现参考

参考日期：2026-10-24。用户要求参考 FlClash。当前本机安装版本为 0.9.1+2026101702；本次阅读的公开源码固定在提交 `8d4f9744948751f1273fe871e1d05276fac09b0e`，不假定与本机二进制实现完全相同。

- [出口源列表及解析入口](https://github.com/chen08209/FlClash/blob/8d4f9744948751f1273fe871e1d05276fac09b0e/lib/common/service_probe.dart)：包含 Cloudflare、ip.sb、ipinfo、ipwho.is、GeoJS 等多个来源。
- [出口探测调度](https://github.com/chen08209/FlClash/blob/8d4f9744948751f1273fe871e1d05276fac09b0e/core/outbound_ip.go)：限制并发、错峰探测、成功后取消无用请求，过滤无效 IP 响应。
- [路由与缓存状态](https://github.com/chen08209/FlClash/blob/8d4f9744948751f1273fe871e1d05276fac09b0e/lib/providers/routed_probe.dart)：跟踪检测状态、缓存与路由变更。
- [代理选择](https://github.com/chen08209/FlClash/blob/8d4f9744948751f1273fe871e1d05276fac09b0e/lib/common/http.dart)：FlClash 可通过自身内核及 mixed-port 获取指定出口。

ClaudePro 借鉴设计思路并独立实现，不复制 FlClash 源码或引入其内核。默认顺序为 GeoJS、ipwho.is、ipinfo、ipapi；并发最多 2，间隔 800ms，失败立即尝试下一源，获得两份有效结果后取消其余请求。整体 10 秒、单源 4 秒，所有连接使用 HTTPS 和正常证书校验。

两份结果用于判断是否可以推荐同一地区模板，而不是对服务进行“投票认定真实位置”。出口 IP、国家或地区模板判断不一致时显示冲突、停止推荐；城市不同但 IANA 时区都映射到同一模板时允许推荐。只有一个可用来源时明确标为未交叉确认。此前同一 IP 的分歧在当前会话保留，不能因一个源暂时失效而消失。

429 尊重 Retry-After，缺失或无效时默认冷却该源 5 分钟。成功缓存 60 秒，失败缓存 30 秒，缓存和限流状态仅存内存，不后台联网。ASN 和运营商缺失显示未知，不影响有效 IP 与国家结果。详细失败原因仅返回固定说明和 HTTP 状态码，不暴露代理凭据。

ClaudePro 继续使用 Go 环境变量代理规则和现有系统网络路径（包括已有 TUN），不读取 FlClash 的订阅、控制接口密钥或修改节点、规则、代理设置；不能声称等同于 FlClash 按指定节点发起的内核检测。用户切换代理节点后，应等待短时缓存到期再手动检测，界面显示结果时间和来源。
