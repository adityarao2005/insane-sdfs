// package: pb
// file: FileSystemService.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as FileSystemService_pb from "./FileSystemService_pb";

interface IFileSystemService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    uploadFile: IFileSystemService_IUploadFile;
    downloadFile: IFileSystemService_IDownloadFile;
    deleteFile: IFileSystemService_IDeleteFile;
    getFileInfo: IFileSystemService_IGetFileInfo;
    listFiles: IFileSystemService_IListFiles;
    createDirectory: IFileSystemService_ICreateDirectory;
    ping: IFileSystemService_IPing;
}

interface IFileSystemService_IUploadFile extends grpc.MethodDefinition<FileSystemService_pb.UploadFileRequestFragment, FileSystemService_pb.SuccessResponse> {
    path: "/pb.FileSystem/UploadFile";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<FileSystemService_pb.UploadFileRequestFragment>;
    requestDeserialize: grpc.deserialize<FileSystemService_pb.UploadFileRequestFragment>;
    responseSerialize: grpc.serialize<FileSystemService_pb.SuccessResponse>;
    responseDeserialize: grpc.deserialize<FileSystemService_pb.SuccessResponse>;
}
interface IFileSystemService_IDownloadFile extends grpc.MethodDefinition<FileSystemService_pb.DownloadFileRequest, FileSystemService_pb.DownloadFileResponseFragment> {
    path: "/pb.FileSystem/DownloadFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<FileSystemService_pb.DownloadFileRequest>;
    requestDeserialize: grpc.deserialize<FileSystemService_pb.DownloadFileRequest>;
    responseSerialize: grpc.serialize<FileSystemService_pb.DownloadFileResponseFragment>;
    responseDeserialize: grpc.deserialize<FileSystemService_pb.DownloadFileResponseFragment>;
}
interface IFileSystemService_IDeleteFile extends grpc.MethodDefinition<FileSystemService_pb.FileInfoRequest, FileSystemService_pb.SuccessResponse> {
    path: "/pb.FileSystem/DeleteFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<FileSystemService_pb.FileInfoRequest>;
    requestDeserialize: grpc.deserialize<FileSystemService_pb.FileInfoRequest>;
    responseSerialize: grpc.serialize<FileSystemService_pb.SuccessResponse>;
    responseDeserialize: grpc.deserialize<FileSystemService_pb.SuccessResponse>;
}
interface IFileSystemService_IGetFileInfo extends grpc.MethodDefinition<FileSystemService_pb.FileInfoRequest, FileSystemService_pb.FileInfo> {
    path: "/pb.FileSystem/GetFileInfo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<FileSystemService_pb.FileInfoRequest>;
    requestDeserialize: grpc.deserialize<FileSystemService_pb.FileInfoRequest>;
    responseSerialize: grpc.serialize<FileSystemService_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<FileSystemService_pb.FileInfo>;
}
interface IFileSystemService_IListFiles extends grpc.MethodDefinition<FileSystemService_pb.FileInfoRequest, FileSystemService_pb.FileInfo> {
    path: "/pb.FileSystem/ListFiles";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<FileSystemService_pb.FileInfoRequest>;
    requestDeserialize: grpc.deserialize<FileSystemService_pb.FileInfoRequest>;
    responseSerialize: grpc.serialize<FileSystemService_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<FileSystemService_pb.FileInfo>;
}
interface IFileSystemService_ICreateDirectory extends grpc.MethodDefinition<FileSystemService_pb.FileInfoRequest, FileSystemService_pb.SuccessResponse> {
    path: "/pb.FileSystem/CreateDirectory";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<FileSystemService_pb.FileInfoRequest>;
    requestDeserialize: grpc.deserialize<FileSystemService_pb.FileInfoRequest>;
    responseSerialize: grpc.serialize<FileSystemService_pb.SuccessResponse>;
    responseDeserialize: grpc.deserialize<FileSystemService_pb.SuccessResponse>;
}
interface IFileSystemService_IPing extends grpc.MethodDefinition<FileSystemService_pb.PingRequest, FileSystemService_pb.PongResponse> {
    path: "/pb.FileSystem/Ping";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<FileSystemService_pb.PingRequest>;
    requestDeserialize: grpc.deserialize<FileSystemService_pb.PingRequest>;
    responseSerialize: grpc.serialize<FileSystemService_pb.PongResponse>;
    responseDeserialize: grpc.deserialize<FileSystemService_pb.PongResponse>;
}

export const FileSystemService: IFileSystemService;

export interface IFileSystemServer extends grpc.UntypedServiceImplementation {
    uploadFile: grpc.handleClientStreamingCall<FileSystemService_pb.UploadFileRequestFragment, FileSystemService_pb.SuccessResponse>;
    downloadFile: grpc.handleServerStreamingCall<FileSystemService_pb.DownloadFileRequest, FileSystemService_pb.DownloadFileResponseFragment>;
    deleteFile: grpc.handleUnaryCall<FileSystemService_pb.FileInfoRequest, FileSystemService_pb.SuccessResponse>;
    getFileInfo: grpc.handleUnaryCall<FileSystemService_pb.FileInfoRequest, FileSystemService_pb.FileInfo>;
    listFiles: grpc.handleServerStreamingCall<FileSystemService_pb.FileInfoRequest, FileSystemService_pb.FileInfo>;
    createDirectory: grpc.handleUnaryCall<FileSystemService_pb.FileInfoRequest, FileSystemService_pb.SuccessResponse>;
    ping: grpc.handleUnaryCall<FileSystemService_pb.PingRequest, FileSystemService_pb.PongResponse>;
}

export interface IFileSystemClient {
    uploadFile(callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientWritableStream<FileSystemService_pb.UploadFileRequestFragment>;
    uploadFile(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientWritableStream<FileSystemService_pb.UploadFileRequestFragment>;
    uploadFile(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientWritableStream<FileSystemService_pb.UploadFileRequestFragment>;
    uploadFile(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientWritableStream<FileSystemService_pb.UploadFileRequestFragment>;
    downloadFile(request: FileSystemService_pb.DownloadFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<FileSystemService_pb.DownloadFileResponseFragment>;
    downloadFile(request: FileSystemService_pb.DownloadFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<FileSystemService_pb.DownloadFileResponseFragment>;
    deleteFile(request: FileSystemService_pb.FileInfoRequest, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    deleteFile(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    deleteFile(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    getFileInfo(request: FileSystemService_pb.FileInfoRequest, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.FileInfo) => void): grpc.ClientUnaryCall;
    getFileInfo(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.FileInfo) => void): grpc.ClientUnaryCall;
    getFileInfo(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.FileInfo) => void): grpc.ClientUnaryCall;
    listFiles(request: FileSystemService_pb.FileInfoRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<FileSystemService_pb.FileInfo>;
    listFiles(request: FileSystemService_pb.FileInfoRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<FileSystemService_pb.FileInfo>;
    createDirectory(request: FileSystemService_pb.FileInfoRequest, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    createDirectory(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    createDirectory(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    ping(request: FileSystemService_pb.PingRequest, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.PongResponse) => void): grpc.ClientUnaryCall;
    ping(request: FileSystemService_pb.PingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.PongResponse) => void): grpc.ClientUnaryCall;
    ping(request: FileSystemService_pb.PingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.PongResponse) => void): grpc.ClientUnaryCall;
}

export class FileSystemClient extends grpc.Client implements IFileSystemClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public uploadFile(callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientWritableStream<FileSystemService_pb.UploadFileRequestFragment>;
    public uploadFile(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientWritableStream<FileSystemService_pb.UploadFileRequestFragment>;
    public uploadFile(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientWritableStream<FileSystemService_pb.UploadFileRequestFragment>;
    public uploadFile(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientWritableStream<FileSystemService_pb.UploadFileRequestFragment>;
    public downloadFile(request: FileSystemService_pb.DownloadFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<FileSystemService_pb.DownloadFileResponseFragment>;
    public downloadFile(request: FileSystemService_pb.DownloadFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<FileSystemService_pb.DownloadFileResponseFragment>;
    public deleteFile(request: FileSystemService_pb.FileInfoRequest, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    public deleteFile(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    public deleteFile(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    public getFileInfo(request: FileSystemService_pb.FileInfoRequest, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public getFileInfo(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public getFileInfo(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public listFiles(request: FileSystemService_pb.FileInfoRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<FileSystemService_pb.FileInfo>;
    public listFiles(request: FileSystemService_pb.FileInfoRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<FileSystemService_pb.FileInfo>;
    public createDirectory(request: FileSystemService_pb.FileInfoRequest, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    public createDirectory(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    public createDirectory(request: FileSystemService_pb.FileInfoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.SuccessResponse) => void): grpc.ClientUnaryCall;
    public ping(request: FileSystemService_pb.PingRequest, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.PongResponse) => void): grpc.ClientUnaryCall;
    public ping(request: FileSystemService_pb.PingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.PongResponse) => void): grpc.ClientUnaryCall;
    public ping(request: FileSystemService_pb.PingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: FileSystemService_pb.PongResponse) => void): grpc.ClientUnaryCall;
}
