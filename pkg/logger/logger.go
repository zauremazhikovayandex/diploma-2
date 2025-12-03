// Package logger предоставляет интерфейсы и реализации логирования приложения.
package logger

import (
	"diploma-2/pkg/logger/drivers"
	"diploma-2/pkg/logger/message"
	"net/http"
	"time"
)

// Log — глобальный активный логгер приложения, реализующий интерфейс Interface.
var Log Interface

// Interface описывает минимальный набор методов, поддерживаемых драйвером логгера.
type Interface interface {
	Debug(msg *message.LogMessage)
	Info(msg *message.LogMessage)
	Warn(msg *message.LogMessage)
	Error(msg *message.LogMessage)
	Fatal(msg *message.LogMessage)
	Panic(msg *message.LogMessage)
}

// New инициализирует глобальный логгер и access-логгер с заданным уровнем.
func New(level string) Interface {
	Log = drivers.MakeStdoutLogger(level)
	Logging = &Writer{}
	return Log
}

// Logging — глобальный access-логгер HTTP-запросов.
var Logging LogWriter

// LogWriter описывает интерфейс для записи сводной информации о HTTP-запросах.
type LogWriter interface {
	WriteToLog(timeStart time.Time, originalURL string, requestType string, responseCode int, responseBody string)
}

// Writer реализует LogWriter и пишет информацию о запросах через глобальный Log.
type Writer struct{}

// WriteToLog формирует и записывает в лог информацию о HTTP-запросе и ответе.
func (l *Writer) WriteToLog(timeStart time.Time, originalURL string, requestType string,
	responseCode int, responseBody string) {
	timeEnd := time.Now()
	duration := timeEnd.Sub(timeStart)

	requestInfo := make(map[string]interface{})
	requestInfo["duration"] = duration
	requestInfo["uri"] = originalURL
	requestInfo["request_type"] = requestType
	requestInfo["response_code"] = responseCode
	requestInfo["response_body"] = responseBody

	Log.Info(&message.LogMessage{Message: "REQUEST INFO: %s",
		Extra: &requestInfo,
	})
}

// RequestLogger — HTTP-middleware, логирующий каждый запрос и ответ.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		Logging.WriteToLog(timeStart, r.RequestURI, r.Method, lrw.statusCode, http.StatusText(lrw.statusCode))
	})
}

// loggingResponseWriter — обертка над http.ResponseWriter, запоминающая статус ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader сохраняет HTTP-статус в обертке и проксирует вызов исходному ResponseWriter.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
