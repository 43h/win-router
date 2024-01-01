# WIN-ROUTERR
---
更新于 2023-12-29

PC<----->LAN|WAN<----->INTERNET


## target
* 路由功能

## Todo
* [x]获取windows网卡，IP/掩码/网卡名
* [x]获取windows路由信息
* [ ]支持arp
* [ ]网口流量监听，兴趣流过滤
* [ ]编写界面
* [ ]标注wan口和lan口
* [ ]实现NAT转发
* [ ]udp checksum


## 注意
windows下本地回环口不能发包

## history
* 2023-12-31 完成单lan口到单wan口转发demo
* 2023-12-29启动