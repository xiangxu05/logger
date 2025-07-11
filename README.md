# Logger

> 一个功能完善的 Go 语言日志库，支持多输出目标（Console、文件），异步写入，日志等级过滤，文件自动轮转，JSON 格式化，以及 panic 自动捕获。

---

## 主要特性

- 🚀 **全局单例 & 异步写日志**，高效且不阻塞业务线程
- 🎯 **多输出目标**：控制台、日志文件（支持自动轮转）
- 🎨 **支持纯文本与 JSON 格式化**，满足不同需求
- 🔒 **日志等级过滤**：DEBUG / INFO / WARN / ERROR，精准控制日志输出
- 🛡 **Panic 自动捕获并记录**，方便调试和运维
- ⚙️ **配置灵活**：通过结构体一键配置所有参数，默认合理，使用简单
- 💾 **文件自动轮转**：基于 `lumberjack`，自动管理日志大小和备份数量

---

## 安装

```bash
go get github.com/xiangxu05/logger
```

---

## 快速开始

```go
package main

import (
    "time"
    "github.com/xiangxu05/logger"
)

func main() {
    // 配置日志选项
    cfg := logger.Config{
        MinLevel:      logger.DEBUG,
        Format:        logger.FormatPlain,        // 可选 logger.FormatJSON
        Targets:       logger.OutputConsole | logger.OutputFile,
        LogPath:       "logs/app.log",
        AllowedPrefix: []string{"main"},          // 仅白名单包名前缀日志额外输出到 logs_allowed/allowed.log
    }

    // 获取全局 Logger 单例
    log := logger.GetLoggerInstance(cfg)

    // 捕获程序panic并记录
    defer logger.RecoverAndLogPanic()
    // 程序退出时关闭Logger，确保日志写入完成
    defer func() {
        time.Sleep(time.Millisecond * 100)
        log.Close()
    }()

    // 记录各种等级日志
    log.Debug("调试信息")
    log.Info("服务启动成功")
    log.Warn("警告信息")
    log.Error("错误信息")

    // 模拟 panic，测试自动捕获
    panic("测试 panic 捕获")
}
```

---

## 配置说明

| 参数          | 类型           | 默认值          | 说明                                                            |
| ------------- | -------------- | --------------- | --------------------------------------------------------------- |
| MinLevel      | `Level`        | `INFO`          | 最低日志输出等级                                                |
| Format        | `Format`       | `FormatPlain`   | 日志格式，支持纯文本和 JSON                                     |
| Targets       | `OutputTarget` | `OutputConsole` | 输出目标，可按位组合（Console、File）                           |
| LogPath       | `string`       | `logs/log.json` | 日志文件路径，支持自动轮转                                      |
| AllowedPrefix | `[]string`     | `[]`（空列表）  | 白名单包名前缀，符合条件日志额外写入 `logs_allowed/allowed.log` |

---

## 支持日志等级

- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`

---

## 输出目标（可组合）

| 名称            | 说明                       |
| --------------- | -------------------------- |
| `OutputConsole` | 输出到终端控制台           |
| `OutputFile`    | 输出到日志文件（自动轮转） |

---

## Panic 自动捕获示例

```go
func main() {
    defer logger.RecoverAndLogPanic()

    log := logger.GetLoggerInstance()

    panic("模拟崩溃")
}
```
