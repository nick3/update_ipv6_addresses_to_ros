# RouterOS IPv6 地址列表更新工具

这是一个用于自动更新 MikroTik RouterOS 防火墙 IPv6 地址列表的工具。它可以获取指定网络接口的公网 IPv6 地址，并自动更新到 RouterOS 的防火墙地址列表中。

## 功能特点

- 自动获取指定网络接口的公网 IPv6 地址
- 自动过滤链路本地地址（fe80::/10）
- 支持通过配置文件进行设置
- 自动清理并更新 RouterOS 防火墙地址列表

## 配置说明

程序使用 JSON 格式的配置文件，默认配置文件名为 `config.json`。配置文件示例：

```json
{
    "networkInterface": "eth0",
    "routerIP": "192.168.1.1",
    "routerPort": 8728,
    "routerUsername": "admin",
    "routerPassword": "password",
    "addressListName": "ipv6-whitelist"
}
```

配置项说明：
- `networkInterface`: 要获取 IPv6 地址的网络接口名称
- `routerIP`: RouterOS 设备的 IP 地址
- `routerPort`: RouterOS API 端口（默认为 8728）
- `routerUsername`: RouterOS 登录用户名
- `routerPassword`: RouterOS 登录密码
- `addressListName`: 要更新的防火墙地址列表名称

## 使用方法

1. 编译程序：
```bash
go build -o update-ipv6-ros
```

2. 创建配置文件 `config.json`

3. 运行程序：
```bash
./update-ipv6-ros
```

可以通过 `-c` 参数指定配置文件路径：
```bash
./update-ipv6-ros -c /path/to/config.json
```

## 建议用途

- 配合定时任务自动更新动态 IPv6 地址
- 自动维护 RouterOS 防火墙白名单
- 远程访问管理时保持 IPv6 地址列表最新

## 注意事项

1. 确保网络接口有可用的公网 IPv6 地址
2. RouterOS 用户需要具有读写防火墙地址列表的权限
3. 建议将程序添加到计划任务中定期执行
