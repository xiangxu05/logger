package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func levelToStr(l Level) string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	default:
		return "UNKNOWN"
	}
}

type Format int

const (
	FormatPlain Format = iota
	FormatJSON
)

type Config struct {
	MinLevel      Level
	Format        Format
	Targets       OutputTarget
	LogPath       string
	AllowedPrefix []string // 白名单包名前缀
}

type OutputTarget int

const (
	OutputNone    OutputTarget = 0
	OutputConsole OutputTarget = 1 << iota
	OutputFile
)

type logMsg struct {
	Level   Level
	Message string
	Time    time.Time
	Caller  string
}

type Logger struct {
	logChan         chan logMsg
	quit            chan struct{}
	config          Config
	fileLogger      *lumberjack.Logger
	allowFileLogger *lumberjack.Logger
}

var (
	instance *Logger
	once     sync.Once
	cfg      = Config{
		MinLevel:      INFO,
		Format:        FormatPlain,
		Targets:       OutputConsole,
		LogPath:       "logs/log.json",
		AllowedPrefix: []string{},
	}
)

func GetLoggerInstance(cfgs ...Config) *Logger {
	once.Do(func() {
		if len(cfgs) > 0 {
			cfg = cfgs[0]
		}

		if cfg.Targets&OutputFile == 1 {
			logDir := filepath.Dir(cfg.LogPath)
			if logDir != "" {
				_ = os.MkdirAll(logDir, 0755)
			}
		}

		// 如果配置了白名单输出，创建 logs_allowed/allowed.log
		if len(cfg.AllowedPrefix) > 0 {
			_ = os.MkdirAll("logs_allowed", 0755)
		}

		instance = &Logger{
			logChan: make(chan logMsg, 1000),
			quit:    make(chan struct{}),
			config:  cfg,
		}

		if cfg.Targets&OutputFile != 0 {
			instance.fileLogger = &lumberjack.Logger{
				Filename:   cfg.LogPath,
				MaxSize:    10,
				MaxBackups: 5,
				MaxAge:     7,
				Compress:   true,
			}
		}
		if len(cfg.AllowedPrefix) > 0 {
			instance.allowFileLogger = &lumberjack.Logger{
				Filename:   "logs_allowed/allowed.log",
				MaxSize:    10,
				MaxBackups: 5,
				MaxAge:     7,
				Compress:   true,
			}
		}

		go instance.start()
	})
	return instance
}

func (l *Logger) start() {
	for {
		select {
		case msg := <-l.logChan:
			formatted := l.formatLog(msg)

			if l.config.Targets&OutputConsole != 0 {
				fmt.Print(colorize(msg.Level, formatted))
			}
			if l.config.Targets&OutputFile != 0 {
				l.fileLogger.Write([]byte(formatted))
			}

			if l.allowFileLogger != nil && l.shouldAllow(msg.Caller) {
				l.allowFileLogger.Write([]byte(formatted))
			}
		case <-l.quit:
			close(l.logChan)
			for msg := range l.logChan {
				formatted := l.formatLog(msg)
				if l.config.Targets&OutputFile != 0 {
					l.fileLogger.Write([]byte(formatted))
				}
				if l.allowFileLogger != nil && l.shouldAllow(msg.Caller) {
					l.allowFileLogger.Write([]byte(formatted))
				}
			}
			return
		}
	}
}

func (l *Logger) shouldAllow(caller string) bool {
	if len(l.config.AllowedPrefix) == 0 {
		return false
	}
	for _, prefix := range l.config.AllowedPrefix {
		if strings.Contains(caller, prefix) {
			return true
		}
	}
	return false
}

func (l *Logger) formatLog(msg logMsg) string {
	if l.config.Format == FormatJSON {
		data := map[string]interface{}{
			"level":   levelToStr(msg.Level),
			"time":    msg.Time.Format(time.RFC3339),
			"message": msg.Message,
			"caller":  msg.Caller,
		}
		b, _ := json.Marshal(data)
		return string(b) + "\n"
	}
	return fmt.Sprintf("[%s] %s %s %s\n",
		levelToStr(msg.Level),
		msg.Time.Format("2006-01-02 15:04:05"),
		msg.Caller,
		msg.Message,
	)
}

func getCaller() string {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc).Name()
	parts := strings.Split(fn, "/")
	shortFunc := parts[len(parts)-1]
	parts = strings.Split(file, "/")
	shortFile := parts[len(parts)-1]
	return fmt.Sprintf("%s:%d %s", shortFile, line, shortFunc)
}

func colorize(level Level, msg string) string {
	switch level {
	case DEBUG:
		return "\033[36m" + msg + "\033[0m" // Cyan
	case INFO:
		return "\033[32m" + msg + "\033[0m" // Green
	case WARN:
		return "\033[33m" + msg + "\033[0m" // Yellow
	case ERROR:
		return "\033[31m" + msg + "\033[0m" // Red
	default:
		return msg
	}
}

func (l *Logger) log(level Level, msg string) {
	if level < l.config.MinLevel {
		return
	}
	l.logChan <- logMsg{
		Level:   level,
		Message: msg,
		Time:    time.Now(),
		Caller:  getCaller(),
	}
}

func (l *Logger) Info(msg string)  { l.log(INFO, msg) }
func (l *Logger) Error(msg string) { l.log(ERROR, msg) }
func (l *Logger) Debug(msg string) { l.log(DEBUG, msg) }
func (l *Logger) Warn(msg string)  { l.log(WARN, msg) }

func (l *Logger) Close() {
	close(l.quit)
	if l.fileLogger != nil {
		_ = l.fileLogger.Close()
	}
	if l.allowFileLogger != nil {
		_ = l.allowFileLogger.Close()
	}
}

func RecoverAndLogPanic() {
	if r := recover(); r != nil {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		msg := fmt.Sprintf("Panic recovered: %v\n%s", r, string(buf[:n]))

		log := GetLoggerInstance()
		formatted := log.formatLog(logMsg{
			Level:   ERROR,
			Message: msg,
			Time:    time.Now(),
			Caller:  getCaller(),
		})

		if log.config.Targets&OutputConsole != 0 {
			fmt.Print(colorize(ERROR, formatted))
		}
		if log.config.Targets&OutputFile != 0 && log.fileLogger != nil {
			log.fileLogger.Write([]byte(formatted))
		}
		if log.allowFileLogger != nil && log.shouldAllow(getCaller()) {
			log.allowFileLogger.Write([]byte(formatted))
		}
	}
}
