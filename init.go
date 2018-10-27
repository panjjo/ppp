package ppp

import (
	"github.com/panjjo/ppp/db"

	"github.com/panjjo/log4go"
)

// DBPool 数据库连接池实例
var DBPool *db.Pool

// Log  日志处理器
var Log *log4go.Logger

// NewLogger 创建log 记录
func NewLogger(level string) {
	Log = log4go.NewLogger(level)
}

// NewDBPool 创建数据库连接池
func NewDBPool(config *db.Config) {
	DBPool = db.GetPool(config)
}

func init() {
	Log = log4go.NewLogger("debug")
}
