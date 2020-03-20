package ppp

import (
	"fmt"
	"github.com/panjjo/ppp/db"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

// DBPool 数据库连接池实例
var DBPool *db.Pool

// NewLogger 创建log 记录
func NewLogger(str string) {
	level, _ := logrus.ParseLevel(str)
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "01-02 15:04:05"})
	logrus.AddHook(newlogHook())
}

// NewDBPool 创建数据库连接池
func NewDBPool(config *db.Config) {
	DBPool = db.GetPool(config)
}

func newlogHook(levels ...logrus.Level)logrus.Hook{
	hook := &loghook{
		skip:8,
		field:"file",
		levels: levels,
	}
	if len(levels) == 0 {
		hook.levels = logrus.AllLevels
	}
	return hook
}

type loghook struct {
	skip int
	field string
	levels []logrus.Level
}

func (h *loghook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (h *loghook) Fire(e *logrus.Entry) error {
	e.Data["file"] = findCaller(8)
	return nil
}
func findCaller(skip int) string {
	file := ""
	line := 0
	for i := 0; i < 10; i++ {
		file, line = getCaller(skip + i)
		if !strings.HasPrefix(file, "logrus") {
			break
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}
func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	n := 0
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return file, line
}
