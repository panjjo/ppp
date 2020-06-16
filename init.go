package ppp

import (
	"github.com/panjjo/ppp/db"
	"github.com/sirupsen/logrus"
)

var DBClient *db.Client

// NewLogger 创建log 记录
func NewLogger(level string) {
	l, e := logrus.ParseLevel(level)
	if e != nil {
		l = logrus.DebugLevel
	}
	logrus.SetLevel(l)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "01-02 15:04:05"})
}

// NewDB 初始化数据
func NewDB(config db.Config) {
	db.InitDB(config)
	DBClient = db.NewClient().SetDB(config.DBName)
}
func init() {
	logrus.SetLevel(logrus.DebugLevel)
}
