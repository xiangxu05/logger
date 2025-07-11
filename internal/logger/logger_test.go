package logger

import (
	"strings"
	"testing"
	"time"
)

// 测试初始化、日志写入、通道关闭等核心逻辑
func TestLoggerBasic(t *testing.T) {
	cfg := Config{
		MinLevel:      DEBUG,
		Format:        FormatPlain,
		Targets:       OutputConsole, // 只控制台，避免文件IO影响测试
		LogPath:       "logs/test.log",
		AllowedPrefix: []string{"logger"},
	}

	log := GetLoggerInstance(cfg)

	// 写入各级别日志
	log.Debug("debug msg")
	log.Info("info msg")
	log.Warn("warn msg")
	log.Error("error msg")

	// 简单延迟，确保日志写入协程处理完
	time.Sleep(100 * time.Millisecond)

	// 由于日志写入是异步的，直接检测通道是否关闭前，先关闭
	log.Close()

	// 关闭后，不能再写入日志
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on writing after Close, but no panic occurred")
		}
	}()
	log.Info("写入关闭后日志，应panic")
}

// 测试 RecoverAndLogPanic 捕获 panic 的逻辑
func TestRecoverAndLogPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RecoverAndLogPanic did not catch panic, got: %v", r)
		}
	}()

	func() {
		defer RecoverAndLogPanic()
		panic("test panic")
	}()
}

// 测试 shouldAllow 功能
func TestShouldAllow(t *testing.T) {
	cfg := Config{
		AllowedPrefix: []string{"logger"},
	}

	log := GetLoggerInstance(cfg)

	cases := []struct {
		caller string
		allow  bool
	}{
		{"mainpkg.func", false},
		{"logger.func", true},
		{"otherpkg.func", false},
		{"", false},
	}

	for _, c := range cases {
		got := log.shouldAllow(c.caller)
		if got != c.allow {
			t.Errorf("shouldAllow(%q) = %v; want %v", c.caller, got, c.allow)
		}
	}
}

// 测试 getCaller 返回合理格式（略做简单断言）
func TestGetCallerFormat(t *testing.T) {
	caller := getCaller()
	if !strings.Contains(caller, "asm_amd64") && !strings.Contains(caller, "runtime.goexit") {
		t.Errorf("getCaller returned unexpected value: %s", caller)
	}
}
