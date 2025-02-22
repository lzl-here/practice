package main

import (
	data "async-order/internal/data"
	server "async-order/internal/server"
	"log"
	"log/slog"
	"os"
)

func main() {

	// 创建日志文件（追加模式）
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("打开日志文件失败:", err)
	}
	defer file.Close()

	// 配置slog日志处理器
	logHandler := slog.NewTextHandler(file, &slog.HandlerOptions{
		AddSource: true,            // 记录调用位置
		Level:     slog.LevelDebug, // 设置日志级别
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey { // 自定义时间格式
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.000"))
			}
			return a
		},
	})

	// 设置全局logger
	slog.SetDefault(slog.New(logHandler))

	// 设置日志输出
	db, err := data.ConnectDB()
	cache := data.ConnectRedis()

	if err != nil {
		panic(err)
	}

	server.AsyncServer(db, cache) // 异步下单
	// server.SyncServer(db, cache) // 同步下单
}
