//go:build !cgo

package speex

import "errors"

const (
	ModeNarrowband    = "nb"
	ModeWideband      = "wb"
	ModeUltraWide     = "uwb"
	DefaultQuality    = 8
	DefaultComplexity = 3
	MaxFrameBytes     = 512
)

var (
	ErrConfig = errors.New("speex config invalid")
	ErrInit   = errors.New("speex init failed")
	ErrEncode = errors.New("speex encode failed")
	ErrDecode = errors.New("speex decode failed")
)

type Config struct {
	Mode       string
	SampleRate int
	Quality    int
	Complexity int
	VBR        bool
}

type Encoder struct{}

type Decoder struct{}

func NewEncoder(Config) (*Encoder, error) {
	return nil, ErrInit
}

func NewDecoder(Config) (*Decoder, error) {
	return nil, ErrInit
}

func (*Encoder) FrameSize() int {
	return 0
}

func (*Decoder) FrameSize() int {
	return 0
}

func (*Encoder) Encode([]float32) ([]byte, error) {
	return nil, ErrEncode
}

func (*Decoder) Decode([]byte) ([]float32, error) {
	return nil, ErrDecode
}

func (*Encoder) Close() error {
	return nil
}

func (*Decoder) Close() error {
	return nil
}
