#include <stdlib.h>
#include <string.h>
#include "speex/speex.h"

typedef struct {
  void *state;
  SpeexBits bits;
  int frame_size;
} SpeexJsCodec;

static const SpeexMode *speex_js_mode(int sample_rate) {
  if (sample_rate == 8000) {
    return speex_lib_get_mode(SPEEX_MODEID_NB);
  }
  if (sample_rate == 16000) {
    return speex_lib_get_mode(SPEEX_MODEID_WB);
  }
  if (sample_rate == 32000) {
    return speex_lib_get_mode(SPEEX_MODEID_UWB);
  }
  return NULL;
}

SpeexJsCodec *speex_js_encoder_create(int sample_rate, int quality, int complexity, int vbr) {
  const SpeexMode *mode = speex_js_mode(sample_rate);
  if (mode == NULL) {
    return NULL;
  }
  SpeexJsCodec *codec = (SpeexJsCodec *)calloc(1, sizeof(SpeexJsCodec));
  if (codec == NULL) {
    return NULL;
  }
  codec->state = speex_encoder_init(mode);
  if (codec->state == NULL) {
    free(codec);
    return NULL;
  }
  speex_bits_init(&codec->bits);
  speex_encoder_ctl(codec->state, SPEEX_SET_SAMPLING_RATE, &sample_rate);
  speex_encoder_ctl(codec->state, SPEEX_SET_QUALITY, &quality);
  speex_encoder_ctl(codec->state, SPEEX_SET_COMPLEXITY, &complexity);
  speex_encoder_ctl(codec->state, SPEEX_SET_VBR, &vbr);
  speex_encoder_ctl(codec->state, SPEEX_GET_FRAME_SIZE, &codec->frame_size);
  return codec;
}

int speex_js_encoder_frame_size(SpeexJsCodec *codec) {
  if (codec == NULL) {
    return 0;
  }
  return codec->frame_size;
}

int speex_js_encode(SpeexJsCodec *codec, spx_int16_t *pcm, char *out, int max_len) {
  if (codec == NULL || pcm == NULL || out == NULL || max_len <= 0) {
    return -1;
  }
  speex_bits_reset(&codec->bits);
  if (speex_encode_int(codec->state, pcm, &codec->bits) < 0) {
    return -1;
  }
  return speex_bits_write(&codec->bits, out, max_len);
}

void speex_js_encoder_destroy(SpeexJsCodec *codec) {
  if (codec == NULL) {
    return;
  }
  if (codec->state != NULL) {
    speex_encoder_destroy(codec->state);
  }
  speex_bits_destroy(&codec->bits);
  free(codec);
}

SpeexJsCodec *speex_js_decoder_create(int sample_rate) {
  const SpeexMode *mode = speex_js_mode(sample_rate);
  if (mode == NULL) {
    return NULL;
  }
  SpeexJsCodec *codec = (SpeexJsCodec *)calloc(1, sizeof(SpeexJsCodec));
  if (codec == NULL) {
    return NULL;
  }
  codec->state = speex_decoder_init(mode);
  if (codec->state == NULL) {
    free(codec);
    return NULL;
  }
  speex_bits_init(&codec->bits);
  speex_decoder_ctl(codec->state, SPEEX_SET_SAMPLING_RATE, &sample_rate);
  speex_decoder_ctl(codec->state, SPEEX_GET_FRAME_SIZE, &codec->frame_size);
  return codec;
}

int speex_js_decoder_frame_size(SpeexJsCodec *codec) {
  if (codec == NULL) {
    return 0;
  }
  return codec->frame_size;
}

int speex_js_decode(SpeexJsCodec *codec, char *frame, int frame_len, spx_int16_t *pcm) {
  if (codec == NULL || frame == NULL || frame_len <= 0 || pcm == NULL) {
    return -1;
  }
  speex_bits_read_from(&codec->bits, frame, frame_len);
  if (speex_decode_int(codec->state, &codec->bits, pcm) < 0) {
    return -1;
  }
  return codec->frame_size;
}

void speex_js_decoder_destroy(SpeexJsCodec *codec) {
  if (codec == NULL) {
    return;
  }
  if (codec->state != NULL) {
    speex_decoder_destroy(codec->state);
  }
  speex_bits_destroy(&codec->bits);
  free(codec);
}
