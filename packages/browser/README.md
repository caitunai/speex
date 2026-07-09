# @caitun/speex

Browser Speex encoder and decoder powered by Speex compiled to WebAssembly.

## Install

```bash
npm install @caitun/speex
```

## Usage

```js
import { SpeexEncoder, SpeexDecoder } from '@caitun/speex'

const encoder = await SpeexEncoder.create({
  sampleRate: 16000,
  quality: 8,
  complexity: 3
})

const pcm = new Float32Array(encoder.frameSize)
const frame = encoder.encode(pcm)
encoder.close()

const decoder = await SpeexDecoder.create({
  sampleRate: 16000
})

const decoded = decoder.decode(frame)
decoder.close()
```

If your bundler serves WASM assets from a custom location, pass `locateFile`:

```js
const encoder = await SpeexEncoder.create({
  locateFile: (path) => `/assets/${path}`
})
```

## Build

```bash
npm run build
```

`build:wasm` requires Emscripten `emcc`.
