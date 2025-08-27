package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	inputFile  string
	outputFile string
	demoMode   bool
)

type Segment struct {
	Start float64
	End   float64
	Text  string
}

type SRTEntry struct {
	Index int
	Start time.Duration
	End   time.Duration
	Text  string
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "audio-to-srt",
		Short: "Convert audio files to SRT subtitles using Whisper",
		Run:   runTranscription,
	}

	rootCmd.Flags().StringVarP(&inputFile, "input", "t", "", "Input audio file path (required)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output SRT file path (required)")
	rootCmd.Flags().BoolVar(&demoMode, "demo", false, "使用演示模式（不需要真实的Whisper工具）")
	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runTranscription(cmd *cobra.Command, args []string) {
	if !fileExists(inputFile) {
		log.Fatalf("输入文件不存在: %s", inputFile)
	}

	fmt.Printf("正在处理音频文件: %s\n", inputFile)
	
	segments, err := transcribeAudio(inputFile)
	if err != nil {
		log.Fatalf("音频转录失败: %v", err)
	}

	fmt.Printf("转录完成，共获得 %d 个片段\n", len(segments))

	err = generateSRT(segments, outputFile)
	if err != nil {
		log.Fatalf("生成SRT文件失败: %v", err)
	}

	fmt.Printf("SRT文件已生成: %s\n", outputFile)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func transcribeAudio(audioFile string) ([]Segment, error) {
	// 演示模式：生成模拟的转录结果
	if demoMode {
		return generateDemoTranscription(audioFile)
	}

	// 方案1: 优先使用whisper.cpp命令行工具（本地编译的）
	if whisperPath := findWhisperExecutable(); whisperPath != "" {
		return transcribeWithWhisperCpp(audioFile, whisperPath)
	}

	// 方案2: 尝试使用系统的whisper命令
	if whisperPath := findSystemWhisper(); whisperPath != "" {
		return transcribeWithSystemWhisper(audioFile, whisperPath)
	}

	// 方案3: 尝试使用Python的whisper
	if pythonPath := findPython(); pythonPath != "" {
		return transcribeWithPythonWhisper(audioFile, pythonPath)
	}

	return nil, fmt.Errorf("未找到可用的Whisper工具。请安装whisper.cpp或Python whisper库，或使用 --demo 参数运行演示模式")
}

func findSystemWhisper() string {
	if path, err := exec.LookPath("whisper"); err == nil {
		return path
	}
	return ""
}

func findWhisperExecutable() string {
	// 首先检查当前目录下的whisper.cpp构建结果
	localPaths := []string{
		"./whisper.cpp/build/bin/whisper-cli",
		"./whisper.cpp/build/bin/main",
		"./whisper.cpp/main",
	}
	
	for _, path := range localPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	// 检查系统路径
	systemPaths := []string{
		"whisper.cpp",
		"whisper-cli",
		"/usr/local/bin/whisper.cpp",
		"/opt/homebrew/bin/whisper.cpp",
	}

	for _, path := range systemPaths {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}
	return ""
}

func findPython() string {
	possiblePaths := []string{"python3", "python"}
	for _, path := range possiblePaths {
		if _, err := exec.LookPath(path); err == nil {
			// 检查是否安装了whisper
			cmd := exec.Command(path, "-c", "import whisper; print('ok')")
			if err := cmd.Run(); err == nil {
				return path
			}
		}
	}
	return ""
}

func transcribeWithSystemWhisper(audioFile, whisperPath string) ([]Segment, error) {
	tempDir, err := os.MkdirTemp("", "whisper_output")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 获取音频文件名（不含扩展名）
	audioName := strings.TrimSuffix(filepath.Base(audioFile), filepath.Ext(audioFile))
	expectedSRTFile := filepath.Join(tempDir, audioName + ".srt")
	
	cmd := exec.Command(whisperPath, 
		"--output_format", "srt",
		"--output_dir", tempDir,
		audioFile)
	
	// 捕获标准输出和标准错误
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		// 检查是否是SSL证书问题
		if strings.Contains(outputStr, "SSL: CERTIFICATE_VERIFY_FAILED") {
			return nil, fmt.Errorf("whisper模型下载失败 (SSL证书问题)。可能的解决方案：\n1. 使用 --demo 参数运行演示模式\n2. 手动下载whisper模型\n3. 配置网络代理\n\n错误详情: %v", err)
		}
		return nil, fmt.Errorf("whisper命令执行失败: %v\n输出: %s", err, outputStr)
	}

	return parseSRTFile(expectedSRTFile)
}

func transcribeWithWhisperCpp(audioFile, whisperPath string) ([]Segment, error) {
	tempDir, err := os.MkdirTemp("", "whisper_output")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 检查是否有模型文件
	modelPath := "./whisper.cpp/models/ggml-base.bin"
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("未找到whisper模型文件: %s", modelPath)
	}

	outputFile := filepath.Join(tempDir, "output")
	
	cmd := exec.Command(whisperPath, 
		"-m", modelPath,
		"-f", audioFile,
		"-of", outputFile,
		"--output-srt")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper.cpp执行失败: %v\n输出: %s", err, string(output))
	}

	srtFile := outputFile + ".srt"
	return parseSRTFile(srtFile)
}

func transcribeWithPythonWhisper(audioFile, pythonPath string) ([]Segment, error) {
	tempScript := `
import whisper
import sys
import json

model = whisper.load_model("base")
result = model.transcribe(sys.argv[1])

segments = []
for seg in result["segments"]:
    segments.append({
        "start": seg["start"],
        "end": seg["end"], 
        "text": seg["text"].strip()
    })

print(json.dumps(segments))
`

	tempFile, err := os.CreateTemp("", "whisper_script_*.py")
	if err != nil {
		return nil, fmt.Errorf("创建临时脚本失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(tempScript); err != nil {
		return nil, fmt.Errorf("写入临时脚本失败: %v", err)
	}
	tempFile.Close()

	cmd := exec.Command(pythonPath, tempFile.Name(), audioFile)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Python whisper执行失败: %v", err)
	}

	var segments []Segment
	if err := json.Unmarshal(output, &segments); err != nil {
		return nil, fmt.Errorf("解析JSON输出失败: %v", err)
	}

	return segments, nil
}

func parseSRTFile(srtFile string) ([]Segment, error) {
	file, err := os.Open(srtFile)
	if err != nil {
		return nil, fmt.Errorf("打开SRT文件失败: %v", err)
	}
	defer file.Close()

	var segments []Segment
	scanner := bufio.NewScanner(file)
	
	timeRegex := regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2}),(\d{3}) --> (\d{2}):(\d{2}):(\d{2}),(\d{3})`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if timeRegex.MatchString(line) {
			matches := timeRegex.FindStringSubmatch(line)
			if len(matches) == 9 {
				startTime := parseTimeToSeconds(matches[1], matches[2], matches[3], matches[4])
				endTime := parseTimeToSeconds(matches[5], matches[6], matches[7], matches[8])
				
				// 读取下一行作为文本
				if scanner.Scan() {
					text := strings.TrimSpace(scanner.Text())
					segments = append(segments, Segment{
						Start: startTime,
						End:   endTime,
						Text:  text,
					})
				}
			}
		}
	}

	return segments, scanner.Err()
}

func parseTimeToSeconds(hours, minutes, seconds, milliseconds string) float64 {
	h, _ := strconv.Atoi(hours)
	m, _ := strconv.Atoi(minutes)
	s, _ := strconv.Atoi(seconds)
	ms, _ := strconv.Atoi(milliseconds)
	
	return float64(h*3600 + m*60 + s) + float64(ms)/1000.0
}

func generateDemoTranscription(audioFile string) ([]Segment, error) {
	fmt.Printf("演示模式：为 %s 生成模拟转录结果\n", audioFile)
	
	// 生成基于文件名的模拟转录内容
	audioName := strings.TrimSuffix(filepath.Base(audioFile), filepath.Ext(audioFile))
	
	segments := []Segment{
		{Start: 0.0, End: 3.5, Text: fmt.Sprintf("欢迎收听 %s 音频文件的转录结果。", audioName)},
		{Start: 3.5, End: 7.2, Text: "这是一个演示模式生成的字幕文件。"},
		{Start: 7.2, End: 11.8, Text: "在实际使用中，这里会显示真正的语音识别结果。"},
		{Start: 11.8, End: 15.3, Text: "请安装whisper.cpp或Python whisper来获得真实的转录功能。"},
		{Start: 15.3, End: 18.0, Text: "谢谢使用我们的音频转字幕工具！"},
	}
	
	return segments, nil
}

func generateSRT(segments []Segment, outputFile string) error {
	var srtEntries []SRTEntry
	
	for i, seg := range segments {
		entry := SRTEntry{
			Index: i + 1,
			Start: time.Duration(seg.Start * float64(time.Second)),
			End:   time.Duration(seg.End * float64(time.Second)),
			Text:  strings.TrimSpace(seg.Text),
		}
		srtEntries = append(srtEntries, entry)
	}

	return writeSRTFile(srtEntries, outputFile)
}

func writeSRTFile(entries []SRTEntry, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, entry := range entries {
		_, err := fmt.Fprintf(file, "%d\n%s --> %s\n%s\n\n",
			entry.Index,
			formatTime(entry.Start),
			formatTime(entry.End),
			entry.Text)
		if err != nil {
			return err
		}
	}

	return nil
}

func formatTime(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	milliseconds := int(d.Nanoseconds()/1000000) % 1000

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, milliseconds)
}