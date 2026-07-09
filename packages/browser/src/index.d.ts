export interface SpeexModuleOptions {
  locateFile?: (path: string) => string
  moduleFactory?: (options?: { locateFile?: (path: string) => string }) => Promise<unknown>
}

export interface SpeexEncoderOptions extends SpeexModuleOptions {
  sampleRate?: number
  quality?: number
  complexity?: number
  vbr?: boolean
}

export interface SpeexDecoderOptions extends SpeexModuleOptions {
  sampleRate?: number
}

export declare const MODE_NARROWBAND: 'nb'
export declare const MODE_WIDEBAND: 'wb'
export declare const MODE_ULTRA_WIDE: 'uwb'
export declare const DEFAULT_SAMPLE_RATE: 16000
export declare const DEFAULT_QUALITY: 8
export declare const DEFAULT_COMPLEXITY: 3

export declare class SpeexEncoder {
  readonly frameSize: number
  static create(options?: SpeexEncoderOptions): Promise<SpeexEncoder>
  encode(samples: Float32Array): Uint8Array
  close(): void
}

export declare class SpeexDecoder {
  readonly frameSize: number
  static create(options?: SpeexDecoderOptions): Promise<SpeexDecoder>
  decode(frame: Uint8Array): Float32Array
  close(): void
}

export declare function createSpeexModule(options?: SpeexModuleOptions): Promise<unknown>
