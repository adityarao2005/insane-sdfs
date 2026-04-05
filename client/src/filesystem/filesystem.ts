import { ChannelCredentials } from "@grpc/grpc-js";
import { FileSystemClient } from "../proto/FileSystemService_grpc_pb"
import { DownloadFileRequest, FileInfoRequest, UploadFileRequestFragment } from "../proto/FileSystemService_pb";
import { gRPC } from "./grpc.filesystem";

export interface IFileSystemService extends Disposable {

    uploadFile(path: string, data: AsyncGenerator<Uint8Array>): Promise<void>;

    downloadFile(path: string): AsyncGenerator<Uint8Array>;

    deleteFile(path: string): Promise<void>;

    getFileInfo(path: string): Promise<{ size: number; isDirectory: boolean; modifiedAt: Date }>;

    listFiles(directoryPath: string): Promise<string[]>;

    createDirectory(directoryPath: string): Promise<void>;

}

export type FileSystemServiceOptions = {
    host: string;
    port: number;
    path?: string;
    authentication?: string;
}

export interface IFileSystemServiceProvider {
    getFileSystemService(options?: FileSystemServiceOptions): Promise<IFileSystemService>;
    protocol: string;
}


export function getFileSystemServiceProviders(): IFileSystemServiceProvider[] {
    return [
        gRPC.createGrpcFileServiceProvider()
    ];
}

export function getFileSystemProvider(protocol: string): IFileSystemServiceProvider | null {
    const providers = getFileSystemServiceProviders();
    return providers.find(provider => provider.protocol === protocol) || null;
}