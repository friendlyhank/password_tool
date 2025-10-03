#!/bin/bash

# 密码管理器数据清除脚本
# 用于清除所有主密码和密码条目数据

DB_DIR="$HOME/.password_tool"
DB_FILE="$DB_DIR/passwords.db"

echo "⚠️  警告：此操作将永久删除所有密码数据！"
echo "📁 数据库位置: $DB_FILE"
echo ""

# 检查数据库文件是否存在
if [ -f "$DB_FILE" ]; then
    echo "📊 当前数据库文件大小: $(ls -lh "$DB_FILE" | awk '{print $5}')"
    echo ""
    
    # 确认操作
    read -p "❓ 确定要删除所有数据吗？输入 'YES' 确认: " confirmation
    
    if [ "$confirmation" = "YES" ]; then
        # 删除数据库文件
        rm "$DB_FILE"
        
        if [ $? -eq 0 ]; then
            echo "✅ 数据库文件已成功删除！"
            echo "🔄 下次启动应用时将重新设置主密码。"
        else
            echo "❌ 删除失败！请检查文件权限。"
            exit 1
        fi
    else
        echo "🚫 操作已取消。"
        exit 0
    fi
else
    echo "ℹ️  数据库文件不存在，无需清除。"
fi

echo ""
echo "📂 当前 .password_tool 目录内容:"
ls -la "$DB_DIR" 2>/dev/null || echo "目录不存在或为空"