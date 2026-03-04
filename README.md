# Netmie

Netmie 是基于 NetBird 的 P2P VPN 解决方案，集成了 V2Ray 代理功能。

## 功能特性

- **NetBird P2P VPN**: 基于 WireGuard 的点对点 VPN 连接
- **V2Ray 代理**: 支持 VMess/VLESS 等协议的代理功能
- **双模式共存**: PC 端支持 NetBird 和 V2Ray 同时运行
- **统一管理**: 通过 daemon 统一管理两种连接方式

## 快速开始

### 开发环境安装

```bash
# 构建并安装
./dev-install.sh

# 或手动安装
go build -o ./netmie ./client
sudo cp ./netmie /usr/local/bin/
sudo netmie service install
sudo netmie service start
```

### NetBird 功能

```bash
# 连接到 NetBird 网络
netmie up

# 查看连接状态
netmie status

# 断开连接
netmie down

# 查看日志
netmie debug log
```

### V2Ray 功能

```bash
# 1. 配置 V2Ray（首次使用或更新配置时）
netmie vconfig /path/to/v2ray-config.json

# 2. 启动 V2Ray 代理
netmie vup

# 3. 查看 V2Ray 状态
netmie vstatus

# 4. 停止 V2Ray 代理
netmie vdown
```

### V2Ray 配置示例

```json
{
  "log": {
    "loglevel": "warning"
  },
  "inbounds": [
    {
      "port": 10808,
      "protocol": "socks",
      "settings": {
        "udp": true
      }
    },
    {
      "port": 10809,
      "protocol": "http"
    }
  ],
  "outbounds": [
    {
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": "your-server.com",
            "port": 443,
            "users": [
              {
                "id": "your-uuid",
                "alterId": 0,
                "security": "auto"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "ws",
        "security": "tls",
        "wsSettings": {
          "path": "/path"
        }
      }
    }
  ]
}
```

配置完成后，代理将在以下端口可用：
- SOCKS5: `127.0.0.1:10808`
- HTTP: `127.0.0.1:10809`

## 服务管理

```bash
# 安装服务
sudo netmie service install

# 启动服务
sudo netmie service start

# 停止服务
sudo netmie service stop

# 卸载服务
sudo netmie service uninstall

# 查看服务状态
sudo systemctl status netbird
```

## 开发相关

### 构建

```bash
# 构建客户端
go build -o ./netmie ./client

# 构建并重启 daemon（开发时使用）
./restart-daemon.sh
```

### 清理环境

```bash
# 清理所有进程和残留文件
./cleanup.sh
```

### 目录结构

```
netmie/
├── client/                    # 客户端代码
│   ├── cmd/                   # CLI 命令
│   │   ├── up.go             # NetBird 连接
│   │   ├── down.go           # NetBird 断开
│   │   ├── status.go         # NetBird 状态
│   │   ├── vup.go            # V2Ray 启动
│   │   ├── vdown.go          # V2Ray 停止
│   │   ├── vstatus.go        # V2Ray 状态
│   │   └── vconfig.go        # V2Ray 配置更新
│   ├── internal/
│   │   └── v2ray/            # V2Ray 模块
│   │       ├── engine.go     # V2Ray 引擎
│   │       ├── config.go     # 配置管理
│   │       └── xray_wrapper.go # v2ray-core 封装
│   ├── proto/                # gRPC 协议定义
│   └── server/               # Daemon 服务器
├── management/               # 管理服务（未修改）
└── release_files/           # 发布相关文件
```

## 技术架构

### PC 端架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Netmie Daemon                           │
│  ┌──────────────────┐         ┌──────────────────────────┐ │
│  │ NetBird Module   │         │ V2Ray Module             │ │
│  │ - wt0 interface  │         │ - v2ray-core             │ │
│  │ - WireGuard      │         │ - Local Proxy            │ │
│  │ - Route Table 1  │         │ - SOCKS5/HTTP            │ │
│  └──────────────────┘         └──────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                    ↑
                    │ gRPC (Unix Socket)
                    │
            ┌───────┴────────┐
            │  Netmie CLI    │
            └────────────────┘
```

### 通信方式

- **CLI ↔ Daemon**: gRPC over Unix Socket (`/var/run/netbird.sock`)
- **Daemon 权限**: Root 权限运行，可以读取任意路径的配置文件
- **配置管理**: CLI 发送绝对路径，Daemon 读取并验证配置

## 常见问题

### 1. Daemon 无法启动

```bash
# 检查 daemon 状态
sudo systemctl status netbird

# 查看日志
sudo journalctl -u netbird -f

# 重启 daemon
sudo systemctl restart netbird
```

### 2. V2Ray 启动失败

```bash
# 检查配置文件格式
cat ~/.netmie/v2ray-config.json | jq .

# 检查端口占用
netstat -tlnp | grep 10808

# 查看详细错误
netmie vstatus
```

### 3. 端口已被占用

```bash
# 查找占用端口的进程
sudo lsof -i :10808

# 停止 V2Ray
netmie vdown

# 或清理所有相关进程
./cleanup.sh
```

### 4. 配置更新不生效

```bash
# 重新配置
netmie vconfig /path/to/new-config.json

# 如果 V2Ray 正在运行，vconfig 会自动重启
# 如果未运行，需要手动启动
netmie vup
```

## 与 NetBird 的区别

Netmie 是 NetBird 的增强版本，主要变更：

1. **项目改名**: `netbird` → `netmie`
2. **新增 V2Ray 支持**: 集成 v2ray-core v5
3. **新增命令**: `vup`, `vdown`, `vstatus`, `vconfig`
4. **配置目录**: `~/.netmie/` (存储 V2Ray 配置)
5. **保持兼容**: 所有原有 NetBird 功能完全保留

## 依赖

- Go 1.25+
- v2ray-core v5.46.0
- Linux: iproute2, iptables/nftables
- macOS/Windows: 无额外依赖

## 许可证

基于 NetBird 项目，遵循 BSD-3-Clause 许可证。

## 致谢

本项目基于 [NetBird](https://github.com/netbirdio/netbird) 开发，感谢 NetBird 团队的优秀工作。
