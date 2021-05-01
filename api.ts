/* eslint-disable */
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import * as Long from "long";

export const protobufPackage = "";

export interface FetchFilesRequest {
  tests: string[];
  pkg: string;
  sort: string;
}

export interface FileMap {
  files: { [key: string]: Uint8Array };
}

export interface FileMap_FilesEntry {
  key: string;
  value: Uint8Array;
}

export interface FetchFilesResponse {
  tests: string[];
  files: FileMap[];
}

export interface CheckoutFilesRequest {
  files: FileMap | undefined;
}

const baseFetchFilesRequest: object = { tests: "", pkg: "", sort: "" };

export const FetchFilesRequest = {
  encode(message: FetchFilesRequest, writer: Writer = Writer.create()): Writer {
    for (const v of message.tests) {
      writer.uint32(10).string(v!);
    }
    if (message.pkg !== "") {
      writer.uint32(18).string(message.pkg);
    }
    if (message.sort !== "") {
      writer.uint32(26).string(message.sort);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): FetchFilesRequest {
    const reader = input instanceof Reader ? input : new Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFetchFilesRequest } as FetchFilesRequest;
    message.tests = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.tests.push(reader.string());
          break;
        case 2:
          message.pkg = reader.string();
          break;
        case 3:
          message.sort = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FetchFilesRequest {
    const message = { ...baseFetchFilesRequest } as FetchFilesRequest;
    message.tests = [];
    if (object.tests !== undefined && object.tests !== null) {
      for (const e of object.tests) {
        message.tests.push(String(e));
      }
    }
    if (object.pkg !== undefined && object.pkg !== null) {
      message.pkg = String(object.pkg);
    } else {
      message.pkg = "";
    }
    if (object.sort !== undefined && object.sort !== null) {
      message.sort = String(object.sort);
    } else {
      message.sort = "";
    }
    return message;
  },

  toJSON(message: FetchFilesRequest): unknown {
    const obj: any = {};
    if (message.tests) {
      obj.tests = message.tests.map((e) => e);
    } else {
      obj.tests = [];
    }
    message.pkg !== undefined && (obj.pkg = message.pkg);
    message.sort !== undefined && (obj.sort = message.sort);
    return obj;
  },

  fromPartial(object: DeepPartial<FetchFilesRequest>): FetchFilesRequest {
    const message = { ...baseFetchFilesRequest } as FetchFilesRequest;
    message.tests = [];
    if (object.tests !== undefined && object.tests !== null) {
      for (const e of object.tests) {
        message.tests.push(e);
      }
    }
    if (object.pkg !== undefined && object.pkg !== null) {
      message.pkg = object.pkg;
    } else {
      message.pkg = "";
    }
    if (object.sort !== undefined && object.sort !== null) {
      message.sort = object.sort;
    } else {
      message.sort = "";
    }
    return message;
  },
};

const baseFileMap: object = {};

export const FileMap = {
  encode(message: FileMap, writer: Writer = Writer.create()): Writer {
    Object.entries(message.files).forEach(([key, value]) => {
      FileMap_FilesEntry.encode(
        { key: key as any, value },
        writer.uint32(10).fork()
      ).ldelim();
    });
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): FileMap {
    const reader = input instanceof Reader ? input : new Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFileMap } as FileMap;
    message.files = {};
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          const entry1 = FileMap_FilesEntry.decode(reader, reader.uint32());
          if (entry1.value !== undefined) {
            message.files[entry1.key] = entry1.value;
          }
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FileMap {
    const message = { ...baseFileMap } as FileMap;
    message.files = {};
    if (object.files !== undefined && object.files !== null) {
      Object.entries(object.files).forEach(([key, value]) => {
        message.files[key] = bytesFromBase64(value as string);
      });
    }
    return message;
  },

  toJSON(message: FileMap): unknown {
    const obj: any = {};
    obj.files = {};
    if (message.files) {
      Object.entries(message.files).forEach(([k, v]) => {
        obj.files[k] = base64FromBytes(v);
      });
    }
    return obj;
  },

  fromPartial(object: DeepPartial<FileMap>): FileMap {
    const message = { ...baseFileMap } as FileMap;
    message.files = {};
    if (object.files !== undefined && object.files !== null) {
      Object.entries(object.files).forEach(([key, value]) => {
        if (value !== undefined) {
          message.files[key] = value;
        }
      });
    }
    return message;
  },
};

const baseFileMap_FilesEntry: object = { key: "" };

export const FileMap_FilesEntry = {
  encode(
    message: FileMap_FilesEntry,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value.length !== 0) {
      writer.uint32(18).bytes(message.value);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): FileMap_FilesEntry {
    const reader = input instanceof Reader ? input : new Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFileMap_FilesEntry } as FileMap_FilesEntry;
    message.value = new Uint8Array();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.key = reader.string();
          break;
        case 2:
          message.value = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FileMap_FilesEntry {
    const message = { ...baseFileMap_FilesEntry } as FileMap_FilesEntry;
    message.value = new Uint8Array();
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = bytesFromBase64(object.value);
    }
    return message;
  },

  toJSON(message: FileMap_FilesEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined &&
      (obj.value = base64FromBytes(
        message.value !== undefined ? message.value : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<FileMap_FilesEntry>): FileMap_FilesEntry {
    const message = { ...baseFileMap_FilesEntry } as FileMap_FilesEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = new Uint8Array();
    }
    return message;
  },
};

const baseFetchFilesResponse: object = { tests: "" };

export const FetchFilesResponse = {
  encode(
    message: FetchFilesResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.tests) {
      writer.uint32(10).string(v!);
    }
    for (const v of message.files) {
      FileMap.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): FetchFilesResponse {
    const reader = input instanceof Reader ? input : new Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFetchFilesResponse } as FetchFilesResponse;
    message.tests = [];
    message.files = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.tests.push(reader.string());
          break;
        case 2:
          message.files.push(FileMap.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FetchFilesResponse {
    const message = { ...baseFetchFilesResponse } as FetchFilesResponse;
    message.tests = [];
    message.files = [];
    if (object.tests !== undefined && object.tests !== null) {
      for (const e of object.tests) {
        message.tests.push(String(e));
      }
    }
    if (object.files !== undefined && object.files !== null) {
      for (const e of object.files) {
        message.files.push(FileMap.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: FetchFilesResponse): unknown {
    const obj: any = {};
    if (message.tests) {
      obj.tests = message.tests.map((e) => e);
    } else {
      obj.tests = [];
    }
    if (message.files) {
      obj.files = message.files.map((e) => (e ? FileMap.toJSON(e) : undefined));
    } else {
      obj.files = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<FetchFilesResponse>): FetchFilesResponse {
    const message = { ...baseFetchFilesResponse } as FetchFilesResponse;
    message.tests = [];
    message.files = [];
    if (object.tests !== undefined && object.tests !== null) {
      for (const e of object.tests) {
        message.tests.push(e);
      }
    }
    if (object.files !== undefined && object.files !== null) {
      for (const e of object.files) {
        message.files.push(FileMap.fromPartial(e));
      }
    }
    return message;
  },
};

const baseCheckoutFilesRequest: object = {};

export const CheckoutFilesRequest = {
  encode(
    message: CheckoutFilesRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.files !== undefined) {
      FileMap.encode(message.files, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): CheckoutFilesRequest {
    const reader = input instanceof Reader ? input : new Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseCheckoutFilesRequest } as CheckoutFilesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.files = FileMap.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CheckoutFilesRequest {
    const message = { ...baseCheckoutFilesRequest } as CheckoutFilesRequest;
    if (object.files !== undefined && object.files !== null) {
      message.files = FileMap.fromJSON(object.files);
    } else {
      message.files = undefined;
    }
    return message;
  },

  toJSON(message: CheckoutFilesRequest): unknown {
    const obj: any = {};
    message.files !== undefined &&
      (obj.files = message.files ? FileMap.toJSON(message.files) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<CheckoutFilesRequest>): CheckoutFilesRequest {
    const message = { ...baseCheckoutFilesRequest } as CheckoutFilesRequest;
    if (object.files !== undefined && object.files !== null) {
      message.files = FileMap.fromPartial(object.files);
    } else {
      message.files = undefined;
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64);
  const arr = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i);
  }
  return arr;
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = [];
  for (let i = 0; i < arr.byteLength; ++i) {
    bin.push(String.fromCharCode(arr[i]));
  }
  return btoa(bin.join(""));
}

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

// If you get a compile-error about 'Constructor<Long> and ... have no overlap',
// add '--ts_proto_opt=esModuleInterop=true' as a flag when calling 'protoc'.
if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
