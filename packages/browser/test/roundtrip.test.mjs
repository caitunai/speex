import test from 'node:test'
import assert from 'node:assert/strict'
import { SpeexEncoder, SpeexDecoder } from '../dist/index.js'

test('Speex browser package exports encoder and decoder classes', () => {
  assert.equal(typeof SpeexEncoder.create, 'function')
  assert.equal(typeof SpeexDecoder.create, 'function')
})

test('Speex browser wrapper encodes, decodes, and closes via module factory', async () => {
  const calls = []
  const memory = {
    next: 8,
    HEAP16: new Int16Array(4096),
    HEAPU8: new Uint8Array(4096)
  }
  const moduleFactory = async () => ({
    HEAP16: memory.HEAP16,
    HEAPU8: memory.HEAPU8,
    _malloc(size) {
      const ptr = memory.next
      memory.next += size + 8
      return ptr
    },
    _free(ptr) {
      calls.push(['free', ptr])
    },
    _speex_js_encoder_create(sampleRate, quality, complexity, vbr) {
      calls.push(['encoder_create', sampleRate, quality, complexity, vbr])
      return 100
    },
    _speex_js_encoder_frame_size() {
      return 4
    },
    _speex_js_encode(_handle, _pcmPtr, outPtr) {
      memory.HEAPU8.set([1, 2, 3], outPtr)
      return 3
    },
    _speex_js_encoder_destroy(handle) {
      calls.push(['encoder_destroy', handle])
    },
    _speex_js_decoder_create(sampleRate) {
      calls.push(['decoder_create', sampleRate])
      return 200
    },
    _speex_js_decoder_frame_size() {
      return 4
    },
    _speex_js_decode(_handle, _framePtr, _frameLen, pcmPtr) {
      memory.HEAP16.set([3277, 6554, 9830, 13107], pcmPtr >> 1)
      return 4
    },
    _speex_js_decoder_destroy(handle) {
      calls.push(['decoder_destroy', handle])
    }
  })

  const encoder = await SpeexEncoder.create({ moduleFactory, sampleRate: 16000 })
  const frame = encoder.encode(new Float32Array([0, 0.1, 0.2, 0.3]))
  encoder.close()

  assert.deepEqual([...frame], [1, 2, 3])

  const decoder = await SpeexDecoder.create({ moduleFactory, sampleRate: 16000 })
  const decoded = decoder.decode(frame)
  decoder.close()

  assert.deepEqual([...decoded].map((value) => Number(value.toFixed(1))), [0.1, 0.2, 0.3, 0.4])
  assert.deepEqual(calls.filter(([name]) => name.endsWith('destroy')), [
    ['encoder_destroy', 100],
    ['decoder_destroy', 200]
  ])
})
