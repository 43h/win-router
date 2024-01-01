# WIN-ROUTER
---
更新于 2023-12-29

PC<----->LAN|WAN<----->INTERNET

## 原理
监听两个网口，流量互转

## 编译环境
go version go1.21.0 windows/amd64  
Windows 11 专业版  23H2 22631.2861


## 局限
目前lan侧仅支持一台PC  
wan口侧需要手动配置下一条mac地址


## 已完成
* 获取windows网卡，IP/掩码/网卡名
* 获取windows路由信息
* 二层转发，修改转发报文IP

## 未完成
* 支持arp
* udp checksum
* 网口流量监听，兴趣流过滤
* 编写界面
* 支持多wan口和lan口
* lan口支持多台PC

## 注意
windows下本地回环口不能发包

## history
* 2023-12-31 完成单lan口到单wan口转发demo
* 2023-12-29启动