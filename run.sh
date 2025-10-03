#!/bin/bash

# 密码管理器运行脚本
# 用于编译和运行密码管理器应用程序

echo "🔧 正在编译密码管理器..."

# 编译程序
go build -o password_tool

# 检查编译是否成功
if [ $? -eq 0 ]; then
    echo "✅ 编译成功！"
    echo "🚀 启动密码管理器..."
    echo ""
    
    # 运行程序
    ./password_tool
else
    echo "❌ 编译失败！请检查代码错误。"
    exit 1
fi