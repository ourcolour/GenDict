package utils

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// InitLogger 初始化全局日志记录器，支持双写（控制台和文件），并模拟Tomcat格式
// mode: 可选 "production" 或 "development"
// logPath: 日志文件存放路径，例如 "./logs/app.log"
func InitLogger(mode, logPath string) error {
	var (
		level  slog.Level
		writer io.Writer
	)

	// 1. 配置日志级别
	switch mode {
	case "development":
		level = slog.LevelDebug
	case "production":
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo
	}

	// 2. 确保日志目录存在
	//if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
	//	return err
	//}
	//
	//// 3. 配置日志切割 (使用lumberjack)
	//fileWriter := &lumberjack.Logger{
	//	Filename:   logPath,
	//	MaxSize:    10,    // 单个文件最大10MB
	//	MaxBackups: 5,     // 保留5个备份
	//	MaxAge:     30,    // 保留30天
	//	Compress:   false, // 不压缩
	//}

	// 4. 创建多写入器：同时输出到文件和控制台
	//writer = io.MultiWriter(fileWriter, os.Stdout)
	writer = io.MultiWriter(os.Stdout)

	// 5. 创建自定义Handler以模拟Tomcat日志格式
	handler := NewTomcatFormatHandler(writer, &slog.HandlerOptions{
		Level: level,
	})

	// 6. 设置为全局默认logger
	slog.SetDefault(slog.New(handler))
	return nil
}

// TomcatFormatHandler 自定义Handler模拟Tomcat格式
type TomcatFormatHandler struct {
	slog.Handler
	writer io.Writer
}

func NewTomcatFormatHandler(w io.Writer, opts *slog.HandlerOptions) *TomcatFormatHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	// 使用JSONHandler作为基础，但我们会重写Format方法
	baseHandler := slog.NewJSONHandler(w, opts)
	return &TomcatFormatHandler{
		Handler: baseHandler,
		writer:  w,
	}
}

// Handle 重写Handle方法，格式化输出以匹配Tomcat风格
// 添加context.Context参数以符合slog.Handler接口
func (h *TomcatFormatHandler) Handle(ctx context.Context, r slog.Record) error {
	// 模拟Tomcat格式: [时间] [级别] [消息] [可选键值对]
	// 示例: [2025-09-19 10:30:45] [INFO] Application started protocol=http
	timeStr := r.Time.Format("2006-01-02 15:04:05")
	levelStr := r.Level.String()
	message := r.Message

	// 构建基本日志行
	logLine := "[" + timeStr + "] [" + levelStr + "] " + message

	// 添加附加属性（键值对）
	r.Attrs(func(attr slog.Attr) bool {
		logLine += " " + attr.Key + "=" + attr.Value.String()
		return true
	})

	logLine += "\n"

	// 写入到输出
	_, err := h.writer.Write([]byte(logLine))
	return err
}

// 添加Enabled方法以符合slog.Handler接口
func (h *TomcatFormatHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// 添加WithAttrs方法以符合slog.Handler接口
func (h *TomcatFormatHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TomcatFormatHandler{
		Handler: h.Handler.WithAttrs(attrs),
		writer:  h.writer,
	}
}

// 添加WithGroup方法以符合slog.Handler接口
func (h *TomcatFormatHandler) WithGroup(name string) slog.Handler {
	return &TomcatFormatHandler{
		Handler: h.Handler.WithGroup(name),
		writer:  h.writer,
	}
}
