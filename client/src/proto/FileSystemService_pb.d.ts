// package: pb
// file: FileSystemService.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class UploadFileRequestFragment extends jspb.Message { 
    getPath(): string;
    setPath(value: string): UploadFileRequestFragment;
    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): UploadFileRequestFragment;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UploadFileRequestFragment.AsObject;
    static toObject(includeInstance: boolean, msg: UploadFileRequestFragment): UploadFileRequestFragment.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UploadFileRequestFragment, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UploadFileRequestFragment;
    static deserializeBinaryFromReader(message: UploadFileRequestFragment, reader: jspb.BinaryReader): UploadFileRequestFragment;
}

export namespace UploadFileRequestFragment {
    export type AsObject = {
        path: string,
        data: Uint8Array | string,
    }
}

export class SuccessResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): SuccessResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SuccessResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SuccessResponse): SuccessResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SuccessResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SuccessResponse;
    static deserializeBinaryFromReader(message: SuccessResponse, reader: jspb.BinaryReader): SuccessResponse;
}

export namespace SuccessResponse {
    export type AsObject = {
        success: boolean,
    }
}

export class DownloadFileRequest extends jspb.Message { 
    getPath(): string;
    setPath(value: string): DownloadFileRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DownloadFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DownloadFileRequest): DownloadFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DownloadFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DownloadFileRequest;
    static deserializeBinaryFromReader(message: DownloadFileRequest, reader: jspb.BinaryReader): DownloadFileRequest;
}

export namespace DownloadFileRequest {
    export type AsObject = {
        path: string,
    }
}

export class DownloadFileResponseFragment extends jspb.Message { 
    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): DownloadFileResponseFragment;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DownloadFileResponseFragment.AsObject;
    static toObject(includeInstance: boolean, msg: DownloadFileResponseFragment): DownloadFileResponseFragment.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DownloadFileResponseFragment, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DownloadFileResponseFragment;
    static deserializeBinaryFromReader(message: DownloadFileResponseFragment, reader: jspb.BinaryReader): DownloadFileResponseFragment;
}

export namespace DownloadFileResponseFragment {
    export type AsObject = {
        data: Uint8Array | string,
    }
}

export class FileInfo extends jspb.Message { 
    getPath(): string;
    setPath(value: string): FileInfo;
    getSize(): number;
    setSize(value: number): FileInfo;
    getLastModifiedTime(): string;
    setLastModifiedTime(value: string): FileInfo;
    getIsDirectory(): boolean;
    setIsDirectory(value: boolean): FileInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FileInfo.AsObject;
    static toObject(includeInstance: boolean, msg: FileInfo): FileInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FileInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FileInfo;
    static deserializeBinaryFromReader(message: FileInfo, reader: jspb.BinaryReader): FileInfo;
}

export namespace FileInfo {
    export type AsObject = {
        path: string,
        size: number,
        lastModifiedTime: string,
        isDirectory: boolean,
    }
}

export class FileInfoRequest extends jspb.Message { 
    getPath(): string;
    setPath(value: string): FileInfoRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FileInfoRequest.AsObject;
    static toObject(includeInstance: boolean, msg: FileInfoRequest): FileInfoRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FileInfoRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FileInfoRequest;
    static deserializeBinaryFromReader(message: FileInfoRequest, reader: jspb.BinaryReader): FileInfoRequest;
}

export namespace FileInfoRequest {
    export type AsObject = {
        path: string,
    }
}
