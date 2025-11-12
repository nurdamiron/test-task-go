package transport

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Logger struct {
	logger *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log("info", msg, fields)
}

func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log("error", msg, fields)
}

func (l *Logger) log(level, msg string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}

	for k, v := range fields {
		entry[k] = v
	}

	data, _ := json.Marshal(entry)
	l.logger.Println(string(data))
}
