package eudore

import (
	"bufio"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"
)

const _hex = "0123456789abcdef"

var (
	levels = [][]byte{
		[]byte("DEBUG"),
		[]byte("INFO"),
		[]byte("WARIRNG"),
		[]byte("ERROR"),
		[]byte("FATAL"),
	}
	part1 = []byte(`{"time":"`)
	part2 = []byte(`","level":"`)
	part3 = []byte(`","fields":{`)
	part4 = []byte("\"")
	part5 = []byte(`,"message":"`)
	part6 = []byte("\"}\n")
	part7 = []byte("}\n")
)

type (
	// LoggerStd 标准日志处理实现，将日志输出到标准输出或者文件。
	LoggerStd struct {
		LoggerStdConfig
		out    *bufio.Writer
		ticker *time.Ticker
		pool   sync.Pool
		mu     sync.Mutex
	}
	// LoggerStdConfig 定义LoggerStd配置信息。
	LoggerStdConfig struct {
		Std        bool        `json:"std" alias:"std"`
		Path       string      `json:"path" alias:"path"`
		Level      LoggerLevel `json:"level" alias:"level"`
		TimeFormat string      `json:"timeformat" alias:"timeformat"`
	}
	// 标准日志条目
	entryStd struct {
		level   LoggerLevel
		time    time.Time
		message string
		data    []byte
		logger  *LoggerStd
	}
)

// NewLoggerStd 创建一个标准日志处理器。
func NewLoggerStd(arg interface{}) Logger {
	// 解析配置
	config := LoggerStdConfig{
		TimeFormat: "2006-01-02 15:04:05",
	}
	ConvertTo(arg, &config)

	log := &LoggerStd{
		LoggerStdConfig: config,
	}
	log.pool.New = func() interface{} {
		data := make([]byte, 2048)
		return &entryStd{
			logger: log,
			data:   data[0:0],
		}
	}
	log.initOut()
	return log
}

// initOut 方法初始化输出流。
func (log *LoggerStd) initOut() {
	log.Path = strings.TrimSpace(log.Path)
	if len(log.Path) == 0 {
		log.out = bufio.NewWriter(os.Stdout)
		return
	}
	// init file
	file, err := os.OpenFile(log.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.out = bufio.NewWriter(os.Stdout)
		log.Error(err)
		return
	}
	if log.Std {
		log.out = bufio.NewWriter(io.MultiWriter(os.Stdout, file))
	} else {
		log.out = bufio.NewWriter(file)
	}
}

// SetLevel 方法设置日志输出级别。
func (log *LoggerStd) SetLevel(level LoggerLevel) {
	log.mu.Lock()
	log.Level = level
	log.mu.Unlock()
}

// Sync 方法将缓冲写入到输出流。
func (log *LoggerStd) Sync() error {
	log.mu.Lock()
	err := log.out.Flush()
	log.mu.Unlock()
	return err
}

func (log *LoggerStd) newEntry() *entryStd {
	entry := log.pool.Get().(*entryStd)
	entry.time = time.Now()
	return entry
}

// Debug 方法输出Debug级别日志。
func (log *LoggerStd) Debug(args ...interface{}) {
	log.newEntry().Debug(args...)
}

// Info 方法输出Info级别日志。
func (log *LoggerStd) Info(args ...interface{}) {
	log.newEntry().Info(args...)
}

// Warning 方法输出Warning级别日志。
func (log *LoggerStd) Warning(args ...interface{}) {
	log.newEntry().Warning(args...)
}

// Error 方法输出Error级别日志。
func (log *LoggerStd) Error(args ...interface{}) {
	log.newEntry().Error(args...)
}

// Fatal 方法输出Fatal级别日志。
func (log *LoggerStd) Fatal(args ...interface{}) {
	log.newEntry().Fatal(args...)
}

// Debugf 方法格式化输出Debug级别日志。
func (log *LoggerStd) Debugf(format string, args ...interface{}) {
	log.newEntry().Debugf(format, args...)
}

// Infof 方法格式化输出Info级别日志。
func (log *LoggerStd) Infof(format string, args ...interface{}) {
	log.newEntry().Infof(format, args...)
}

// Warningf 方法格式化输出Warning级别日志。
func (log *LoggerStd) Warningf(format string, args ...interface{}) {
	log.newEntry().Warningf(format, args...)
}

// Errorf 方法格式化输出Error级别日志。
func (log *LoggerStd) Errorf(format string, args ...interface{}) {
	log.newEntry().Errorf(format, args...)
}

// Fatalf 方法格式化输出Fatal级别日志。
func (log *LoggerStd) Fatalf(format string, args ...interface{}) {
	log.newEntry().Fatalf(format, args...)
}

// WithField 方法设置日志属性。
func (log *LoggerStd) WithField(key string, value interface{}) Logout {
	return log.newEntry().WithField(key, value)
}

// WithFields 方法设置多个日志属性。
func (log *LoggerStd) WithFields(fields Fields) Logout {
	return log.newEntry().WithFields(fields)
}

// Debug 方法条目输出Debug级别日志。
func (entry *entryStd) Debug(args ...interface{}) {
	if entry.logger.Level < 1 {
		entry.level = 0
		entry.message = fmt.Sprintln(args...)
		entry.message = entry.message[:len(entry.message)-1]
		entry.logger.mu.Lock()
		entry.writeTo(entry.logger.out)
		entry.logger.mu.Unlock()
		entry.logger.pool.Put(entry)
	}
}

// Info 方法条目输出Info级别日志。
func (entry *entryStd) Info(args ...interface{}) {
	if entry.logger.Level < 2 {
		entry.level = 1
		entry.message = fmt.Sprintln(args...)
		entry.message = entry.message[:len(entry.message)-1]
		entry.logger.mu.Lock()
		entry.writeTo(entry.logger.out)
		entry.logger.mu.Unlock()
		entry.logger.pool.Put(entry)
	}
}

// Warning 方法条目输出Warning级别日志。
func (entry *entryStd) Warning(args ...interface{}) {
	if entry.logger.Level < 3 {
		entry.level = 2
		entry.message = fmt.Sprintln(args...)
		entry.message = entry.message[:len(entry.message)-1]
		entry.logger.mu.Lock()
		entry.writeTo(entry.logger.out)
		entry.logger.mu.Unlock()
		entry.logger.pool.Put(entry)
	}
}

// Error 方法条目输出Error级别日志。
func (entry *entryStd) Error(args ...interface{}) {
	if entry.logger.Level < 4 {
		entry.level = 3
		entry.message = fmt.Sprintln(args...)
		entry.message = entry.message[:len(entry.message)-1]
		entry.logger.mu.Lock()
		entry.writeTo(entry.logger.out)
		entry.logger.mu.Unlock()
		entry.logger.pool.Put(entry)
	}
}

// Fatal 方法条目输出Fatal级别日志。
func (entry *entryStd) Fatal(args ...interface{}) {
	entry.level = 4
	entry.message = fmt.Sprintln(args...)
	entry.message = entry.message[:len(entry.message)-1]
	entry.logger.mu.Lock()
	entry.writeTo(entry.logger.out)
	entry.logger.mu.Unlock()
	entry.logger.pool.Put(entry)
}

// Debugf 方法格式化写入流Debug级别日志
func (entry *entryStd) Debugf(format string, args ...interface{}) {
	if entry.logger.Level < 1 {
		entry.level = 0
		entry.message = fmt.Sprintf(format, args...)
		entry.logger.mu.Lock()
		entry.writeTo(entry.logger.out)
		entry.logger.mu.Unlock()
		entry.logger.pool.Put(entry)
	}
}

// Infof 方法格式写入流出Info级别日志
func (entry *entryStd) Infof(format string, args ...interface{}) {
	if entry.logger.Level < 2 {
		entry.level = 1
		entry.message = fmt.Sprintf(format, args...)
		entry.logger.mu.Lock()
		entry.writeTo(entry.logger.out)
		entry.logger.mu.Unlock()
		entry.logger.pool.Put(entry)
	}
}

// Warningf 方法格式化输出写入流rning级别日志
func (entry *entryStd) Warningf(format string, args ...interface{}) {
	if entry.logger.Level < 3 {
		entry.level = 2
		entry.message = fmt.Sprintf(format, args...)
		entry.logger.mu.Lock()
		entry.writeTo(entry.logger.out)
		entry.logger.mu.Unlock()
		entry.logger.pool.Put(entry)
	}
}

// Errorf 方法格式化写入流Error级别日志
func (entry *entryStd) Errorf(format string, args ...interface{}) {
	if entry.logger.Level < 4 {
		entry.level = 3
		entry.message = fmt.Sprintf(format, args...)
		entry.logger.mu.Lock()
		entry.writeTo(entry.logger.out)
		entry.logger.mu.Unlock()
		entry.logger.pool.Put(entry)
	}
}

// Fatalf 方法格式化写入流Fatal级别日志
func (entry *entryStd) Fatalf(format string, args ...interface{}) {
	entry.level = 4
	entry.message = fmt.Sprintf(format, args...)
	entry.logger.mu.Lock()
	entry.writeTo(entry.logger.out)
	entry.logger.mu.Unlock()
	entry.logger.pool.Put(entry)
}

// WithFields 方法设置多个条目属性。
func (entry *entryStd) WithFields(fields Fields) Logout {
	for k, v := range fields {
		entry.WithField(k, v)
	}
	return entry
}

// WithField 方法设置一个日志属性。
func (entry *entryStd) WithField(key string, value interface{}) Logout {
	if key == "time" {
		var ok bool
		entry.time, ok = value.(time.Time)
		if ok {
			return entry
		}
	}
	entry.data = append(entry.data, '"')
	entry.data = append(entry.data, key...)
	entry.data = append(entry.data, '"', ':')
	entry.WriteValue(value)
	entry.data = append(entry.data, ',')
	return entry
}

// WriteValue 方法写入值。
func (entry *entryStd) WriteValue(value interface{}) {
	iValue := reflect.ValueOf(value)
	entry.writeReflect(iValue)
}

// writeReflect 方法写入值。
func (entry *entryStd) writeReflect(iValue reflect.Value) {
	if iValue.Kind() == reflect.Invalid {
		entry.data = append(entry.data, '"', '"')
		return
	}
	// 检查接口
	switch val := iValue.Interface().(type) {
	case json.Marshaler:
		body, err := val.MarshalJSON()
		entry.data = append(entry.data, '"')
		if err == nil {
			entry.writeBytes(body)
		} else {
			entry.writeString(err.Error())
		}
		entry.data = append(entry.data, '"')
		return
	case encoding.TextMarshaler:
		body, err := val.MarshalText()
		entry.data = append(entry.data, '"')
		if err == nil {
			entry.writeBytes(body)
		} else {
			entry.writeString(err.Error())
		}
		entry.data = append(entry.data, '"')
		return
	case fmt.Stringer:
		entry.data = append(entry.data, '"')
		entry.writeString(val.String())
		entry.data = append(entry.data, '"')
		return
	}
	// 写入类型
	switch iValue.Kind() {
	case reflect.Bool:
		entry.data = strconv.AppendBool(entry.data, iValue.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		entry.data = strconv.AppendInt(entry.data, iValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		entry.data = strconv.AppendUint(entry.data, iValue.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		entry.data = strconv.AppendFloat(entry.data, iValue.Float(), 'f', -1, 64)
	case reflect.Complex64, reflect.Complex128:
		val := iValue.Complex()
		r, i := float64(real(val)), float64(imag(val))
		entry.data = append(entry.data, '"')
		entry.data = strconv.AppendFloat(entry.data, r, 'f', -1, 64)
		entry.data = append(entry.data, '+')
		entry.data = strconv.AppendFloat(entry.data, i, 'f', -1, 64)
		entry.data = append(entry.data, 'i')
		entry.data = append(entry.data, '"')
	case reflect.String:
		entry.data = append(entry.data, '"')
		entry.writeString(iValue.String())
		entry.data = append(entry.data, '"')
	case reflect.Array, reflect.Slice:
		entry.data = append(entry.data, '[')
		for i := 0; i < iValue.Len(); i++ {
			entry.writeReflect(iValue.Index(i))
			entry.data = append(entry.data, ',')
		}
		entry.data[len(entry.data)-1] = ']'
	case reflect.Map:
		entry.data = append(entry.data, '{')
		for _, key := range iValue.MapKeys() {
			entry.writeReflect(key)
			entry.data = append(entry.data, ':')
			entry.writeReflect(iValue.MapIndex(key))
			entry.data = append(entry.data, ',')
		}
		entry.data[len(entry.data)-1] = '}'
	case reflect.Struct:
		entry.data = append(entry.data, '{')
		iType := iValue.Type()
		for i := 0; i < iValue.NumField(); i++ {
			if iValue.Field(i).CanInterface() {
				entry.writeString(iType.Field(i).Name)
				entry.data = append(entry.data, ':')
				entry.writeReflect(iValue.Field(i))
				entry.data = append(entry.data, ',')
			}
		}
		entry.data[len(entry.data)-1] = '}'
	case reflect.Ptr, reflect.Interface:
		entry.writeReflect(iValue.Elem())
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		entry.data = append(entry.data, '0', 'x')
		entry.data = strconv.AppendUint(entry.data, uint64(iValue.Pointer()), 16)
	}
}

// writeString 方法安全写入字符串。
func (entry *entryStd) writeString(s string) {
	for i := 0; i < len(s); {
		if entry.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if entry.tryAddRuneError(r, size) {
			i++
			continue
		}
		entry.data = append(entry.data, s[i:i+size]...)
		i += size
	}
}

// writeBytes 方法安全写入[]byte的字符串数据。
func (entry *entryStd) writeBytes(s []byte) {
	for i := 0; i < len(s); {
		if entry.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if entry.tryAddRuneError(r, size) {
			i++
			continue
		}
		entry.data = append(entry.data, s[i:i+size]...)
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (entry *entryStd) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		entry.data = append(entry.data, b)
		return true
	}
	switch b {
	case '\\', '"':
		entry.data = append(entry.data, '\\')
		entry.data = append(entry.data, b)
	case '\n':
		entry.data = append(entry.data, '\\')
		entry.data = append(entry.data, 'n')
	case '\r':
		entry.data = append(entry.data, '\\')
		entry.data = append(entry.data, 'r')
	case '\t':
		entry.data = append(entry.data, '\\')
		entry.data = append(entry.data, 't')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		entry.data = append(entry.data, `\u00`...)
		entry.data = append(entry.data, _hex[b>>4])
		entry.data = append(entry.data, _hex[b&0xF])
	}
	return true
}

func (entry *entryStd) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		entry.data = append(entry.data, `\ufffd`...)
		return true
	}
	return false
}

// writeTo 将数据写入到输出。
func (entry *entryStd) writeTo(w io.Writer) {
	w.Write(part1)
	timestr := time.Now().Format(entry.logger.TimeFormat)
	w.Write(*(*[]byte)(unsafe.Pointer(&timestr)))
	w.Write(part2)
	w.Write(levels[entry.level])

	if len(entry.data) > 1 {
		w.Write(part3)
		entry.data[len(entry.data)-1] = '}'
		w.Write(entry.data)
		entry.data = entry.data[0:0]
	} else {
		w.Write(part4)
	}

	if len(entry.message) > 0 {
		w.Write(part5)
		entry.writeString(entry.message)
		w.Write(entry.data)
		entry.data = entry.data[0:0]
		w.Write(part6)
	} else {
		w.Write(part7)
	}
}
