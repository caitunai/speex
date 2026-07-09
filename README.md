# Speex Codecs for Go and Browser

This repository packages the Speex narrowband, wideband, and ultra-wideband encoder/decoder for both Go and browser JavaScript.

- Go module: `github.com/caitunai/speex`
- npm package: `@caitun/speex`
- C source: vendored in `internal/csrc/speex`
- Browser runtime: Emscripten WebAssembly built from the same C source

The published npm package includes prebuilt `dist/wasm/speex-wasm.js` and `dist/wasm/speex-wasm.wasm`, so consumers do not need Emscripten unless they want to rebuild the WebAssembly artifact.

## Go Usage

```go
package main

import (
	"log"

	"github.com/caitunai/speex"
)

func main() {
	encoder, err := speex.NewEncoder(speex.Config{
		SampleRate: 16000,
		Mode:       speex.ModeWideband,
		Quality:    speex.DefaultQuality,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer encoder.Close()

	decoder, err := speex.NewDecoder(speex.Config{
		SampleRate: 16000,
		Mode:       speex.ModeWideband,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer decoder.Close()

	pcm := make([]float32, encoder.FrameSize())
	frame, err := encoder.Encode(pcm)
	if err != nil {
		log.Fatal(err)
	}
	decoded, err := decoder.Decode(frame)
	if err != nil {
		log.Fatal(err)
	}
	_ = decoded
}
```

Supported modes:

- `speex.ModeNarrowband`: 8 kHz, 160 samples per frame
- `speex.ModeWideband`: 16 kHz, 320 samples per frame
- `speex.ModeUltraWide`: 32 kHz, 640 samples per frame

The Go package uses cgo and the vendored Speex C source. With `CGO_ENABLED=0`, constructors return `speex.ErrInit` so callers can detect unsupported builds cleanly.

## Browser Usage

```bash
npm install @caitun/speex
```

```js
import { SpeexEncoder, SpeexDecoder } from '@caitun/speex'

const encoder = await SpeexEncoder.create({ sampleRate: 16000 })
const decoder = await SpeexDecoder.create({ sampleRate: 16000 })

const pcm = new Float32Array(encoder.frameSize)
const frame = encoder.encode(pcm)
const decoded = decoder.decode(frame)

encoder.close()
decoder.close()
```

If your bundler serves the `.wasm` file from a custom path, pass `locateFile`:

```js
const encoder = await SpeexEncoder.create({
  sampleRate: 16000,
  locateFile: (path) => `/assets/speex/${path}`
})
```

## Development

Run Go tests:

```bash
GOCACHE="$PWD/.cache/go-build" go test ./...
CGO_ENABLED=0 GOCACHE="$PWD/.cache/go-build" go test ./...
```

Build and test the browser package:

```bash
cd packages/browser
npm run build
npm test
npm pack --dry-run
```

`npm run build` requires Emscripten only when rebuilding `dist/wasm`. The CI workflow installs Emscripten with `mymindstorm/setup-emsdk`.

## Release

1. Update `packages/browser/package.json` to the target version.
2. Commit the changes.
3. Push a tag matching the npm version, for example `v0.1.0`.
4. GitHub Actions runs Go tests, rebuilds the browser package, verifies the tag and package versions match, and publishes `@caitun/speex` to npm.

The npm publish job requires an `NPM_TOKEN` repository secret.

## License

This wrapper package is MIT licensed. The vendored Speex source keeps its original BSD-style license. See `THIRD_PARTY_LICENSES.md` and `internal/csrc/speex/COPYING`.
