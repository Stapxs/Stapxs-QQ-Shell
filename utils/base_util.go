package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// UpdateMsg 页面固定周期 Update 消息
type UpdateMsg struct{}

// FilterStack 过滤堆栈信息
func FilterStack(stack []byte, packageName string) string {
	lines := strings.Split(string(stack), "\n")
	var filteredLines []string
	for _, line := range lines {
		if strings.Contains(line, packageName) {
			line = strings.Replace(line, packageName+"/", "", 1)
			filteredLines = append(filteredLines, line)
		}
	}
	return strings.Join(filteredLines, "\n")
}

// FilterError 过滤错误信息
func FilterError(err error, packageName string) string {
	return FilterStack([]byte(err.Error()), packageName)
}

// WriteLogToFile 记录日志到文件
func WriteLogToFile(logMessage string) error {
	// 获取当前日期并格式化为文件名
	currentDate := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("%s.log", currentDate)

	// 创建或打开文件以写入
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("关闭文件失败: %v\n", err)
		}
	}(file)

	// 获取当前时间并格式化
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 获取调用信息
	_, filePath, line, ok := runtime.Caller(1) // Caller(1) 获取直接调用该方法的地方
	if !ok {
		filePath = "未知文件"
		line = 0
	}
	functionName := "未知函数"
	if pc, _, _, ok := runtime.Caller(1); ok {
		functionName = runtime.FuncForPC(pc).Name()
	}

	// 拼接日志内容
	logEntry := fmt.Sprintf("[%s] [%s:%d %s] %s\n", timestamp, filePath, line, functionName, logMessage)

	// 写入文件
	if _, err := file.WriteString(logEntry); err != nil {
		return fmt.Errorf("写入日志失败: %w", err)
	}

	return nil
}

// GetTimeStr 将时间戳转换为字符串：2021-01-01 12:00:00
func GetTimeStr(timeUnix float64) string {
	t := time.Unix(int64(timeUnix), 0)
	return t.Format("2006-01-02 15:04:05")
}

func InArray(array []string, item string) bool {
	for _, v := range array {
		if v == item {
			return true
		}
	}
	return false
}
