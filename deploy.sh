#!/bin/bash

# 部署脚本：压缩 dist 目录，同步到远端，并将每个文件单独压缩存储
# 用法: ./deploy.sh [选项] [远端主机] [远端路径]
# 示例: ./deploy.sh user@example.com /var/www/app
#
# 远端最终结构：每个原文件都会被压缩为 .gz 格式
# 例如：dist/js/app.js -> /var/www/app/js/app.js.gz

set -e

# 配置
DIST_DIR="${DIST_DIR:-./dist}"
ARCHIVE_NAME="${ARCHIVE_NAME:-dist.tar.gz}"
SSH_PORT="${SSH_PORT:-22}"

# 模式
DEPLOY_MODE="full"  # full, install-only

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助
show_help() {
    cat << EOF
部署脚本

用法: $0 [选项] [远端主机] [远端路径]

选项:
    -i, --install-only     只部署 install.sh，不部署 dist 目录
    -h, --help             显示此帮助信息

环境变量:
    DIST_DIR               要压缩的目录 (默认: ./dist)
    ARCHIVE_NAME           压缩包名称 (默认: dist.tar.gz)
    SSH_PORT               SSH 端口 (默认: 22)

示例:
    # 完整部署（dist + install.sh）
    $0 user@example.com /var/www/app

    # 只部署 install.sh
    $0 -i user@example.com /var/www/app

    # 只部署 install.sh（使用环境变量）
    DEPLOY_MODE=install-only $0 user@example.com /var/www/app
EOF
}

# 解析参数
REMOTE_HOST=""
REMOTE_PATH=""

while [ $# -gt 0 ]; do
    case "$1" in
        -i|--install-only)
            DEPLOY_MODE="install-only"
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        -*)
            error "未知选项: $1"
            show_help
            exit 1
            ;;
        *)
            if [ -z "$REMOTE_HOST" ]; then
                REMOTE_HOST="$1"
            elif [ -z "$REMOTE_PATH" ]; then
                REMOTE_PATH="$1"
            else
                error "多余的参数: $1"
                exit 1
            fi
            shift
            ;;
    esac
done

# 检查参数
if [ -z "$REMOTE_HOST" ] || [ -z "$REMOTE_PATH" ]; then
    error "缺少必要参数"
    show_help
    exit 1
fi

info "开始部署流程..."
info "部署模式: $DEPLOY_MODE"
info "远端主机: $REMOTE_HOST"
info "远端路径: $REMOTE_PATH"

# 部署 install.sh 的函数
deploy_install_script() {
    info "部署 install.sh..."
    local install_script="./install.sh"
    if [ -f "$install_script" ]; then
        info "找到 install.sh，开始部署..."
        # 直接上传 install.sh 到远端（不压缩）
        rsync -avz --progress -e "ssh -p $SSH_PORT" "$install_script" "$REMOTE_HOST:$REMOTE_PATH/install.sh"

        # 设置权限
        ssh -p "$SSH_PORT" "$REMOTE_HOST" "sudo chmod 644 $REMOTE_PATH/install.sh"

        info "install.sh 已部署到: $REMOTE_PATH/install.sh"
    else
        warn "未找到 install.sh，跳过部署"
        return 1
    fi
}

# 部署 dist 目录的函数
deploy_dist() {
    # 检查 dist 目录是否存在
    if [ ! -d "$DIST_DIR" ]; then
        error "目录不存在: $DIST_DIR"
        exit 1
    fi

    info "源目录: $DIST_DIR"

    # 步骤 1: 压缩 dist 目录
    info "步骤 1/3: 压缩 $DIST_DIR 目录..."
    if [ -f "$ARCHIVE_NAME" ]; then
        warn "删除旧的压缩包: $ARCHIVE_NAME"
        rm -f "$ARCHIVE_NAME"
    fi

    tar -czf "$ARCHIVE_NAME" -C "$DIST_DIR" .
    info "压缩完成: $ARCHIVE_NAME ($(du -h "$ARCHIVE_NAME" | cut -f1))"

    # 步骤 2: 使用 rsync 同步到远端
    info "步骤 2/3: 同步到远端..."
    rsync -avz --progress -e "ssh -p $SSH_PORT" "$ARCHIVE_NAME" "$REMOTE_HOST:/tmp/$ARCHIVE_NAME"
    info "同步完成"

    # 步骤 3: 在远端解压缩并压缩每个文件
    info "步骤 3/3: 在远端解压并压缩每个文件..."
    ssh -p "$SSH_PORT" "$REMOTE_HOST" << EOF
    set -e
    echo "创建临时解压目录..."
    TEMP_DIR="/tmp/deploy_\$(date +%s)"
    mkdir -p "\$TEMP_DIR"

    echo "解压文件到临时目录..."
    tar -xzf "/tmp/$ARCHIVE_NAME" -C "\$TEMP_DIR"
    rm -f "/tmp/$ARCHIVE_NAME"

    echo "备份现有文件 (如果有)..."
    if [ -d "$REMOTE_PATH" ] && [ "\$(ls -A $REMOTE_PATH 2>/dev/null)" ]; then
        BACKUP_DIR="${REMOTE_PATH}_backup_\$(date +%Y%m%d_%H%M%S)"
        sudo mv "$REMOTE_PATH" "\$BACKUP_DIR"
        echo "已备份到: \$BACKUP_DIR"
    fi

    echo "创建目标目录并压缩每个文件..."
    sudo mkdir -p "$REMOTE_PATH"

    # 遍历所有文件，单独压缩每个文件
    find "\$TEMP_DIR" -type f | while read file; do
        # 计算相对路径
        rel_path="\${file#\$TEMP_DIR/}"
        target_dir="$REMOTE_PATH/\$(dirname "\$rel_path")"
        target_file="$REMOTE_PATH/\$rel_path"

        # 创建目标子目录
        sudo mkdir -p "\$target_dir"

        # 使用 gzip 压缩单个文件（保留原文件权限）
        sudo gzip -c "\$file" > "\$target_file.gz"

        # 复制原文件权限
        perms=\$(stat -c %a "\$file" 2>/dev/null || stat -f %Lp "\$file")
        sudo chmod "\$perms" "\$target_file.gz" 2>/dev/null || true
    done

    echo "清理临时目录..."
    rm -rf "\$TEMP_DIR"

    echo "设置权限..."
    sudo chown -R \$(whoami):\$(whoami) "$REMOTE_PATH" 2>/dev/null || true

    echo "统计压缩后的文件..."
    file_count=\$(find "$REMOTE_PATH" -type f -name "*.gz" | wc -l)
    total_size=\$(du -sh "$REMOTE_PATH" | cut -f1)
    echo "共部署 \$file_count 个压缩文件，总大小: \$total_size"

    echo "远端部署完成"
EOF

    # 清理本地压缩包
    rm -f "$ARCHIVE_NAME"
    info "已删除本地压缩包"
}

# 根据模式执行部署
case "$DEPLOY_MODE" in
    install-only)
        deploy_install_script
        ;;
    full|*)
        deploy_dist
        deploy_install_script
        ;;
esac

info "部署成功完成！"
info "文件已部署到: $REMOTE_HOST:$REMOTE_PATH"
