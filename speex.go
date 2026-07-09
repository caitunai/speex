//go:build cgo

//nolint:gocritic // Speex cgo macros trigger false dupSubExpr reports.
package speex

/*
#cgo CFLAGS: -I${SRCDIR}/internal/csrc/speex/include -I${SRCDIR}/internal/csrc/speex/include/speex -I${SRCDIR}/internal/csrc/speex/libspeex -DFLOATING_POINT -DUSE_SMALLFT -DEXPORT=
#cgo !darwin LDFLAGS: -lm
#include <stdlib.h>
#include "speex/speex.h"
*/
import "C"

import (
	"errors"
	"math"
	"unsafe"
)

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

type Encoder struct {
	state      unsafe.Pointer
	bits       C.SpeexBits
	frameSize  int
	sampleRate int
	closed     bool
}

type Decoder struct {
	state      unsafe.Pointer
	bits       C.SpeexBits
	frameSize  int
	sampleRate int
	closed     bool
}

func NewEncoder(cfg Config) (*Encoder, error) {
	cfg = normalizeConfig(cfg)
	mode, err := modeForConfig(cfg)
	if err != nil {
		return nil, err
	}
	state := C.speex_encoder_init(mode)
	if state == nil {
		return nil, ErrInit
	}
	encoder := &Encoder{
		state:      state,
		sampleRate: cfg.SampleRate,
	}
	C.speex_bits_init(&encoder.bits)

	if err := encoder.setInt(C.SPEEX_SET_SAMPLING_RATE, cfg.SampleRate, ErrInit); err != nil {
		_ = encoder.Close()
		return nil, err
	}
	if err := encoder.setInt(C.SPEEX_SET_QUALITY, cfg.Quality, ErrInit); err != nil {
		_ = encoder.Close()
		return nil, err
	}
	if err := encoder.setInt(C.SPEEX_SET_COMPLEXITY, cfg.Complexity, ErrInit); err != nil {
		_ = encoder.Close()
		return nil, err
	}
	vbr := 0
	if cfg.VBR {
		vbr = 1
	}
	if err := encoder.setInt(C.SPEEX_SET_VBR, vbr, ErrInit); err != nil {
		_ = encoder.Close()
		return nil, err
	}
	frameSize, err := encoder.getFrameSize(ErrInit)
	if err != nil {
		_ = encoder.Close()
		return nil, err
	}
	encoder.frameSize = frameSize
	return encoder, nil
}

func NewDecoder(cfg Config) (*Decoder, error) {
	cfg = normalizeConfig(cfg)
	mode, err := modeForConfig(cfg)
	if err != nil {
		return nil, err
	}
	state := C.speex_decoder_init(mode)
	if state == nil {
		return nil, ErrInit
	}
	decoder := &Decoder{
		state:      state,
		sampleRate: cfg.SampleRate,
	}
	C.speex_bits_init(&decoder.bits)
	if err := decoder.setInt(C.SPEEX_SET_SAMPLING_RATE, cfg.SampleRate, ErrInit); err != nil {
		_ = decoder.Close()
		return nil, err
	}
	enhance := 1
	if err := decoder.setInt(C.SPEEX_SET_ENH, enhance, ErrInit); err != nil {
		_ = decoder.Close()
		return nil, err
	}
	frameSize, err := decoder.getFrameSize(ErrInit)
	if err != nil {
		_ = decoder.Close()
		return nil, err
	}
	decoder.frameSize = frameSize
	return decoder, nil
}

func (e *Encoder) FrameSize() int {
	if e == nil {
		return 0
	}
	return e.frameSize
}

func (d *Decoder) FrameSize() int {
	if d == nil {
		return 0
	}
	return d.frameSize
}

func (e *Encoder) Encode(samples []float32) ([]byte, error) {
	if e == nil || e.closed || e.state == nil {
		return nil, ErrEncode
	}
	if len(samples) != e.frameSize {
		return nil, ErrConfig
	}
	pcm := float32ToInt16(samples)
	out := make([]byte, MaxFrameBytes)
	C.speex_bits_reset(&e.bits)
	ret := C.speex_encode_int(e.state, (*C.spx_int16_t)(unsafe.Pointer(&pcm[0])), &e.bits)
	if ret < 0 {
		return nil, ErrEncode
	}
	n := C.speex_bits_write(&e.bits, (*C.char)(unsafe.Pointer(&out[0])), C.int(len(out)))
	if n < 0 {
		return nil, ErrEncode
	}
	return out[:int(n)], nil
}

func (d *Decoder) Decode(frame []byte) ([]float32, error) {
	if d == nil || d.closed || d.state == nil {
		return nil, ErrDecode
	}
	if len(frame) == 0 {
		return nil, nil
	}
	pcm := make([]int16, d.frameSize)
	C.speex_bits_read_from(&d.bits, (*C.char)(unsafe.Pointer(&frame[0])), C.int(len(frame)))
	ret := C.speex_decode_int(d.state, &d.bits, (*C.spx_int16_t)(unsafe.Pointer(&pcm[0])))
	if ret < 0 {
		return nil, ErrDecode
	}
	return int16ToFloat32(pcm), nil
}

func (e *Encoder) Close() error {
	if e == nil || e.closed {
		return nil
	}
	e.closed = true
	if e.state != nil {
		C.speex_encoder_destroy(e.state)
		e.state = nil
	}
	C.speex_bits_destroy(&e.bits)
	return nil
}

func (d *Decoder) Close() error {
	if d == nil || d.closed {
		return nil
	}
	d.closed = true
	if d.state != nil {
		C.speex_decoder_destroy(d.state)
		d.state = nil
	}
	C.speex_bits_destroy(&d.bits)
	return nil
}

func (e *Encoder) setInt(request C.int, value int, sentinel error) error {
	cValue := C.int(value)
	if C.speex_encoder_ctl(e.state, request, unsafe.Pointer(&cValue)) != 0 {
		return sentinel
	}
	return nil
}

func (d *Decoder) setInt(request C.int, value int, sentinel error) error {
	cValue := C.int(value)
	if C.speex_decoder_ctl(d.state, request, unsafe.Pointer(&cValue)) != 0 {
		return sentinel
	}
	return nil
}

func (e *Encoder) getFrameSize(sentinel error) (int, error) {
	var frameSize C.int
	if C.speex_encoder_ctl(e.state, C.SPEEX_GET_FRAME_SIZE, unsafe.Pointer(&frameSize)) != 0 {
		return 0, sentinel
	}
	return int(frameSize), nil
}

func (d *Decoder) getFrameSize(sentinel error) (int, error) {
	var frameSize C.int
	if C.speex_decoder_ctl(d.state, C.SPEEX_GET_FRAME_SIZE, unsafe.Pointer(&frameSize)) != 0 {
		return 0, sentinel
	}
	return int(frameSize), nil
}

func normalizeConfig(cfg Config) Config {
	if cfg.SampleRate == 0 {
		cfg.SampleRate = 16000
	}
	if cfg.Quality == 0 {
		cfg.Quality = DefaultQuality
	}
	if cfg.Complexity == 0 {
		cfg.Complexity = DefaultComplexity
	}
	if cfg.Mode == "" {
		switch cfg.SampleRate {
		case 8000:
			cfg.Mode = ModeNarrowband
		case 16000:
			cfg.Mode = ModeWideband
		case 32000:
			cfg.Mode = ModeUltraWide
		}
	}
	return cfg
}

func modeForConfig(cfg Config) (*C.SpeexMode, error) {
	switch cfg.Mode {
	case ModeNarrowband:
		if cfg.SampleRate != 8000 {
			return nil, ErrConfig
		}
		return C.speex_lib_get_mode(C.SPEEX_MODEID_NB), nil
	case ModeWideband:
		if cfg.SampleRate != 16000 {
			return nil, ErrConfig
		}
		return C.speex_lib_get_mode(C.SPEEX_MODEID_WB), nil
	case ModeUltraWide:
		if cfg.SampleRate != 32000 {
			return nil, ErrConfig
		}
		return C.speex_lib_get_mode(C.SPEEX_MODEID_UWB), nil
	default:
		return nil, ErrConfig
	}
}

func float32ToInt16(samples []float32) []int16 {
	pcm := make([]int16, len(samples))
	for i, sample := range samples {
		sample = min(max(sample, -1), 1)
		pcm[i] = int16(math.Round(float64(sample * 32767)))
	}
	return pcm
}

func int16ToFloat32(samples []int16) []float32 {
	pcm := make([]float32, len(samples))
	for i, sample := range samples {
		pcm[i] = float32(sample) / 32768
	}
	return pcm
}
