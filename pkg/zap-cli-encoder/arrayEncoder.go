package zap_cli_encoder

import (
	"fmt"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type bufferArrayEncoder struct {
	buffer *buffer.Buffer
}

func (t *bufferArrayEncoder) AppendComplex128(v complex128) {
	r, i := real(v), imag(v)
	t.buffer.AppendFloat(r, 64)
	t.buffer.AppendByte('+')
	t.buffer.AppendFloat(i, 64)
	t.buffer.AppendByte('i')
}

func (t *bufferArrayEncoder) AppendComplex64(v complex64) {
	//noinspection GoRedundantConversion
	t.AppendComplex128(complex128(v))
}

func (t *bufferArrayEncoder) AppendArray(v zapcore.ArrayMarshaler) error {
	enc := &bufferArrayEncoder{}
	err := v.MarshalLogArray(enc)
	_, _ = fmt.Fprintf(t.buffer, "%v", enc.buffer)
	return err
}

func (t *bufferArrayEncoder) AppendObject(v zapcore.ObjectMarshaler) error {
	m := zapcore.NewMapObjectEncoder()
	err := v.MarshalLogObject(m)
	_, _ = fmt.Fprintf(t.buffer, "%v", m.Fields)
	return err
}

func (t *bufferArrayEncoder) AppendReflected(v interface{}) error {
	_, _ = fmt.Fprintf(t.buffer, "%v", v)
	return nil
}

func (t *bufferArrayEncoder) AppendBool(v bool) {
	t.buffer.AppendBool(v)
}

func (t *bufferArrayEncoder) AppendByteString(v []byte) {
	t.buffer.AppendString(string(v))
}

func (t *bufferArrayEncoder) AppendDuration(v time.Duration) {
	t.AppendString(v.String())
}

func (t *bufferArrayEncoder) AppendFloat64(v float64) { t.buffer.AppendFloat(v, 64) }
func (t *bufferArrayEncoder) AppendFloat32(v float32) { t.buffer.AppendFloat(float64(v), 32) }
func (t *bufferArrayEncoder) AppendInt(v int)         { t.buffer.AppendInt(int64(v)) }
func (t *bufferArrayEncoder) AppendInt64(v int64)     { t.buffer.AppendInt(v) }
func (t *bufferArrayEncoder) AppendInt32(v int32)     { t.buffer.AppendInt(int64(v)) }
func (t *bufferArrayEncoder) AppendInt16(v int16)     { t.buffer.AppendInt(int64(v)) }
func (t *bufferArrayEncoder) AppendInt8(v int8)       { t.buffer.AppendInt(int64(v)) }
func (t *bufferArrayEncoder) AppendString(v string)   { t.buffer.AppendString(v) }
func (t *bufferArrayEncoder) AppendTime(v time.Time)  { t.buffer.AppendString(v.String()) }
func (t *bufferArrayEncoder) AppendUint(v uint)       { t.buffer.AppendUint(uint64(v)) }
func (t *bufferArrayEncoder) AppendUint64(v uint64)   { t.buffer.AppendUint(v) }
func (t *bufferArrayEncoder) AppendUint32(v uint32)   { t.buffer.AppendUint(uint64(v)) }
func (t *bufferArrayEncoder) AppendUint16(v uint16)   { t.buffer.AppendUint(uint64(v)) }
func (t *bufferArrayEncoder) AppendUint8(v uint8)     { t.buffer.AppendUint(uint64(v)) }
func (t *bufferArrayEncoder) AppendUintptr(v uintptr) { t.buffer.AppendUint(uint64(v)) }
