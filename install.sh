#!/bin/bash
# Netmie 安装脚本
# 自动检测 OS 和 Arch，下载对应版本的二进制包并安装服务

set -e

# 配置
APP_NAME="netmie"
CLI_NAME="netmie"

# 下载地址配置（可自定义）
DOWNLOAD_BASE_URL=${DOWNLOAD_BASE_URL:-"https://www.miemie.tech/netmie/release"}

# 安装目录
INSTALL_DIR=${INSTALL_DIR:-"/usr/local/bin"}

# 检测 sudo
SUDO=""
if command -v sudo > /dev/null 2>&1 && [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
elif command -v doas > /dev/null 2>&1 && [ "$(id -u)" -ne 0 ]; then
    SUDO="doas"
fi

# 检测操作系统
detect_os() {
    local os_type=""

    case "$(uname -s)" in
        Linux*)
            os_type="linux"
            ;;
        Darwin*)
            os_type="darwin"
            ;;
        FreeBSD*)
            os_type="freebsd"
            ;;
        OpenBSD*)
            os_type="openbsd"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            os_type="windows"
            ;;
        *)
            echo "错误: 不支持的操作系统: $(uname -s)"
            exit 1
            ;;
    esac

    echo "$os_type"
}

# 检测架构
detect_arch() {
    local arch=""

    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        i386|i686|x86)
            arch="386"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        armv7l|armv6l|arm)
            arch="arm"
            ;;
        mips)
            arch="mips"
            ;;
        mipsel|mipsle)
            arch="mipsle"
            ;;
        mips64)
            arch="mips64"
            ;;
        mips64el|mips64le)
            arch="mips64le"
            ;;
        ppc64le)
            arch="ppc64le"
            ;;
        riscv64)
            arch="riscv64"
            ;;
        s390x)
            arch="s390x"
            ;;
        *)
            echo "错误: 不支持的架构: $(uname -m)"
            exit 1
            ;;
    esac

    echo "$arch"
}

# 下载二进制文件
download_binary() {
    local os_type=$1
    local arch=$2

    # 构建下载 URL
    # 格式: netbird-{os}-{arch}.gz
    local filename="netbird-${os_type}-${arch}.gz"
    local download_url="${DOWNLOAD_BASE_URL}/${filename}"

    local tmp_dir="/tmp/netmie-install-$$"
    mkdir -p "$tmp_dir"

    echo "下载 Netmie for ${os_type}/${arch}..."
    echo "URL: ${download_url}"

    # 下载文件
    local download_success=false
    if command -v curl > /dev/null 2>&1; then
        if curl -fsSL -o "${tmp_dir}/${filename}" "$download_url" 2>/dev/null; then
            download_success=true
        fi
    elif command -v wget > /dev/null 2>&1; then
        if wget -q -O "${tmp_dir}/${filename}" "$download_url" 2>/dev/null; then
            download_success=true
        fi
    else
        echo "错误: 需要 curl 或 wget 来下载文件"
        rm -rf "$tmp_dir"
        exit 1
    fi

    if [ "$download_success" = false ]; then
        echo "错误: 下载失败，无法从 ${download_url} 下载文件"
        echo ""
        echo "可能的原因："
        echo "  1. 该平台的二进制文件尚未发布"
        echo "  2. 下载地址配置错误"
        echo "  3. 网络连接问题"
        echo ""
        echo "支持的格式: netbird-{os}-{arch}.gz"
        echo "例如: netbird-linux-amd64.gz"
        rm -rf "$tmp_dir"
        exit 1
    fi

    echo "下载完成，正在解压..."

    # 解压 .gz 文件
    cd "$tmp_dir"
    if ! gunzip -f "$filename" 2>/dev/null; then
        echo "错误: 解压失败"
        rm -rf "$tmp_dir"
        exit 1
    fi

    # 解压后的文件名（去掉 .gz 后缀）
    local binary_name="${filename%.gz}"

    if [ ! -f "$binary_name" ]; then
        echo "错误: 解压后找不到文件: $binary_name"
        rm -rf "$tmp_dir"
        exit 1
    fi

    echo "找到二进制文件: $binary_name"

    # 安装二进制文件
    echo "安装二进制文件到 ${INSTALL_DIR}/${CLI_NAME}..."
    ${SUDO} mkdir -p "$INSTALL_DIR"
    ${SUDO} cp "$binary_name" "${INSTALL_DIR}/${CLI_NAME}"
    ${SUDO} chmod +x "${INSTALL_DIR}/${CLI_NAME}"

    # 清理临时文件
    cd /
    rm -rf "$tmp_dir"

    echo "二进制文件安装完成"
}

# 安装服务
install_service() {
    echo ""
    echo "正在安装 Netmie 服务..."

    if ! command -v "${INSTALL_DIR}/${CLI_NAME}" > /dev/null 2>&1; then
        echo "错误: 找不到 ${CLI_NAME} 命令"
        exit 1
    fi

    # 检查服务是否已存在，如果存在则先卸载
    local service_status
    service_status=$(${SUDO} "${INSTALL_DIR}/${CLI_NAME}" service status 2>&1 || true)
    if echo "$service_status" | grep -q "Running\|Stopped"; then
        echo "检测到已存在的服务，正在卸载旧服务..."
        ${SUDO} "${INSTALL_DIR}/${CLI_NAME}" service stop 2>/dev/null || true
        ${SUDO} "${INSTALL_DIR}/${CLI_NAME}" service uninstall 2>/dev/null || true
        echo "旧服务已卸载"
    fi

    # 安装服务
    echo "正在安装新服务..."
    if ! ${SUDO} "${INSTALL_DIR}/${CLI_NAME}" service install 2>&1; then
        echo "错误: 服务安装失败"
        exit 1
    fi

    # 启动服务
    echo "正在启动 Netmie 服务..."
    if ! ${SUDO} "${INSTALL_DIR}/${CLI_NAME}" service start 2>&1; then
        echo "警告: 服务启动可能失败"
    fi
}

# 显示帮助
show_help() {
    cat << EOF
Netmie 安装脚本

用法: $0 [选项]

选项:
    -u, --url URL          指定下载基础 URL
    -d, --dir DIR          指定安装目录 (默认: /usr/local/bin)
    -h, --help             显示此帮助信息
    --no-service           不安装和启动系统服务

环境变量:
    DOWNLOAD_BASE_URL      下载基础 URL
                           (默认: https://www.miemie.tech/netmie/release)
    INSTALL_DIR            安装目录 (默认: /usr/local/bin)

示例:
    $0                     # 使用默认配置安装
    $0 -u https://your-domain.com/releases  # 使用自定义下载地址
    $0 --no-service        # 只安装二进制文件，不安装服务

下载文件格式:
    {DOWNLOAD_BASE_URL}/netbird-{os}-{arch}.gz
    例如: https://www.miemie.tech/netmie/release/netbird-linux-amd64.gz
EOF
}

# 主函数
main() {
    local no_service=false

    # 解析参数
    while [ $# -gt 0 ]; do
        case "$1" in
            -u|--url)
                DOWNLOAD_BASE_URL="$2"
                shift 2
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --no-service)
                no_service=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done

    echo "====================================="
    echo "  Netmie 安装脚本"
    echo "====================================="
    echo ""

    # 检测系统信息
    echo "正在检测系统信息..."
    OS_TYPE=$(detect_os)
    ARCH=$(detect_arch)
    echo "操作系统: $OS_TYPE"
    echo "架构: $ARCH"
    echo "下载地址: ${DOWNLOAD_BASE_URL}"
    echo "安装目录: $INSTALL_DIR"
    echo ""

    # 检查是否已安装
    if command -v "${INSTALL_DIR}/${CLI_NAME}" > /dev/null 2>&1; then
        echo "检测到已安装的 Netmie"
        local current_version
        current_version=$("${INSTALL_DIR}/${CLI_NAME}" version 2>/dev/null || echo "unknown")
        echo "当前版本: $current_version"
        echo ""
        read -p "是否重新安装? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "安装已取消"
            exit 0
        fi
        echo ""
    fi

    # 下载并安装
    download_binary "$OS_TYPE" "$ARCH"

    # 安装服务
    if [ "$no_service" = false ]; then
        install_service
    fi

    echo ""
    echo "====================================="
    echo "  安装完成!"
    echo "====================================="
    echo ""
    echo "Netmie 已安装到: ${INSTALL_DIR}/${CLI_NAME}"
    echo ""
    echo "常用命令:"
    echo "  ${CLI_NAME} up              # 连接到 NetBird 网络"
    echo "  ${CLI_NAME} down            # 断开连接"
    echo "  ${CLI_NAME} status          # 查看状态"
    echo "  ${CLI_NAME} vconfig <file>  # 配置 V2Ray"
    echo "  ${CLI_NAME} vup             # 启动 V2Ray 代理"
    echo "  ${CLI_NAME} vstatus         # 查看 V2Ray 状态"
    echo "  ${CLI_NAME} service status  # 查看服务状态"
    echo ""
    echo "更多信息请查看文档或运行: ${CLI_NAME} --help"
}

# 运行主函数
main "$@"
