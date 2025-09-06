package logger

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/widiskel/poseidon-voice-bot/internal/model"
	"github.com/widiskel/poseidon-voice-bot/internal/utils"
	"github.com/widiskel/poseidon-voice-bot/internal/utils/spinner"
)

var (
	fileLogger *log.Logger
	once       sync.Once
	logFile    *os.File
)

func Init(path string) error {
	var err error
	once.Do(func() {
		os.Remove(path)
		if err = os.MkdirAll(dirOf(path), 0o755); err != nil {
			return
		}
		logFile, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return
		}
		fileLogger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	})
	return err
}

func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

func dirOf(path string) string {
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "."
	}
	return path[:i]
}

type ClassLogger struct {
	class   string
	session *model.Session
}

func NewLogger(v interface{}, session *model.Session) *ClassLogger {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &ClassLogger{class: t.Name(), session: session}
}

func NewNamed(name string, session *model.Session) *ClassLogger {
	return &ClassLogger{class: name, session: session}
}

func (l *ClassLogger) Log(msg string, durationMs ...int) {
	totalDuration := 300 * time.Millisecond
	if len(durationMs) > 0 {
		totalDuration = time.Duration(durationMs[0]) * time.Millisecond
	}

	if fileLogger != nil {
		funcName := callerFunc(2)
		fileLogger.Printf("[%s][%s] %s", l.class, funcName, msg)
	}

	if totalDuration > 0 {
		interval := 1 * time.Second

		for remaining := totalDuration; remaining > 0; remaining -= interval {
			spinner.UpdateStatus(*l.session, msg, remaining)

			sleepTime := interval
			if remaining < interval {
				sleepTime = remaining
			}
			time.Sleep(sleepTime)
		}
	}

	spinner.UpdateStatus(*l.session, msg, 0)
}

func (l *ClassLogger) JustLog(msg string) {

	if fileLogger != nil {
		funcName := callerFunc(2)
		fileLogger.Printf("[%s][%s] %s", l.class, funcName, msg)
	}
}

func (l *ClassLogger) LogObject(msg string, obj interface{}) {
	if fileLogger != nil {
		formattedString, err := utils.FormatObject(obj)
		if err != nil {
			l.JustLog(fmt.Sprintf("Error formatting object: %v", err))
			return
		}
		l.JustLog(fmt.Sprintf("%s : \n%v", msg, formattedString))
	}
}

func callerFunc(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	parts := strings.Split(fn.Name(), ".")
	return parts[len(parts)-1]
}
