/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";

export const protobufPackage = "";

export interface FetchFilesRequest {
  tests: string[];
  pkg: string;
  sort: FetchFilesRequest_SortType;
}

export enum FetchFilesRequest_SortType {
  HARDCODED = 0,
  RAW = 1,
  NET = 2,
  IMPORTANCE = 3,
  UNRECOGNIZED = -1,
}

export function fetchFilesRequest_SortTypeFromJSON(
  object: any
): FetchFilesRequest_SortType {
  switch (object) {
    case 0:
    case "HARDCODED":
      return FetchFilesRequest_SortType.HARDCODED;
    case 1:
    case "RAW":
      return FetchFilesRequest_SortType.RAW;
    case 2:
    case "NET":
      return FetchFilesRequest_SortType.NET;
    case 3:
    case "IMPORTANCE":
      return FetchFilesRequest_SortType.IMPORTANCE;
    case -1:
    case "UNRECOGNIZED":
    default:
      return FetchFilesRequest_SortType.UNRECOGNIZED;
  }
}

export function fetchFilesRequest_SortTypeToJSON(
  object: FetchFilesRequest_SortType
): string {
  switch (object) {
    case FetchFilesRequest_SortType.HARDCODED:
      return "HARDCODED";
    case FetchFilesRequest_SortType.RAW:
      return "RAW";
    case FetchFilesRequest_SortType.NET:
      return "NET";
    case FetchFilesRequest_SortType.IMPORTANCE:
      return "IMPORTANCE";
    default:
      return "UNKNOWN";
  }
}

export interface FetchFilesResponse {
  id: string;
}

export interface CheckoutFilesRequest {
  files: FileMap | undefined;
}

export interface JobStatusResponse {
  complete: boolean;
  details: string;
  error: string;
  results: JobResults | undefined;
}

export interface JobResults {
  tests: string[];
  files: FileMap[];
}

export interface FileMap {
  files: { [key: string]: Uint8Array };
}

export interface FileMap_FilesEntry {
  key: string;
  value: Uint8Array;
}

const baseFetchFilesRequest: object = { tests: "", pkg: "", sort: 0 };

export const FetchFilesRequest = {
  encode(
    message: FetchFilesRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.tests) {
      writer.uint32(10).string(v!);
    }
    if (message.pkg !== "") {
      writer.uint32(18).string(message.pkg);
    }
    if (message.sort !== 0) {
      writer.uint32(24).int32(message.sort);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FetchFilesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
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
          message.sort = reader.int32() as any;
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
      message.sort = fetchFilesRequest_SortTypeFromJSON(object.sort);
    } else {
      message.sort = 0;
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
    message.sort !== undefined &&
      (obj.sort = fetchFilesRequest_SortTypeToJSON(message.sort));
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
      message.sort = 0;
    }
    return message;
  },
};

const baseFetchFilesResponse: object = { id: "" };

export const FetchFilesResponse = {
  encode(
    message: FetchFilesResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.id !== "") {
      writer.uint32(10).string(message.id);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FetchFilesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFetchFilesResponse } as FetchFilesResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = reader.string();
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
    if (object.id !== undefined && object.id !== null) {
      message.id = String(object.id);
    } else {
      message.id = "";
    }
    return message;
  },

  toJSON(message: FetchFilesResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(object: DeepPartial<FetchFilesResponse>): FetchFilesResponse {
    const message = { ...baseFetchFilesResponse } as FetchFilesResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = "";
    }
    return message;
  },
};

const baseCheckoutFilesRequest: object = {};

export const CheckoutFilesRequest = {
  encode(
    message: CheckoutFilesRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.files !== undefined) {
      FileMap.encode(message.files, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): CheckoutFilesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
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

const baseJobStatusResponse: object = {
  complete: false,
  details: "",
  error: "",
};

export const JobStatusResponse = {
  encode(
    message: JobStatusResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.complete === true) {
      writer.uint32(8).bool(message.complete);
    }
    if (message.details !== "") {
      writer.uint32(18).string(message.details);
    }
    if (message.error !== "") {
      writer.uint32(26).string(message.error);
    }
    if (message.results !== undefined) {
      JobResults.encode(message.results, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): JobStatusResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseJobStatusResponse } as JobStatusResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.complete = reader.bool();
          break;
        case 2:
          message.details = reader.string();
          break;
        case 3:
          message.error = reader.string();
          break;
        case 4:
          message.results = JobResults.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): JobStatusResponse {
    const message = { ...baseJobStatusResponse } as JobStatusResponse;
    if (object.complete !== undefined && object.complete !== null) {
      message.complete = Boolean(object.complete);
    } else {
      message.complete = false;
    }
    if (object.details !== undefined && object.details !== null) {
      message.details = String(object.details);
    } else {
      message.details = "";
    }
    if (object.error !== undefined && object.error !== null) {
      message.error = String(object.error);
    } else {
      message.error = "";
    }
    if (object.results !== undefined && object.results !== null) {
      message.results = JobResults.fromJSON(object.results);
    } else {
      message.results = undefined;
    }
    return message;
  },

  toJSON(message: JobStatusResponse): unknown {
    const obj: any = {};
    message.complete !== undefined && (obj.complete = message.complete);
    message.details !== undefined && (obj.details = message.details);
    message.error !== undefined && (obj.error = message.error);
    message.results !== undefined &&
      (obj.results = message.results
        ? JobResults.toJSON(message.results)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<JobStatusResponse>): JobStatusResponse {
    const message = { ...baseJobStatusResponse } as JobStatusResponse;
    if (object.complete !== undefined && object.complete !== null) {
      message.complete = object.complete;
    } else {
      message.complete = false;
    }
    if (object.details !== undefined && object.details !== null) {
      message.details = object.details;
    } else {
      message.details = "";
    }
    if (object.error !== undefined && object.error !== null) {
      message.error = object.error;
    } else {
      message.error = "";
    }
    if (object.results !== undefined && object.results !== null) {
      message.results = JobResults.fromPartial(object.results);
    } else {
      message.results = undefined;
    }
    return message;
  },
};

const baseJobResults: object = { tests: "" };

export const JobResults = {
  encode(
    message: JobResults,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.tests) {
      writer.uint32(10).string(v!);
    }
    for (const v of message.files) {
      FileMap.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): JobResults {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseJobResults } as JobResults;
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

  fromJSON(object: any): JobResults {
    const message = { ...baseJobResults } as JobResults;
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

  toJSON(message: JobResults): unknown {
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

  fromPartial(object: DeepPartial<JobResults>): JobResults {
    const message = { ...baseJobResults } as JobResults;
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

const baseFileMap: object = {};

export const FileMap = {
  encode(
    message: FileMap,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    Object.entries(message.files).forEach(([key, value]) => {
      FileMap_FilesEntry.encode(
        { key: key as any, value },
        writer.uint32(10).fork()
      ).ldelim();
    });
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FileMap {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
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
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value.length !== 0) {
      writer.uint32(18).bytes(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FileMap_FilesEntry {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
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

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}
