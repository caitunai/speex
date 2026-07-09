#!/usr/bin/env sh
set -eu

PACKAGE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
ROOT_DIR="$(CDPATH= cd -- "$PACKAGE_DIR/../.." && pwd)"
OUT_DIR="$PACKAGE_DIR/dist/wasm"
export EM_CACHE="$ROOT_DIR/.cache/emscripten"
mkdir -p "$EM_CACHE"
mkdir -p "$OUT_DIR"

emcc \
  "$ROOT_DIR/speex_all.c" \
  "$PACKAGE_DIR/src/speex_wasm.c" \
  -I"$ROOT_DIR/internal/csrc/speex/include" \
  -I"$ROOT_DIR/internal/csrc/speex/include/speex" \
  -I"$ROOT_DIR/internal/csrc/speex/libspeex" \
  -DFLOATING_POINT \
  -DUSE_SMALLFT \
  -DEXPORT= \
  -O3 \
  -s WASM=1 \
  -s MODULARIZE=1 \
  -s EXPORT_ES6=1 \
  -s ENVIRONMENT=web,worker \
  -s ALLOW_MEMORY_GROWTH=1 \
  -s EXPORTED_RUNTIME_METHODS='["HEAP16","HEAPU8"]' \
  -s EXPORTED_FUNCTIONS='["_malloc","_free","_speex_js_encoder_create","_speex_js_encoder_frame_size","_speex_js_encode","_speex_js_encoder_destroy","_speex_js_decoder_create","_speex_js_decoder_frame_size","_speex_js_decode","_speex_js_decoder_destroy"]' \
  -o "$OUT_DIR/speex-wasm.js"
