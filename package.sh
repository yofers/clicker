#!/bin/bash

# Configuration
APP_NAME="Clicker"
VERSION="v1.2.14"
OUTPUT_DIR="build/release"
BIN_DIR="build/bin"
WAILS_CMD="/Users/mahiro/go/bin/wails"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "🚀 开始构建流程 - 版本: $VERSION"

# Build binaries
echo "🔨 正在编译应用 (macOS arm64 & Windows amd64)..."
$WAILS_CMD build -platform darwin/arm64,windows/amd64

if [ $? -ne 0 ]; then
    echo "❌ 编译失败，请检查错误日志"
    exit 1
fi
echo "✅ 编译成功！"

echo "🚀 开始打包流程..."

# ==========================================
# macOS Packaging
# ==========================================
echo ""
echo "🍎 [macOS] 正在处理 macOS 包..."
MAC_APP_PATH="$BIN_DIR/${APP_NAME}.app"
MAC_ZIP_NAME="clicker-release-${VERSION}-MAC-Arm64.zip"

if [ -d "$MAC_APP_PATH" ]; then
    echo "   - 📝 正在应用临时签名..."
    codesign --force --deep --sign - "$MAC_APP_PATH"
    
    echo "   - 📦 正在生成 ZIP 包: $MAC_ZIP_NAME"
    # Use ditto for best macOS compatibility
    ditto -c -k --sequesterRsrc --keepParent "$MAC_APP_PATH" "$OUTPUT_DIR/$MAC_ZIP_NAME"
    echo "   - ✅ macOS 打包完成"
else
    echo "   - ❌ 错误: 找不到 macOS 应用文件 $MAC_APP_PATH"
fi

# ==========================================
# Windows Packaging
# ==========================================
echo ""
echo "🪟 [Windows] 正在处理 Windows 包..."
WIN_EXE_NAME="clicker-release-${VERSION}-amd64.exe"
WIN_EXE_PATH="$BIN_DIR/$WIN_EXE_NAME"

if [ -f "$WIN_EXE_PATH" ]; then
    echo "   - 📋 正在复制 EXE 文件并重命名..."
    # Copy and rename to remove -amd64 suffix
    TARGET_NAME="clicker-release-${VERSION}-Win-amd64.exe"
    cp "$WIN_EXE_PATH" "$OUTPUT_DIR/$TARGET_NAME"
    echo "   - ✅ Windows 处理完成: $TARGET_NAME"
else
    echo "   - ❌ 错误: 找不到 Windows 可执行文件 $WIN_EXE_PATH"
fi

echo ""
echo "🎉 所有打包任务已完成！文件位于 $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
