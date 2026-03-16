package logger

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

// CustomTextFormatter formats logs into text
type CustomTextFormatter struct {
	DisableColors  bool
	DisablePadding bool
}

// Format renders a single log entry
func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestampFormat := time.RFC3339
	f.printFormatted(b, entry, keys, data, timestampFormat)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *CustomTextFormatter) printFormatted(b *bytes.Buffer, entry *logrus.Entry, keys []string, data logrus.Fields, timestampFormat string) {
	levelText := strings.ToUpper(entry.Level.String())[:4]

	var formattedLevel string
	if f.DisableColors {
		formattedLevel = levelText
	} else {
		levelColor := f.getLevelColor(entry.Level)
		formattedLevel = fmt.Sprintf("\x1b[%dm%s\x1b[0m", levelColor, levelText)
	}

	paddingWidth := 65
	if f.DisablePadding {
		paddingWidth = 0
	}

	paddingFormat := fmt.Sprintf("%%s [%%s] %%-%ds ", paddingWidth)

	fmt.Fprintf(b, paddingFormat, entry.Time.Format(timestampFormat), formattedLevel, entry.Message)

	for _, k := range keys {
		v := data[k]
		if f.DisableColors {
			fmt.Fprintf(b, " | %s=", k)
		} else {
			levelColor := f.getLevelColor(entry.Level)
			fmt.Fprintf(b, " | \x1b[%dm%s\x1b[0m=", levelColor, k)
		}
		f.appendValue(b, v)
	}
}

func (f *CustomTextFormatter) getLevelColor(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		return gray
	case logrus.WarnLevel:
		return yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return red
	case logrus.InfoLevel:
		return blue
	default:
		return blue
	}
}

func (f *CustomTextFormatter) needsQuoting(text string) bool {
	for _, ch := range text {
		if (ch < 'a' || ch > 'z') &&
			(ch < 'A' || ch > 'Z') &&
			(ch < '0' || ch > '9') &&
			ch != '-' && ch != '.' && ch != '_' && ch != '/' && ch != '@' && ch != '^' && ch != '+' {
			return true
		}
	}
	return false
}

func (f *CustomTextFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		fmt.Fprintf(b, "%q", stringVal)
	}
}
