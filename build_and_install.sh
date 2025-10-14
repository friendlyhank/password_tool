#!/bin/bash

# 密码管理器自动构建和安装脚本
# 功能：编译、打包、自动安装到系统应用程序文件夹

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查必要的工具
check_requirements() {
    print_info "检查构建环境..."
    
    # 检查 Go 是否安装
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装，请先安装 Go 语言环境"
        exit 1
    fi
    
    # 检查 fyne 工具是否安装
    if ! command -v fyne &> /dev/null && [ ! -f "$HOME/go/bin/fyne" ]; then
        print_warning "fyne 打包工具未安装，正在安装..."
        go install fyne.io/tools/cmd/fyne@latest
        if ! command -v fyne &> /dev/null && [ ! -f "$HOME/go/bin/fyne" ]; then
            print_error "fyne 工具安装失败"
            exit 1
        fi
        print_success "fyne 工具安装成功"
    fi
    
    # 检查图标文件是否存在
    if [ ! -f "icon.png" ]; then
        print_error "图标文件 icon.png 不存在"
        exit 1
    fi
    
    print_success "构建环境检查完成"
}

# 清理旧的构建文件
clean_old_builds() {
    print_info "清理旧的构建文件..."
    
    if [ -f "password_tool" ]; then
        rm -f password_tool
        print_info "删除旧的可执行文件"
    fi
    
    if [ -d "password_tool.app" ]; then
        rm -rf password_tool.app
        print_info "删除旧的应用包"
    fi
    
    print_success "清理完成"
}

# 编译应用程序
build_app() {
    print_info "开始编译应用程序..."
    
    # 安装依赖
    print_info "安装 Go 模块依赖..."
    go mod tidy
    
    # 编译应用
    print_info "编译可执行文件..."
    go build -o password_tool .
    
    if [ ! -f "password_tool" ]; then
        print_error "编译失败"
        exit 1
    fi
    
    print_success "应用程序编译完成"
}

# 打包 macOS 应用
package_app() {
    print_info "开始打包 macOS 应用..."
    
    # 使用完整路径调用 fyne 命令
    if command -v fyne &> /dev/null; then
        fyne package -os darwin -icon icon.png
    elif [ -f "$HOME/go/bin/fyne" ]; then
        $HOME/go/bin/fyne package -os darwin -icon icon.png
    else
        print_error "找不到 fyne 工具"
        exit 1
    fi
    
    if [ ! -d "password_tool.app" ]; then
        print_error "打包失败"
        exit 1
    fi
    
    print_success "macOS 应用打包完成"
}

# 安装应用到系统
install_app() {
    print_info "开始安装应用到系统..."
    
    # 检查 /Applications 目录权限
    if [ ! -w "/Applications" ]; then
        print_warning "需要管理员权限来安装应用到 /Applications 目录"
        print_info "请输入密码以继续安装..."
        sudo cp -R password_tool.app /Applications/
    else
        cp -R password_tool.app /Applications/
    fi
    
    # 验证安装
    if [ -d "/Applications/password_tool.app" ]; then
        print_success "应用已成功安装到 /Applications/password_tool.app"
    else
        print_error "应用安装失败"
        exit 1
    fi
}

# 显示完成信息
show_completion_info() {
    echo ""
    print_success "🎉 构建和安装完成！"
    echo ""
    print_info "应用信息："
    print_info "  • 应用名称: password_tool"
    print_info "  • 安装位置: /Applications/password_tool.app"
    print_info "  • 图标: 自定义锁和数据库图标"
    echo ""
    print_info "启动方式："
    print_info "  1. 从启动台找到 'password_tool' 并点击启动"
    print_info "  2. 从应用程序文件夹双击启动"
    print_info "  3. 使用命令行: open /Applications/password_tool.app"
    echo ""
    print_warning "首次运行注意事项："
    print_warning "  • 系统可能会提示'无法验证开发者'"
    print_warning "  • 解决方法: 右键点击应用 → 选择'打开' → 确认打开"
    print_warning "  • 或在'系统偏好设置' → '安全性与隐私' → '通用'中点击'仍要打开'"
    echo ""
}

# 主函数
main() {
    echo ""
    print_info "🔐 密码管理器自动构建和安装脚本"
    print_info "================================================"
    echo ""
    
    # 执行构建流程
    check_requirements
    clean_old_builds
    build_app
    package_app
    install_app
    show_completion_info
    
    print_success "✅ 所有操作完成！"
}

# 运行主函数
main "$@"