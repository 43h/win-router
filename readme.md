# WIN-ROUTER
---
更新于 2024-04-27

```txt
        PC(ROUTER)
          |    |
PC1<----->LAN--WAN<----->INTERNET
```

## 原理
PCAP监听流量，转发

## 编译环境
go version go1.21.0 windows/amd64  
Windows 11 专业版  23H2 22631.2861

## 已完成
* 转发TCP流量
* ICMP流量转发
* 获取windows网卡，IP/掩码/网卡名


## 未完成
* 支持arp
* 更新UDP校验和
* 支持源端口NAT
* lan口源与目的IP过滤
* 支持多wan口和lan口
* lan口支持多台PC
* 编写界面

## 注意
windows下本地回环口不能发包

## history
* 2024-07-06 修改转发表生命周期
* 2024-04-27 修改流匹配，支持多个源IP
* 2024-03-02 重构版本开始
* 2023-12-31 完成单lan口到单wan口转发demo
* 2023-12-29启动