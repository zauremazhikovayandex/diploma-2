// Package message содержит типы и константы сообщений логгера.
package message

// Уровни логирования, поддерживаемые подсистемой логгера.
const (
	PanicLevel = "panic"
	FatalLevel = "fatal"
	ErrorLevel = "error"
	WarnLevel  = "warn"
	InfoLevel  = "info"
	DebugLevel = "debug"
	TraceLevel = "trace"
)

// LogMessage описывает структуру сообщения, передаваемого драйверу логгера.
type LogMessage struct {
	Message     string                  `json:"message"`
	FullMessage *string                 `json:"full_message,omitempty"`
	Host        *string                 `json:"host,omitempty"`
	Timestamp   *float64                `json:"timestamp,omitempty"`
	Facility    *string                 `json:"facility,omitempty"`
	Extra       *map[string]interface{} `json:"extra,omitempty"`
}
