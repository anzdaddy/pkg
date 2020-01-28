package log

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arr-ai/frozen"
	"github.com/sirupsen/logrus"
)

const keyFields = "_fields"

type standardLogger struct {
	internal *logrus.Logger
	fields   frozen.Map
}

type standardFormat struct{}
type jsonFormat struct{}

func (sf *standardFormat) Format(entry *logrus.Entry) ([]byte, error) {
	message := strings.Builder{}
	message.WriteString(entry.Time.Format(time.RFC3339Nano))
	message.WriteByte(' ')

	if entry.Data[keyFields] != nil && entry.Data[keyFields].(frozen.Map).Count() != 0 {
		message.WriteString(getFormattedField(entry.Data[keyFields].(frozen.Map)))
		message.WriteByte(' ')
	}

	message.WriteString(strings.ToUpper(entry.Level.String()))
	message.WriteByte(' ')

	if entry.Message != "" {
		message.WriteString(entry.Message)
		message.WriteByte(' ')
	}

	// TODO: add codelinker's message here
	message.WriteByte('\n')
	return []byte(message.String()), nil
}

func (jf *jsonFormat) Format(entry *logrus.Entry) ([]byte, error) {
	jsonFile := make(map[string]interface{})
	jsonFile["timestamp"] = entry.Time.Format(time.RFC3339Nano)
	jsonFile["message"] = entry.Message
	jsonFile["level"] = strings.ToUpper(entry.Level.String())
	if entry.Data[keyFields] != nil && entry.Data[keyFields].(frozen.Map).Count() != 0 {
		fields := make(map[string]interface{})
		for i := entry.Data[keyFields].(frozen.Map).Range(); i.Next(); {
			fields[i.Key().(string)] = i.Value()
		}
		jsonFile["fields"] = fields
	}
	data, err := json.Marshal(jsonFile)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), err
}

// NewStandardLogger returns a logger with logrus standard logger as the internal logger
func NewStandardLogger() Logger {
	logger := logrus.New()
	logger.SetFormatter(&standardFormat{})

	// makes sure that it always logs every level
	logger.SetLevel(logrus.DebugLevel)

	// explicitly set it to os.Stderr
	logger.SetOutput(os.Stderr)

	return &standardLogger{internal: logger}
}

func (sl *standardLogger) Debug(args ...interface{}) {
	sl.setInfo().Debug(args...)
}

func (sl *standardLogger) Debugf(format string, args ...interface{}) {
	sl.setInfo().Debugf(format, args...)
}

func (sl *standardLogger) Info(args ...interface{}) {
	sl.setInfo().Info(args...)
}

func (sl *standardLogger) Infof(format string, args ...interface{}) {
	sl.setInfo().Infof(format, args...)
}

func (sl *standardLogger) PutFields(fields frozen.Map) Logger {
	sl.fields = fields
	return sl
}

func (sl *standardLogger) SetConfig(configs frozen.Map) Logger {
	for i := configs.Range(); i.Next(); {
		switch i.Key().(int) {
		case formatter:
			sl.setFormatter(i.Value().(int))
		default:
			panic("unknown configuration")
		}
	}
	return sl
}

func (sl *standardLogger) setFormatter(formatterType int) {
	switch formatterType {
	case jsonFormatter:
		sl.internal.SetFormatter(&jsonFormat{})
	default:
		sl.internal.SetFormatter(&standardFormat{})
	}
}

func (sl *standardLogger) Copy() Logger {
	return &standardLogger{sl.getCopiedInternalLogger(), sl.fields}
}

func (sl *standardLogger) setInfo() *logrus.Entry {
	// TODO: set linker here

	return sl.internal.WithFields(logrus.Fields{
		keyFields: sl.fields,
	})
}

func getFormattedField(fields frozen.Map) string {
	if fields.Count() == 0 {
		return ""
	}

	formattedFields := strings.Builder{}
	i := fields.Range()
	i.Next()
	formattedFields.WriteString(fmt.Sprintf("%v=%v", i.Key(), i.Value()))
	for i.Next() {
		formattedFields.WriteString(fmt.Sprintf(" %v=%v", i.Key(), i.Value()))
	}
	return formattedFields.String()
}

func (sl *standardLogger) getCopiedInternalLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(sl.internal.Formatter)
	logger.SetLevel(sl.internal.Level)
	logger.SetOutput(sl.internal.Out)

	return logger
}