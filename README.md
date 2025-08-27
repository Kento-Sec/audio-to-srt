# Audio to SRT 音频转字幕工具

一个基于Go语言开发的高性能音频转字幕工具，支持多种Whisper后端，可将音频文件快速转换为标准SRT格式字幕。

## ✨ 功能特性

- 🚀 **多种Whisper后端支持**：whisper.cpp、Python whisper、演示模式
- 🎯 **命令行界面**：简单易用的CLI工具
- 📱 **跨平台支持**：Windows、macOS、Linux
- ⚡ **高性能**：基于whisper.cpp，转录速度快
- 🎬 **标准格式**：生成标准SRT字幕格式
- 🛠️ **开箱即用**：内置演示模式，无需外部依赖

## 📋 系统要求

- Go 1.21 或更高版本
- macOS、Linux 或 Windows 系统
- 至少 4GB RAM（推荐8GB）

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd audio-to-srt-project
```

### 2. 编译项目

```bash
go build -o audio-to-srt
```

### 3. 快速测试（演示模式）

```bash
./audio-to-srt -t your-audio.wav -o output.srt --demo
```

## 🛠️ 安装Whisper后端

### 方法一：使用whisper.cpp（推荐）

项目已包含whisper.cpp源码，只需编译和下载模型：

```bash
# 编译whisper.cpp
cd whisper.cpp
make

# 下载base模型（约147MB）
bash ./models/download-ggml-model.sh base

# 返回项目根目录
cd ..
```

### 方法二：安装Python whisper

```bash
pip install openai-whisper
```

### 方法三：使用Homebrew安装whisper.cpp

```bash
brew install whisper-cpp
```

## 📖 使用说明

### 基本命令格式

```bash
./audio-to-srt -t <输入文件> -o <输出文件> [选项]
```

### 参数说明

| 参数 | 长参数 | 描述 | 必需 |
|------|--------|------|------|
| `-t` | `--input` | 输入音频文件路径 | ✅ |
| `-o` | `--output` | 输出SRT文件路径 | ✅ |
| | `--demo` | 使用演示模式（生成模拟字幕） | ❌ |
| `-h` | `--help` | 显示帮助信息 | ❌ |

### 使用示例

#### 1. 基本转录（自动选择最佳后端）
```bash
./audio-to-srt -t meeting.wav -o meeting.srt
```

#### 2. 演示模式（无需Whisper工具）
```bash
./audio-to-srt -t podcast.mp3 -o podcast.srt --demo
```

#### 3. 批量处理
```bash
# 处理多个文件
for file in *.wav; do
    ./audio-to-srt -t "$file" -o "${file%.wav}.srt"
done
```

### 支持的音频格式

- ✅ WAV (推荐)
- ✅ MP3
- ✅ MP4/M4A
- ✅ FLAC
- ✅ OGG
- ✅ 其他FFmpeg支持的格式

## 🏗️ 项目结构

```
audio-to-srt-project/
├── README.md              # 项目说明文档
├── main.go               # 主程序源码
├── go.mod               # Go模块配置
├── go.sum               # Go依赖校验
├── audio-to-srt         # 编译后的可执行文件
└── whisper.cpp/         # Whisper.cpp源码和模型
    ├── build/bin/       # 编译后的whisper工具
    ├── models/          # Whisper模型文件
    └── ...
```

## ⚙️ 工作原理

1. **自动检测**：程序按优先级检测可用的Whisper工具
   - 优先：本地whisper.cpp (./whisper.cpp/build/bin/whisper-cli)
   - 次选：系统whisper命令
   - 备选：Python whisper库
   - 最后：演示模式

2. **音频处理**：调用相应的Whisper工具进行语音识别

3. **格式转换**：将识别结果转换为标准SRT格式

4. **文件输出**：保存为指定的SRT字幕文件

## 🔧 高级配置

### 模型选择

默认使用base模型，如需其他模型：

```bash
# 下载其他模型
cd whisper.cpp
bash ./models/download-ggml-model.sh small   # 约466MB
bash ./models/download-ggml-model.sh medium  # 约1.5GB
bash ./models/download-ggml-model.sh large   # 约2.9GB
```

### 性能调优

- **CPU使用**：whisper.cpp会自动使用所有可用CPU核心
- **内存占用**：base模型约需1GB RAM，large模型需4GB+
- **GPU加速**：如需GPU支持，请重新编译whisper.cpp并启用CUDA/Metal

## 🚨 故障排除

### 常见问题

#### 1. "未找到可用的Whisper工具"
**解决方案**：
```bash
# 使用演示模式测试
./audio-to-srt -t input.wav -o output.srt --demo

# 或编译whisper.cpp
cd whisper.cpp && make
```

#### 2. "SSL证书验证失败"
**解决方案**：
- 使用演示模式：`--demo`
- 或手动下载模型文件
- 或配置网络代理

#### 3. "音频文件格式不支持"
**解决方案**：
```bash
# 使用FFmpeg转换为WAV格式
ffmpeg -i input.mp4 -ar 16000 output.wav
```

#### 4. 转录结果不准确
**解决方案**：
- 使用更大的模型（medium、large）
- 确保音频质量良好
- 尝试不同的音频格式

### 调试模式

如遇问题，可查看详细错误信息：

```bash
# Go程序会显示详细的错误信息
./audio-to-srt -t problem.wav -o output.srt
```

## 🔄 更新和维护

### 更新whisper.cpp
```bash
cd whisper.cpp
git pull
make clean
make
```

### 更新Go依赖
```bash
go mod tidy
go mod download
```

## 📊 性能基准

| 模型 | 文件大小 | 内存占用 | 处理速度* | 准确度 |
|------|----------|----------|-----------|--------|
| tiny | ~37MB | ~500MB | ~10x | 中等 |
| base | ~147MB | ~1GB | ~6x | 良好 |
| small | ~466MB | ~2GB | ~4x | 优秀 |
| medium | ~1.5GB | ~4GB | ~2x | 极佳 |

*相对于音频实际时长的倍数，在Apple M1芯片上测试

## 🤝 贡献指南

欢迎提交问题和改进建议！

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启Pull Request

## 📄 许可证

本项目基于 MIT 许可证开源 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) - 高性能的Whisper实现
- [OpenAI Whisper](https://github.com/openai/whisper) - 原始的Whisper模型
- [Cobra](https://github.com/spf13/cobra) - Go CLI框架

## 📞 支持

如有问题或建议，请：
- 提交 [Issue](https://github.com/your-repo/issues)
- 发送邮件至：your-email@domain.com

---

**开始享受高效的音频转字幕体验吧！** 🎉