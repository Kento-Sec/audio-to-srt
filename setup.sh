#!/bin/bash

# Audio-to-SRT 项目安装脚本
# 自动编译Go程序和whisper.cpp

set -e

echo "🚀 开始安装 Audio-to-SRT 项目..."

# 检查Go版本
check_go() {
    if ! command -v go &> /dev/null; then
        echo "❌ 未找到Go，请先安装Go 1.21+: https://golang.org/dl/"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "✅ 发现Go版本: $GO_VERSION"
}

# 编译Go程序
build_go() {
    echo "📦 编译Go程序..."
    go build -o audio-to-srt
    chmod +x audio-to-srt
    echo "✅ Go程序编译完成"
}

# 编译whisper.cpp
build_whisper() {
    echo "🔨 编译whisper.cpp..."
    cd whisper.cpp
    make
    cd ..
    echo "✅ whisper.cpp编译完成"
}

# 下载Whisper模型
download_model() {
    echo "⬇️ 下载Whisper base模型..."
    cd whisper.cpp
    if [ ! -f "models/ggml-base.bin" ]; then
        bash ./models/download-ggml-model.sh base
        echo "✅ 模型下载完成"
    else
        echo "✅ 模型已存在，跳过下载"
    fi
    cd ..
}

# 测试安装
test_installation() {
    echo "🧪 测试安装..."
    if ./audio-to-srt --help > /dev/null 2>&1; then
        echo "✅ 安装测试通过"
        echo ""
        echo "🎉 安装完成！"
        echo ""
        echo "📖 使用方法："
        echo "  ./audio-to-srt -t input.wav -o output.srt --demo  # 演示模式"
        echo "  ./audio-to-srt -t input.wav -o output.srt         # 真实转录"
        echo ""
        echo "📚 更多信息请查看 README.md"
    else
        echo "❌ 安装测试失败"
        exit 1
    fi
}

# 主安装流程
main() {
    check_go
    build_go
    
    # 询问是否编译whisper.cpp
    read -p "是否编译whisper.cpp以获得最佳性能？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        build_whisper
        
        read -p "是否下载Whisper base模型？(约147MB) (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            download_model
        fi
    else
        echo "⚠️ 跳过whisper.cpp编译，可使用演示模式或系统whisper"
    fi
    
    test_installation
}

# 运行主函数
main "$@"