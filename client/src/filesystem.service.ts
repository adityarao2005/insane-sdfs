

export interface IFileSystemService {

    uploadFile(path: string, data: Generator<Uint8Array, void, unknown>): Promise<void>;

    downloadFile(path: string): Promise<Generator<Uint8Array, void, unknown>>;

    deleteFile(path: string): Promise<void>;

    getFileInfo(path: string): Promise<{ size: number; createdAt: Date; modifiedAt: Date }>;

    listFiles(directoryPath: string): Promise<string[]>;

    createDirectory(directoryPath: string): Promise<void>;
}

export type FileSystemServiceOptions = {
    host: string;
    port: number;
    path: string;
    authentication?: string;
}

export interface IFileSystemServiceProvider {
    getFileSystemService(options?: FileSystemServiceOptions): IFileSystemService;
}