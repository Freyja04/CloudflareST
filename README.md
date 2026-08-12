
如果平均延迟非常低（如 0.xx），则说明 CFST 测速时走了代理，请先关闭代理软件后再测速  
注意！我发现电脑开机后第一次测速延迟会明显偏高（手动 TCPing 也一样），后续测速都正常  
因此建议大家开机后第一次正式测速前，先随便测几个 IP（无需等待延迟测速完成，只要进度条动了就可以直接关了）  

## 参数：
    -n 200
        延迟测速线程；越多延迟测速越快，性能弱的设备 (如路由器) 请勿太高；(默认 200 最多 1000)
    -t 4
        延迟测速次数；单个 IP 延迟测速的次数；(默认 4 次)
    -dn 10
        下载测速数量；延迟测速并排序后，从最低延迟起下载测速的数量；(默认 10 个)
    -dt 10
        下载测速时间；单个 IP 下载测速最长时间，不能太短；(默认 10 秒)
    -tp 443
        指定测速端口；延迟测速/下载测速时使用的端口；(默认 443 端口)
    -url https://cf.xiu2.xyz/url
        指定测速地址；延迟测速(HTTPing)/下载测速时使用的地址，默认地址不保证可用性，建议自建；

    -httping
        切换测速模式；延迟测速模式改为 HTTP 协议，所用测试地址为 [-url] 参数；(默认 TCPing)
        当使用 HTTP 测速模式时，软件会从 HTTP 响应头中获取该 IP 当前地区码（支持 Cloudflare、AWS CloudFront、Fastly、Gcore、CDN77、Bunny 等 CDN）并显示出来。
        注意：HTTPing 本质上也算一种 网络扫描 行为，因此如果你在服务器上面运行，需要降低并发(-n)，否则可能会被一些严格的商家暂停服务。
        如果你遇到 HTTPing 首次测速可用 IP 数量正常，后续测速越来越少甚至直接为 0，但停一段时间后又恢复了的情况，那么也可能是被 运营商、Cloudflare CDN 认为你在网络扫描而 触发临时限制机制，因此才会过一会儿就恢复了，建议降低并发(-n)减少这种情况的发生。
    -httping-code 200
        有效状态代码；HTTPing 延迟测速时网页返回的有效 HTTP 状态码，仅限一个；(默认 200 301 302)
    -cfcolo HKG,KHH,NRT,LAX,SEA,SJC,FRA,MAD
        匹配指定地区；IATA 机场地区码或国家/城市码，英文逗号分隔，大小写均可，仅 HTTPing 模式可用；(默认 所有地区)
        支持 Cloudflare、AWS CloudFront、Fastly、Gcore、CDN77、Bunny 等 CDN
        其中 Cloudflare、AWS CloudFront、Fastly 使用的是 IATA 三字机场地区码，如：HKG,LAX
        其中 CDN77、Bunny 使用的是 二字国家/区域码，如：US,CN
        其中 Gcore 使用的是 二字城市码，如：FR,AM
        因此大家使用 -cfcolo 指定地区码时要根据不同的 CDN 来指定不同类型的地区码。

    -tl 200
        平均延迟上限；只输出低于指定平均延迟的 IP，各上下限条件可搭配使用；(默认 9999 ms)
    -tll 40
        平均延迟下限；只输出高于指定平均延迟的 IP；(默认 0 ms)
    -tlr 0.2
        丢包几率上限；只输出低于/等于指定丢包率的 IP，范围 0.00~1.00，0 过滤掉任何丢包的 IP；(默认 1.00)
    -sl 5
        下载速度下限；只输出高于指定下载速度的 IP，凑够指定数量 [-dn] 才会停止测速；(默认 0.00 MB/s)

    -p 10
        显示结果数量；测速后直接显示指定数量的结果，为 0 时不显示结果直接退出；(默认 10 个)
    -f ip.txt
        IP段数据文件；如路径含有空格请加上引号；支持其他 CDN IP段；(默认 ip.txt)
    -ip 1.1.1.1,2.2.2.2/24,2606:4700::/32
        指定IP段数据；直接通过参数指定要测速的 IP 段数据，英文逗号分隔；(默认 空)
    -o result.csv
        写入结果文件；如路径含有空格请加上引号；值为空时不写入文件 [-o ""]；(默认 result.csv)
        注意：在一些环境下使用 -o "" 可能会被忽略掉这个空参数导致报错，可加个空格 -o " " 解决

    -dd
        禁用下载测速；禁用后测速结果会按延迟排序 (默认按下载速度排序)；(默认 启用)
    -allip
        测速全部的IP；对 IP 段中的每个 IP (仅支持 IPv4) 进行测速；(默认 每个 /24 段随机测速一个 IP)

    -debug
        调试输出模式；会在一些非预期情况下输出更多日志以便判断原因；(默认 关闭)
        目前该功能仅针对 HTTPing 延迟测速过程 及 下载测速过程，当过程中因为各种原因导致当前 IP 测速中断都会输出错误原因
        例如：HTTPing 延迟测速过程中，因为 HTTP 状态码不符合或测速地址有问题或超时等原因而终止测速
        例如：下载测速过程中，因为下载测速地址有问题（被阻断、403状态码、超时）等原因而终止测速（导致显示 0.00）

    -v
        打印程序版本 + 检查版本更新
    -h
        打印帮助说明
```


建议进入 CFST 程序目录下，以**相对路径**方式运行：

**方式 一**：
1. 打开 CFST 程序所在目录  
2. 空白处按下 <kbd>Shift + 鼠标右键</kbd> 显示右键菜单  
3. 选择 **\[在此处打开命令窗口\]** 来打开 CMD 窗口，此时默认就位于当前目录下  
4. 输入带参数的命令，如：`cfst -tl 200 -dn 20` 即可运行

**方式 二**：
1. 打开 CFST 程序所在目录  
2. 直接在文件夹地址栏中全选(或清空)并输入 `cmd` 回车就能打开 CMD 窗口，此时默认就位于当前目录下  
4. 输入带参数的命令，如：`cfst -tl 200 -dn 20` 即可运行



> **提示**：如果用的是 **PowerShell** 只需把命令中的 `cfst` 改为 `.\cfst` 即可。  
> **注意**：在 **PowerShell** 下使用 `-o ""` 会被忽略掉空参数导致报错，可加个空格 `-o " "` 解决


> 注意：HTTPing 本质上也算一种**网络扫描**行为，因此如果你在服务器上面运行，需要**降低并发**(`-n`)，否则可能会被一些严格的商家暂停服务。如果你遇到 HTTPing 首次测速可用 IP 数量正常，后续测速越来越少甚至直接为 0，但停一段时间后又恢复了的情况，那么也可能是被 运营商、Cloudflare CDN 认为你在网络扫描而**触发临时限制机制**，因此才会过一会儿就恢复了，建议**降低并发**(`-n`)减少这种情况的发生。

Cloudflare CDN 的节点 IP 是 Anycast IP，即每个 IP 对应的服务器节点及地区不是固定的，而是动态变化的，**不同地区、不同运营商、不同时间段访问同一个IP分配到的服务器节点地区和路线也都是不一样的

Cloudflare、AWS CloudFront、Fastly** 都使用的是 **`IATA 三字机场地区码`**，如：HKG,LAX  
cfst -httping -cfcolo HKG,KHH,NRT,LAX,SEA,SJC,FRA,MAD

> **`IATA 三字机场地区码`**，可见：https://www.cloudflarestatus.com/  

****

****

为了方便，我是在编译的时候将版本号写入代码中的 version 变量，因此你手动编译时，需要像下面这样在 `go build` 命令后面加上 `-ldflags` 参数来指定版本号。本项目只打包 **Windows x86_64 / amd64** 版本，产物为 `Releases\cf_windows_amd64.zip`：

```bat
set version=v1.0.0
mkdir Releases\cf_windows_amd64
copy ip.txt Releases\cf_windows_amd64\ >nul
copy ipv6.txt Releases\cf_windows_amd64\ >nul
go build -o Releases\cf_windows_amd64\cf_windows_amd64.exe -ldflags "-s -w -X main.version=%version%"
tar -a -c -f Releases\cf_windows_amd64.zip -C Releases\cf_windows_amd64 .
```

</details>

****

## License

The GPL-3.0 License.
