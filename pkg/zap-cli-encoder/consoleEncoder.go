package zap_cli_encoder

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var linePool = buffer.NewPool()
var spacePrefix = []byte(" ")

const infoLevelIndicator = "  • "
const errorLevelIndicator = "  ⨯ "
var levelIndicatorRuneCount = utf8.RuneCountInString(infoLevelIndicator)

type consoleEncoder struct {
	encoderConfig zapcore.EncoderConfig
	colored       bool

	// With will write fields to buffer, so, EncodeEntry will simply add this buffer to result
	// for such entries stacked layout not supported for now
	fieldNames  []string
	fieldValues [][]byte
}

func NewConsoleEncoder(encoderConfig zapcore.EncoderConfig, colored bool) zapcore.Encoder {
	return &consoleEncoder{
		encoderConfig: encoderConfig,
		colored:       colored,
	}
}

func (t *consoleEncoder) Clone() zapcore.Encoder {
	return &consoleEncoder{encoderConfig: t.encoderConfig, colored: t.colored}
}

//noinspection ALL
const (
	Black uint8 = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

func (t *consoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	levelColor, levelIndicator := getLevelColorAndIndicator(&entry)
	line := linePool.Get()

	// <space><space><indicator char><space>
	if t.colored {
		_, _ = fmt.Fprintf(line, "\x1b[%dm%s\x1b[0m", levelColor, levelIndicator)
	} else {
		line.AppendString(levelIndicator)
	}

	if t.colored && entry.Level >= zapcore.ErrorLevel {
		_, _ = fmt.Fprintf(line, "\x1b[%dm%s\x1b[0m", levelColor, entry.Message)
	} else {
		line.AppendString(entry.Message)
	}

	paddingSizeAfterMessage := max(2, 16-len(entry.Message))
	line.AppendString(strings.Repeat(" ", paddingSizeAfterMessage))
	// n can be used because ASCII only (so, byte count equals to char count)

	extraFieldNumber := len(t.fieldNames)
	fieldNameAndValueList := make([]*buffer.Buffer, len(fields) + extraFieldNumber)
	totalLength := 0

	arrayEncoder := &bufferArrayEncoder{buffer: linePool.Get()}

	totalLength += t.encodeExtraFields(levelColor, &fieldNameAndValueList)

fieldLoop:
	for index, field := range fields {
		if field.Type == zapcore.SkipType {
			continue
		}

		buf := linePool.Get()
		if t.colored {
			_, _ = fmt.Fprintf(buf, "\x1b[%dm%s\x1b[0m", levelColor, field.Key)
		} else {
			buf.AppendString(field.Key)
		}
		buf.AppendString("=")

		var v string

		switch field.Type {
		case zapcore.ArrayMarshalerType:
			arrayEncoder.buffer.Reset()

			err := field.Interface.(zapcore.ArrayMarshaler).MarshalLogArray(arrayEncoder)
			v = arrayEncoder.buffer.String()
			if err != nil {
				return nil, err
			}

		case zapcore.ObjectMarshalerType, zapcore.BinaryType, zapcore.Complex128Type, zapcore.Complex64Type, zapcore.ReflectType, zapcore.NamespaceType:
			return nil, fmt.Errorf("unsupported field type: %v", field)
		case zapcore.BoolType:
			if field.Integer == 1 {
				v = "true"
			} else {
				v = "false"
			}
		case zapcore.StringerType:
			v = field.Interface.(fmt.Stringer).String()
		case zapcore.DurationType:
			v = time.Duration(field.Integer).String()
		case zapcore.Float64Type:
			v = strconv.FormatFloat(math.Float64frombits(uint64(field.Integer)), 'f', 6, 64)
		case zapcore.Float32Type:
			v = strconv.FormatFloat(float64(math.Float32frombits(uint32(field.Integer))), 'f', 6, 32)
		case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type, zapcore.UintptrType:
			v = strconv.FormatInt(field.Integer, 10)
		case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
			v = strconv.FormatUint(uint64(field.Integer), 10)
		case zapcore.ByteStringType:
			v = string(field.Interface.([]byte))
		case zapcore.StringType:
			v = field.String
		case zapcore.TimeType:
			if field.Interface != nil {
				v = time.Unix(0, field.Integer).In(field.Interface.(*time.Location)).String()
			} else {
				// Fall back to UTC if location is nil.
				v = time.Unix(0, field.Integer).String()
			}
		case zapcore.ErrorType:
			v = fmt.Sprintf("%+v", field.Interface.(error))
		case zapcore.SkipType:
			continue fieldLoop
		default:
			return nil, fmt.Errorf("unknown field type: %v", field)
		}

		totalLength += len(field.Key) + 1 + len(v)

		if totalLength > 180 {
			appendPaddedString(v, buf)
		} else {
			buf.AppendString(v)
		}

		fieldNameAndValueList[index+extraFieldNumber] = buf
	}

	fieldOffset := levelIndicatorRuneCount + utf8.RuneCountInString(entry.Message) + paddingSizeAfterMessage
	fieldPrefix := getFieldPrefix(totalLength > 180, fieldOffset)

	writtenIndex := 0
	for _, v := range fieldNameAndValueList {
		if v.Len() == 0 {
			continue
		}

		if writtenIndex > 0 {
			_, _ = line.Write(fieldPrefix)
		}

		_, _ = line.Write(v.Bytes())
		v.Free()

		writtenIndex++
	}

	line.AppendString("\n")
	return line, nil
}

func getFieldPrefix(stacked bool, fieldOffset int) []byte {
	if !stacked {
		return spacePrefix
	}

	fieldPrefix := make([]byte, fieldOffset+1)
	fieldPrefix[0] = 10
	for i := 1; i < len(fieldPrefix); i++ {
		fieldPrefix[i] = 32
	}
	return fieldPrefix
}

func appendPaddedString(v string, buf *buffer.Buffer) {
	index := 0
	for {
		index = strings.IndexByte(v, '\n')
		if index < 0 {
			buf.AppendString(v)
			break
		}

		buf.AppendString(v[:index+1])
		for i := 0; i < levelIndicatorRuneCount; i++ {
			buf.AppendByte(' ')
		}

		v = v[index+1:]
	}
}

func (t *consoleEncoder) encodeExtraFields(levelColor uint8, fieldNameAndValueList *[]*buffer.Buffer) int {
	addedLength := 0
	for index, key := range t.fieldNames {
		v := t.fieldValues[index]
		addedLength += len(key) + 1 + len(v)

		buf := linePool.Get()
		if t.colored {
			_, _ = fmt.Fprintf(buf, "\x1b[%dm%s\x1b[0m", levelColor, key)
		} else {
			buf.AppendString(key)
		}
		buf.AppendString("=")
		_, _ = buf.Write(v)

		(*fieldNameAndValueList)[index] = buf
	}
	return addedLength
}

func getLevelColorAndIndicator(entry *zapcore.Entry) (uint8, string) {
	var levelColor uint8
	var levelIndicator string
	if entry.Level < zapcore.ErrorLevel {
		levelIndicator = infoLevelIndicator
	} else {
		levelIndicator = errorLevelIndicator
	}
	switch entry.Level {
	case zapcore.DebugLevel:
		levelColor = White
	case zapcore.InfoLevel:
		levelColor = Blue
	case zapcore.WarnLevel:
		levelColor = Yellow
	default:
		levelColor = Red
	}
	return levelColor, levelIndicator
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (t *consoleEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	arrayEncoder := &bufferArrayEncoder{buffer: linePool.Get()}
	defer arrayEncoder.buffer.Free()

	err := marshaler.MarshalLogArray(arrayEncoder)
	if err != nil {
		return err
	}

	t.fieldNames = append(t.fieldNames, key)
	t.fieldValues = append(t.fieldValues, arrayEncoder.buffer.Bytes())
	return nil
}

func (t *consoleEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	panic("implement me")
}

func (t *consoleEncoder) AddBinary(key string, value []byte) {
	t.fieldNames = append(t.fieldNames, key)

	enc := base64.StdEncoding
	buf := make([]byte, enc.EncodedLen(len(value)))
	enc.Encode(buf, value)
	t.fieldValues = append(t.fieldValues, buf)
}

func (t *consoleEncoder) AddByteString(key string, value []byte) {
	t.fieldNames = append(t.fieldNames, key)
	t.fieldValues = append(t.fieldValues, value)
}

func (t *consoleEncoder) AddBool(key string, value bool) {
	// problem is that to encode, we need to know color, but color is not yet known
	t.fieldNames = append(t.fieldNames, key)
	if value {
		t.fieldValues = append(t.fieldValues, []byte("true"))
	} else {
		t.fieldValues = append(t.fieldValues, []byte("false"))
	}
}

func (t *consoleEncoder) AddComplex128(key string, value complex128) {
	panic("implement me")
}

func (t *consoleEncoder) AddComplex64(key string, value complex64) {
	panic("implement me")
}

func (t *consoleEncoder) AddDuration(key string, value time.Duration) {
	t.fieldNames = append(t.fieldNames, key)
	t.fieldValues = append(t.fieldValues, []byte(value.String()))
}

func (t *consoleEncoder) AddFloat64(key string, value float64) {
	panic("implement me")
}

func (t *consoleEncoder) AddFloat32(key string, value float32) {
	panic("implement me")
}

func (t *consoleEncoder) AddInt(key string, value int) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddInt64(key string, value int64) {
	t.fieldNames = append(t.fieldNames, key)
	t.fieldValues = append(t.fieldValues, []byte(strconv.FormatInt(value, 10)))
}

func (t *consoleEncoder) AddInt32(key string, value int32) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddInt16(key string, value int16) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddInt8(key string, value int8) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddString(key, value string) {
	t.fieldNames = append(t.fieldNames, key)
	t.fieldValues = append(t.fieldValues, []byte(value))
}

func (t *consoleEncoder) AddTime(key string, value time.Time) {
	t.AddString(key, value.String())
}

func (t *consoleEncoder) AddUint(key string, value uint) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddUint64(key string, value uint64) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddUint32(key string, value uint32) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddUint16(key string, value uint16) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddUint8(key string, value uint8) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddUintptr(key string, value uintptr) {
	t.AddInt64(key, int64(value))
}

func (t *consoleEncoder) AddReflected(key string, value interface{}) error {
	panic("implement me")
}

func (t *consoleEncoder) OpenNamespace(key string) {
	panic("implement me")
}