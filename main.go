package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/go-routeros/routeros"
)

// 配置参数
type Config struct {
    NetworkInterface string `json:"networkInterface"`
    RouterIP         string `json:"routerIP"`
    RouterPort       int    `json:"routerPort"`
    RouterUsername   string `json:"routerUsername"`
    RouterPassword   string `json:"routerPassword"`
    AddressListName  string `json:"addressListName"`
}

// 获取指定网络接口的所有公网 IPv6 地址
func getPublicIPv6Addresses(interfaceName string) ([]string, error) {
    iface, err := net.InterfaceByName(interfaceName)
    if err != nil {
        return nil, fmt.Errorf("无法找到网络接口 %s: %v", interfaceName, err)
    }

    addrs, err := iface.Addrs()
    if err != nil {
        return nil, fmt.Errorf("获取接口地址时出错: %v", err)
    }

    var ipv6Addresses []string
    for _, addr := range addrs {
        var ip net.IP
        switch v := addr.(type) {
        case *net.IPNet:
            ip = v.IP
        case *net.IPAddr:
            ip = v.IP
        }

        if ip == nil || ip.IsLoopback() {
            continue
        }

        if ip.To16() != nil && ip.To4() == nil {
            // 排除链路本地地址（fe80::/10）
            if !strings.HasPrefix(ip.String(), "fe80") {
                ipv6Addresses = append(ipv6Addresses, ip.String())
            }
        }
    }

    if len(ipv6Addresses) == 0 {
        return nil, errors.New("未找到公网 IPv6 地址")
    }

    return ipv6Addresses, nil
}

// 更新 RouterOS 的 IPv6 防火墙地址列表
func updateRouterOSIPv6Address(cfg Config, ipv6Addresses ...string) error {
    routerIP := fmt.Sprintf("%s:%d", cfg.RouterIP, cfg.RouterPort)
    routerUsername := cfg.RouterUsername
    routerPassword := cfg.RouterPassword
    addressListName := cfg.AddressListName

    // 建立与 RouterOS 的连接
    c, err := routeros.Dial(routerIP, routerUsername, routerPassword)
    if err != nil {
        return fmt.Errorf("连接到RouterOS时出错: %v", err)
    }
    defer c.Close()

    // 检查地址列表是否存在并删除旧条目
    listCheckCmds := []string{
        "/ipv6/firewall/address-list/print",
        fmt.Sprintf("?list=%s", addressListName),
    }

    response, err := c.RunArgs(listCheckCmds)
    if err != nil {
        return fmt.Errorf("查询地址列表时出错: %v", err)
    }

    if len(response.Re) > 0 {
        for _, re := range response.Re {
            id := re.Map[".id"]
            deleteCmds := []string{
                "/ipv6/firewall/address-list/remove",
                fmt.Sprintf("=.id=%s", id),
            }
            _, err = c.RunArgs(deleteCmds)
            if err != nil {
                return fmt.Errorf("删除旧地址时出错: %v", err)
            }
        }
        log.Printf("已删除地址列表 %s 中的旧条目\n", addressListName)
    }

    // 添加所有新的 IPv6 地址到地址列表
    for _, ipv6Address := range ipv6Addresses {
        addCmds := []string{
            "/ipv6/firewall/address-list/add",
            fmt.Sprintf("=address=%s", ipv6Address),
            fmt.Sprintf("=list=%s", addressListName),
        }

        _, err = c.RunArgs(addCmds)
        if err != nil {
            return fmt.Errorf("添加IPv6地址 %s 到地址列表时出错: %v", ipv6Address, err)
        }
        log.Printf("已将IPv6地址 %s 添加到地址列表 %s\n", ipv6Address, addressListName)
    }

    return nil
}

func main() {
    // 定义命令行参数
    configPath := flag.String("c", "config.json", "配置文件路径")
    flag.Parse()

    file, err := os.Open(*configPath)
    if err != nil {
        log.Fatalf("无法打开配置文件 %s: %v", *configPath, err)
    }
    defer file.Close()

    var cfg Config
    if err := json.NewDecoder(file).Decode(&cfg); err != nil {
        log.Fatalf("解析配置文件失败: %v", err)
    }

    // 获取所有公网 IPv6 地址
    ipv6Addresses, err := getPublicIPv6Addresses(cfg.NetworkInterface)
    if err != nil {
        log.Fatalf("获取公网IPv6地址失败: %v", err)
    }

    fmt.Printf("找到以下公网 IPv6 地址:\n")
    for _, addr := range ipv6Addresses {
        fmt.Printf("- %s\n", addr)
    }

    // 更新 RouterOS 的 IPv6 防火墙地址列表
    err = updateRouterOSIPv6Address(cfg, ipv6Addresses...)
    if err != nil {
        log.Fatalf("更新RouterOS时出错: %v", err)
    }

    os.Exit(0)
}
