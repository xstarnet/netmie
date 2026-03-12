#!/bin/bash

# 多平台编译脚本
# 支持常见的 OS + Arch 组合，产物后缀带 os+arch
# 用法: ./build.sh <os> <arch>

set -e

# 版本信息
VERSION=${VERSION:-"dev"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# 输出目录
OUTPUT_DIR=${OUTPUT_DIR:-"./dist"}
mkdir -p "$OUTPUT_DIR"

# 定义支持的平台
# 格式: OS:ARCH
SUPPORTED_TARGETS=(
    "linux:amd64"
    "linux:arm64"
    "linux:arm"
    "linux:386"
    "linux:mips"
    "linux:mipsle"
    "linux:mips64"
    "linux:mips64le"
    "linux:ppc64le"
    "linux:riscv64"
    "linux:s390x"
    "darwin:amd64"
    "darwin:arm64"
    "windows:amd64"
    "windows:arm64"
    "windows:386"
    "freebsd:amd64"
    "freebsd:arm64"
    "openbsd:amd64"
    "openbsd:arm64"
    "netbsd:amd64"
    "dragonfly:amd64"
    "illumos:amd64"
)

# 检查平台是否支持
is_supported() {
    local goos=$1
    local goarch=$2

    for target in "${SUPPORTED_TARGETS[@]}"; do
        if [ "$target" = "${goos}:${goarch}" ]; then
            return 0
        fi
    done
    return 1
}

# 编译组件
build() {
    local goos=$1
    local goarch=$2

    local output_name="netbird-${goos}-${goarch}"
    if [ "$goos" = "windows" ]; then
        output_name="${output_name}.exe"
    fi

    local output_path="$OUTPUT_DIR/$output_name"

    echo "Building: netbird for $goos/$goarch"

    local ldflags="-s -w"
    ldflags="$ldflags -X github.com/netbirdio/netbird/version.version=$VERSION"
    ldflags="$ldflags -X main.commit=$COMMIT"
    ldflags="$ldflags -X main.date=$DATE"
    ldflags="$ldflags -X main.builtBy=build-script"

    local env_vars=(
        "GOOS=$goos"
        "GOARCH=$goarch"
        "CGO_ENABLED=0"
    )

    # 特殊处理 MIPS 的浮点设置
    if [[ "$goarch" == mips* ]]; then
        env_vars+=("GOMIPS=hardfloat")
    fi

    env "${env_vars[@]}" go build \
        -ldflags "$ldflags" \
        -tags load_wgnt_from_rsrc \
        -o "$output_path" \
        ./client

    echo "✓ Built: $output_path"
}

# 显示帮助
usage() {
    cat << EOF
Usage: $0 [options] <os> <arch>

Options:
    -v VERSION  Set version (default: dev)
    -o DIR      Set output directory (default: ./dist)
    -h          Show this help message

Examples:
    $0 linux amd64          # Build for linux/amd64
    $0 darwin arm64         # Build for darwin/arm64 (Apple Silicon)
    $0 windows amd64        # Build for windows/amd64
    VERSION=v1.0.0 $0 linux amd64
    $0 -o ./build linux amd64

Supported Platforms:
    linux:    amd64, arm64, arm, 386, mips, mipsle, mips64, mips64le, ppc64le, riscv64, s390x
    darwin:   amd64, arm64
    windows:  amd64, arm64, 386
    freebsd:  amd64, arm64
    openbsd:  amd64, arm64
    netbsd:   amd64
    dragonfly: amd64
    illumos:  amd64
EOF
}

# 清理函数
clean() {
    echo "Cleaning build artifacts..."
    rm -rf "$OUTPUT_DIR"
    echo "Cleaned: $OUTPUT_DIR"
}

# 解析参数
while getopts "v:o:h" opt; do
    case $opt in
        v)
            VERSION="$OPTARG"
            ;;
        o)
            OUTPUT_DIR="$OPTARG"
            mkdir -p "$OUTPUT_DIR"
            ;;
        h)
            usage
            exit 0
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done

shift $((OPTIND-1))

# 处理命令
case "${1:-}" in
    clean)
        clean
        exit 0
        ;;
    help|-h|--help)
        usage
        exit 0
        ;;
esac

# 检查参数
if [ $# -lt 2 ]; then
    echo "Error: Missing arguments"
    echo ""
    usage
    exit 1
fi

GOOS=$1
GOARCH=$2

# 检查平台是否支持
if ! is_supported "$GOOS" "$GOARCH"; then
    echo "Error: Unsupported platform '$GOOS/$GOARCH'"
    echo ""
    echo "Run '$0 -h' to see supported platforms"
    exit 1
fi

# 执行编译
build "$GOOS" "$GOARCH"
