import { ChannelCredentials } from "@grpc/grpc-js";
import { FileSystemClient } from "../proto/FileSystemService_grpc_pb";
import { DownloadFileRequest, FileInfoRequest, PingRequest, UploadFileRequestFragment } from "../proto/FileSystemService_pb";
import { FileSystemServiceOptions, IFileSystemService, IFileSystemServiceProvider } from "./filesystem";

export namespace gRPC {
    class FileSystemServiceClient implements IFileSystemService {

        private grpcClient: FileSystemClient

        constructor(options: FileSystemServiceOptions) {
            this.grpcClient = new FileSystemClient(`${options.host}:${options.port}`, ChannelCredentials.createInsecure());
        }

        async ensureConnected(): Promise<void> {
            return new Promise((resolve, reject) => {
                this.grpcClient.ping(new PingRequest().setMessage("ping"), (err, response) => {
                    if (err) {
                        reject(err);
                    } else {
                        resolve();
                    }
                });
            });
        }

        [Symbol.dispose](): void {
            this.grpcClient.close();
        }

        uploadFile(path: string, data: AsyncGenerator<Uint8Array>): Promise<void> {
            const promise = new Promise<void>(async (resolve, reject) => {
                const writableStream = this.grpcClient.uploadFile((err, response) => {
                    if (err) {
                        reject(err);
                    } else {
                        resolve();
                    }
                });

                for await (const chunk of data) {
                    writableStream.write(new UploadFileRequestFragment().setPath(path).setData(chunk));
                }

                writableStream.end();
            });

            return promise;
        }

        async *downloadFile(path: string): AsyncGenerator<Uint8Array> {
            const stream = this.grpcClient.downloadFile(new DownloadFileRequest().setPath(path));

            for await (const response of stream) {
                yield response.getData();
            }
        }

        deleteFile(path: string): Promise<void> {
            return new Promise((resolve, reject) => {
                this.grpcClient.deleteFile(new FileInfoRequest().setPath(path), (err) => {
                    if (err) {
                        reject(err);
                    } else {
                        resolve();
                    }
                });
            });
        }

        getFileInfo(path: string): Promise<{ size: number; isDirectory: boolean; modifiedAt: Date }> {
            return new Promise((resolve, reject) => {
                this.grpcClient.getFileInfo(new FileInfoRequest().setPath(path), (err, response) => {
                    if (err) {
                        reject(err);
                    } else {
                        resolve({
                            size: response.getSize(),
                            isDirectory: response.getIsDirectory(),
                            modifiedAt: new Date(response.getLastModifiedTime())
                        });
                    }
                });
            });
        }

        listFiles(directoryPath: string): Promise<string[]> {
            return new Promise((resolve, reject) => {
                const values = this.grpcClient.listFiles(new FileInfoRequest().setPath(directoryPath));

                const fileNames: string[] = [];

                values.on("data", (response) => {
                    fileNames.push(response.getName());
                });

                values.on("end", () => {
                    resolve(fileNames);
                });

                values.on("error", (err) => {
                    reject(err);
                });
            });
        }

        createDirectory(directoryPath: string): Promise<void> {
            return new Promise((resolve, reject) => {
                this.grpcClient.createDirectory(new FileInfoRequest().setPath(directoryPath), (err) => {
                    if (err) {
                        reject(err);
                    } else {
                        resolve();
                    }
                });
            });
        }

    }

    export function createGrpcFileServiceProvider(): IFileSystemServiceProvider {
        return {
            getFileSystemService: async (opts?: FileSystemServiceOptions) => {
                const client = new FileSystemServiceClient(opts || { host: "localhost", port: 8080 });
                await client.ensureConnected();
                return client;
            },
            protocol: "grpc"
        }
    }
}
