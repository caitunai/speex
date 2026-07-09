package speex

import (
	"errors"
	"math"
	"testing"
)

func TestSpeexEncodeDecodeWideband(t *testing.T) {
	encoder, err := NewEncoder(Config{SampleRate: 16000, Mode: ModeWideband})
	if err != nil {
		if errors.Is(err, ErrInit) {
			t.Skipf("speex unavailable: %v", err)
		}
		t.Fatalf("new encoder: %v", err)
	}
	defer func() {
		if err := encoder.Close(); err != nil {
			t.Fatalf("close encoder: %v", err)
		}
	}()
	decoder, err := NewDecoder(Config{SampleRate: 16000, Mode: ModeWideband})
	if err != nil {
		t.Fatalf("new decoder: %v", err)
	}
	defer func() {
		if err := decoder.Close(); err != nil {
			t.Fatalf("close decoder: %v", err)
		}
	}()

	if encoder.FrameSize() != 320 {
		t.Fatalf("frame size = %d, want 320", encoder.FrameSize())
	}
	samples := make([]float32, encoder.FrameSize())
	for i := range samples {
		samples[i] = float32(math.Sin(2*math.Pi*float64(i)/80) * 0.25)
	}
	frame, err := encoder.Encode(samples)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(frame) == 0 {
		t.Fatal("encoded frame is empty")
	}
	decoded, err := decoder.Decode(frame)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(decoded) != encoder.FrameSize() {
		t.Fatalf("decoded samples = %d, want %d", len(decoded), encoder.FrameSize())
	}
}
