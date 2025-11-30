package less_go

type LogListener interface {
	Error(msg any)
	Warn(msg any)
	Info(msg any)
	Debug(msg any)
}

type LogListenerPartial map[string]func(msg any)

type Logger struct {
	listeners []any
}

func NewLogger() *Logger {
	return &Logger{
		listeners: make([]any, 0),
	}
}

func (l *Logger) Error(msg any) {
	l.fireEvent("Error", msg)
}

func (l *Logger) Warn(msg any) {
	l.fireEvent("Warn", msg)
}

func (l *Logger) Info(msg any) {
	l.fireEvent("Info", msg)
}

func (l *Logger) Debug(msg any) {
	l.fireEvent("Debug", msg)
}

func (l *Logger) AddListener(listener any) {
	l.listeners = append(l.listeners, listener)
}

func (l *Logger) RemoveListener(listener any) {
	for i := 0; i < len(l.listeners); i++ {
		if l.listeners[i] == listener {
			l.listeners = append(l.listeners[:i], l.listeners[i+1:]...)
			return
		}
	}
}

func (l *Logger) fireEvent(eventType string, msg any) {
	for i := 0; i < len(l.listeners); i++ {
		listener := l.listeners[i]

		switch v := listener.(type) {
		case LogListener:
			switch eventType {
			case "Error":
				v.Error(msg)
			case "Warn":
				v.Warn(msg)
			case "Info":
				v.Info(msg)
			case "Debug":
				v.Debug(msg)
			}
		case LogListenerPartial:
			if logFunction, exists := v[eventType]; exists && logFunction != nil {
				logFunction(msg)
			}
		case map[string]func(msg any):
			if logFunction, exists := v[eventType]; exists && logFunction != nil {
				logFunction(msg)
			}
		}
	}
}

func (l *Logger) GetListeners() []any {
	result := make([]any, len(l.listeners))
	copy(result, l.listeners)
	return result
}

var DefaultLogger = NewLogger()

func Error(msg any) {
	DefaultLogger.Error(msg)
}

func Warn(msg any) {
	DefaultLogger.Warn(msg)
}

func Info(msg any) {
	DefaultLogger.Info(msg)
}

func Debug(msg any) {
	DefaultLogger.Debug(msg)
}

func AddListener(listener any) {
	DefaultLogger.AddListener(listener)
}

func RemoveListener(listener any) {
	DefaultLogger.RemoveListener(listener)
}