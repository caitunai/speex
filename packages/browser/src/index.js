const SPEEX_MAX_FRAME_BYTES = 512

export const MODE_NARROWBAND = 'nb'
export const MODE_WIDEBAND = 'wb'
export const MODE_ULTRA_WIDE = 'uwb'
export const DEFAULT_SAMPLE_RATE = 16000
export const DEFAULT_QUALITY = 8
export const DEFAULT_COMPLEXITY = 3

export class SpeexEncoder {
  constructor(module, handle, frameSize) {
    this.module = module
    this.handle = handle
    this.frameSize = frameSize
    this.pcmPtr = 0
    this.outPtr = 0
    try {
      this.pcmPtr = allocateWasmMemory(module, frameSize * 2)
      this.outPtr = allocateWasmMemory(module, SPEEX_MAX_FRAME_BYTES)
    } catch (error) {
      this.close()
      throw error
    }
  }

  static async create(options = {}) {
    const module = await createSpeexModule(options)
    const sampleRate = options.sampleRate ?? DEFAULT_SAMPLE_RATE
    const quality = options.quality ?? DEFAULT_QUALITY
    const complexity = options.complexity ?? DEFAULT_COMPLEXITY
    const vbr = options.vbr ? 1 : 0
    const handle = module._speex_js_encoder_create(sampleRate, quality, complexity, vbr)
    if (!handle) {
      throw new Error('Speex encoder initialization failed.')
    }
    const frameSize = module._speex_js_encoder_frame_size(handle)
    if (frameSize <= 0) {
      module._speex_js_encoder_destroy(handle)
      throw new Error('Speex frame size is invalid.')
    }
    return new SpeexEncoder(module, handle, frameSize)
  }

  encode(samples) {
    if (!this.handle) {
      throw new Error('Speex encoder is closed.')
    }
    if (samples.length !== this.frameSize) {
      throw new Error(`Speex encoder expects ${this.frameSize} samples.`)
    }
    const pcm = new Int16Array(this.frameSize)
    for (let i = 0; i < samples.length; i++) {
      const sample = Math.max(-1, Math.min(1, samples[i]))
      pcm[i] = Math.round(sample * 32767)
    }
    this.module.HEAP16.set(pcm, this.pcmPtr >> 1)
    const length = this.module._speex_js_encode(
      this.handle,
      this.pcmPtr,
      this.outPtr,
      SPEEX_MAX_FRAME_BYTES
    )
    if (length < 0) {
      throw new Error('Speex encode failed.')
    }
    return this.module.HEAPU8.slice(this.outPtr, this.outPtr + length)
  }

  close() {
    if (!this.module) {
      return
    }
    if (this.handle) {
      this.module._speex_js_encoder_destroy(this.handle)
      this.handle = 0
    }
    if (this.pcmPtr) {
      this.module._free(this.pcmPtr)
      this.pcmPtr = 0
    }
    if (this.outPtr) {
      this.module._free(this.outPtr)
      this.outPtr = 0
    }
    this.module = null
  }
}

export class SpeexDecoder {
  constructor(module, handle, frameSize) {
    this.module = module
    this.handle = handle
    this.frameSize = frameSize
    this.framePtr = 0
    this.pcmPtr = 0
    try {
      this.framePtr = allocateWasmMemory(module, SPEEX_MAX_FRAME_BYTES)
      this.pcmPtr = allocateWasmMemory(module, frameSize * 2)
    } catch (error) {
      this.close()
      throw error
    }
  }

  static async create(options = {}) {
    const module = await createSpeexModule(options)
    const sampleRate = options.sampleRate ?? DEFAULT_SAMPLE_RATE
    const handle = module._speex_js_decoder_create(sampleRate)
    if (!handle) {
      throw new Error('Speex decoder initialization failed.')
    }
    const frameSize = module._speex_js_decoder_frame_size(handle)
    if (frameSize <= 0) {
      module._speex_js_decoder_destroy(handle)
      throw new Error('Speex frame size is invalid.')
    }
    return new SpeexDecoder(module, handle, frameSize)
  }

  decode(frame) {
    if (!this.handle) {
      throw new Error('Speex decoder is closed.')
    }
    if (frame.byteLength > SPEEX_MAX_FRAME_BYTES) {
      throw new Error('Speex frame is too large.')
    }
    this.module.HEAPU8.set(frame, this.framePtr)
    const sampleCount = this.module._speex_js_decode(this.handle, this.framePtr, frame.byteLength, this.pcmPtr)
    if (sampleCount < 0) {
      throw new Error('Speex decode failed.')
    }
    const pcm = this.module.HEAP16.slice(this.pcmPtr >> 1, (this.pcmPtr >> 1) + sampleCount)
    const samples = new Float32Array(sampleCount)
    for (let i = 0; i < pcm.length; i++) {
      samples[i] = pcm[i] / 32768
    }
    return samples
  }

  close() {
    if (!this.module) {
      return
    }
    if (this.handle) {
      this.module._speex_js_decoder_destroy(this.handle)
      this.handle = 0
    }
    if (this.framePtr) {
      this.module._free(this.framePtr)
      this.framePtr = 0
    }
    if (this.pcmPtr) {
      this.module._free(this.pcmPtr)
      this.pcmPtr = 0
    }
    this.module = null
  }
}

export async function createSpeexModule(options = {}) {
  const factory = options.moduleFactory ?? await loadSpeexModuleFactory()
  return factory({
    locateFile: options.locateFile
  })
}

async function loadSpeexModuleFactory() {
  const module = await import('./wasm/speex-wasm.js')
  return module.default
}

function allocateWasmMemory(module, size) {
  const pointer = module._malloc(size)
  if (!pointer) {
    throw new Error('Speex WASM memory allocation failed.')
  }
  return pointer
}
