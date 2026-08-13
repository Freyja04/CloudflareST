# CloudflareST

## 使用提示

- 如果平均延迟非常低（如 `0.xx`），说明测速流量可能经过代理，请先关闭代理软件后再测速。
- 电脑开机后第一次测速的延迟可能明显偏高。建议正式测速前先随便测试几个 IP；只要进度条开始移动即可停止。
- HTTPing 本质上属于网络扫描行为。在服务器上运行时请降低并发（`-n`），避免触发服务商、运营商或 CDN 的临时限制。若首次可用 IP 正常、随后逐渐减少甚至归零，暂停一段时间后恢复，通常也应降低并发后重试。
- Cloudflare CDN 使用 Anycast IP。同一 IP 在不同地区、运营商和时间可能分配到不同节点与线路，测速结果会随网络环境变化。
- HTTPing 模式可从响应头识别 Cloudflare、AWS CloudFront、Fastly、Gcore、CDN77、Bunny 等 CDN 的地区码：Cloudflare、AWS CloudFront、Fastly 使用 IATA 三字码（如 `HKG`、`LAX`）；CDN77、Bunny 使用二字国家或区域码（如 `US`、`CN`）；Gcore 使用二字城市码（如 `FR`、`AM`）。
- 在 PowerShell 中使用 `-o ""` 时空参数可能被忽略，可改用 `-o " "`。
- 使用 `-debug` 可输出 HTTPing 与下载测速的中断原因，例如 HTTP 状态码不符、地址错误、超时、连接被阻断或 403。

HTTPing 指定地区示例：

```text
yx_windows_amd64.exe -httping -cfcolo HKG,KHH,NRT,LAX,SEA,SJC,FRA,MAD
```

## 上传优选 IP

程序第一次启动时会在 exe 同目录自动生成 `cfyx.json`。先填写要使用的上传方式对应字段，再重新测速；该文件含 Token，已被 Git 忽略，请勿上传或分享。

```json
{
  "cf_api_token": "",
  "cf_zone_id": "",
  "cf_base_domain": "",
  "cf_proxied": true,
  "github_token": "",
  "github_repo": "",
  "github_file_path": "cfyx.txt"
}
```

- 上传到 Cloudflare DNS：填写 `cf_api_token`、`cf_zone_id`、`cf_base_domain`。Token 需有该 Zone 的 DNS 编辑权限；例如基础域名为 `example.com`，选择的 IP 会写入或更新 `yx1.example.com`、`yx2.example.com` 等记录。IPv4 写 A 记录，IPv6 写 AAAA 记录，`cf_proxied` 控制是否开启代理。
- 上传到 GitHub 文件：填写 `github_token`、`github_repo`（格式 `用户名/仓库名`），可选 `github_file_path`（默认 `cfyx.txt`）。Token 需有目标仓库 Contents 读写权限；文件不存在时自动创建，已有纯 IP 行会按本次选择结果依次替换，其他内容保留。

测速有结果且显示结果数量 `-p` 大于 0 时，程序会询问是否上传：输入 `2` 选择 Cloudflare DNS，输入 `3` 选择 GitHub 文件，然后输入要上传的结果编号（如 `1,2,3`）；直接回车则不上传。

## 参数说明

```text
-n 200
    延迟测速线程；越多延迟测速越快，性能较弱的设备请勿设置过高。（默认 200，最多 1000；HTTPing 模式默认 100）
-t 4
    延迟测速次数；单个 IP 的延迟测速次数。（默认 4 次）
-dn 10
    下载测速数量；延迟测速排序后，从最低延迟开始进行下载测速的数量。（默认 10 个）
-dt 10
    下载测速时间；单个 IP 下载测速的最长时间。（默认 10 秒）
-tp 443
    指定测速端口；延迟测速和下载测速使用的端口。（默认 443）
-url https://speed.cloudflare.com/__down?bytes=99000000
    指定测速地址；HTTPing 延迟测速和下载测速使用的地址。

-httping
    将延迟测速模式切换为 HTTP 协议，使用 -url 指定的测速地址。（默认 TCPing）
-httping-code 200
    HTTPing 延迟测速时接受的单个 HTTP 状态码；未指定时接受 200、301、302。
-cfcolo HKG,KHH,NRT,LAX,SEA,SJC,FRA,MAD
    匹配指定地区；地区码以英文逗号分隔，大小写均可，仅 HTTPing 模式可用。（默认所有地区）

-tl 200
    平均延迟上限；只输出低于指定平均延迟的 IP。（默认 200 ms）
-tll 40
    平均延迟下限；只输出高于指定平均延迟的 IP。（默认 40 ms）
-tlr 0.2
    丢包率上限；只输出低于或等于指定丢包率的 IP，范围为 0.00 到 1.00。（默认 0.20）
-sl 5
    下载速度下限；只输出高于指定下载速度的 IP，达到 -dn 数量后停止测速。（默认 5.00 MB/s）

-p 10
    显示结果数量；测速后直接显示指定数量的结果，为 0 时不显示结果。（默认 10 个）
-f ip.txt
    IP 段数据文件；路径含空格时请加引号，支持其他 CDN 的 IP 段。（默认 ip.txt）
-ip 1.1.1.1,2.2.2.2/24,2606:4700::/32
    直接指定 IP 段数据，以英文逗号分隔。（默认空）
-o result.csv
    写入结果文件；路径含空格时请加引号，使用空值可不写入文件。（默认 result.csv）

-dd
    禁用下载测速；禁用后按延迟排序。（默认启用下载测速并按下载速度排序）
-allip
    测试 IP 段中的全部 IPv4 地址。（默认每个 /24 段随机测试一个 IP）

-debug
    调试输出模式；输出 HTTPing 和下载测速中的详细错误信息。（默认关闭）

-v
    打印程序版本并检查更新。
-h
    打印帮助说明。
```

## License

The GPL-3.0 License.
