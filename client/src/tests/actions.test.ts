import { describe, expect, mock, test } from "bun:test"
import type { IFileSystemService } from "../filesystem/filesystem"
import {
    type CliActionDeps,
    connectToFileSystemService,
    createDirectory,
    deleteFile,
    downloadFile,
    listFiles,
    uploadFile
} from "../cli/actions"

const CANCEL = Symbol("cancel")

function makeClient(overrides: Partial<IFileSystemService> = {}): IFileSystemService {
    const base: IFileSystemService = {
        async uploadFile() {
            return
        },
        async *downloadFile() {
            yield new Uint8Array()
        },
        async deleteFile() {
            return
        },
        async getFileInfo() {
            return { size: 1, isDirectory: false, modifiedAt: new Date() }
        },
        async listFiles() {
            return []
        },
        async createDirectory() {
            return
        },
        [Symbol.dispose]() {
            return
        }
    }

    return { ...base, ...overrides }
}

function makeDeps(overrides: Partial<CliActionDeps> = {}): { deps: CliActionDeps; notes: string[] } {
    const notes: string[] = []

    const deps: CliActionDeps = {
        text: async () => "",
        isCancel: (value) => value === CANCEL,
        note: (message) => {
            notes.push(message)
        },
        selectClient: async () => null,
        selectLocalFile: async () => "",
        selectRemoteFile: async () => "",
        selectRemoteDirectory: async () => "",
        getFileSystemProvider: () => null,
        fileFromPath: () => ({
            stream: () => new ReadableStream<Uint8Array>(),
            writer: () => ({
                async write() {
                    return
                },
                async end() {
                    return
                }
            })
        }),
        async *streamToGenerator(stream) {
            const reader = stream.getReader()
            try {
                while (true) {
                    const { done, value } = await reader.read()
                    if (done || !value) {
                        break
                    }
                    yield value
                }
            } finally {
                reader.releaseLock()
            }
        },
        ...overrides
    }

    return { deps, notes }
}

describe("connectToFileSystemService", () => {
    test("rejects malformed address", async () => {
        const { deps, notes } = makeDeps({ text: async () => "bad-address" })
        const clients = new Map<string, IFileSystemService>()

        await connectToFileSystemService(clients, deps)

        expect(clients.size).toBe(0)
        expect(notes.at(-1)).toContain("Invalid address format")
    })

    test("stores client after duplicate alias retry", async () => {
        const client = makeClient()
        const textAnswers = ["grpc://localhost:8080", "existing", "fresh-alias"]
        const provider = {
            protocol: "grpc",
            getFileSystemService: async () => client
        }

        const { deps, notes } = makeDeps({
            text: async () => textAnswers.shift() ?? "",
            getFileSystemProvider: () => provider
        })
        const clients = new Map<string, IFileSystemService>([["existing", makeClient()]])

        await connectToFileSystemService(clients, deps)

        expect(clients.get("fresh-alias")).toBe(client)
        expect(notes).toContain("A client with this alias already exists. Please choose a different alias.")
    })
})

describe("uploadFile", () => {
    test("uploads selected file to selected remote path", async () => {
        const uploadedChunks: Uint8Array[] = []
        let uploadedPath = ""
        const client = makeClient({
            uploadFile: async (path, data) => {
                uploadedPath = path
                for await (const chunk of data) {
                    uploadedChunks.push(chunk)
                }
            }
        })

        const testChunk = new Uint8Array([1, 2, 3])
        const { deps, notes } = makeDeps({
            selectClient: async () => client,
            selectLocalFile: async () => "/tmp/local.txt",
            selectRemoteFile: async () => "/remote/out.txt",
            fileFromPath: () => ({
                stream: () =>
                    new ReadableStream<Uint8Array>({
                        start(controller) {
                            controller.enqueue(testChunk)
                            controller.close()
                        }
                    }),
                writer: () => {
                    throw new Error("writer should not be used in upload")
                }
            })
        })

        const clients = new Map<string, IFileSystemService>([["default", client]])
        await uploadFile(clients, deps)

        expect(uploadedPath).toBe("/remote/out.txt")
        expect(uploadedChunks).toEqual([testChunk])
        expect(notes.at(-1)).toContain("File uploaded successfully")
    })

    test("exits when client selection is canceled", async () => {
        const { deps, notes } = makeDeps({
            selectClient: async () => CANCEL
        })

        await uploadFile(new Map(), deps)
        expect(notes.at(-1)).toBe("Selection canceled.")
    })
})

describe("downloadFile", () => {
    test("writes all stream chunks and ends writer", async () => {
        const chunks = [new Uint8Array([10]), new Uint8Array([20, 30])]
        const writes: Uint8Array[] = []
        let ended = false

        const client = makeClient({
            async *downloadFile() {
                yield chunks[0]
                yield chunks[1]
            }
        })

        const { deps, notes } = makeDeps({
            selectClient: async () => client,
            selectRemoteFile: async () => "/remote/in.bin",
            selectLocalFile: async () => "/tmp/in.bin",
            fileFromPath: () => ({
                stream: () => {
                    throw new Error("stream should not be used in download")
                },
                writer: () => ({
                    async write(chunk) {
                        writes.push(chunk)
                    },
                    async end() {
                        ended = true
                    }
                })
            })
        })

        await downloadFile(new Map([["default", client]]), deps)

        expect(writes).toEqual(chunks)
        expect(ended).toBe(true)
        expect(notes.at(-1)).toContain("File downloaded successfully")
    })
})

describe("list/delete/create", () => {
    test("lists files for selected directory", async () => {
        const listFilesSpy = mock(async () => ["a.txt", "b.txt"])
        const client = makeClient({ listFiles: listFilesSpy })
        const { deps, notes } = makeDeps({
            selectClient: async () => client,
            selectRemoteDirectory: async () => "/shared"
        })

        await listFiles(new Map([["default", client]]), deps)

        expect(listFilesSpy).toHaveBeenCalledWith("/shared")
        expect(notes.at(-1)).toBe("a.txt\nb.txt")
    })

    test("deletes selected remote path", async () => {
        const deleteSpy = mock(async () => undefined)
        const client = makeClient({ deleteFile: deleteSpy })
        const { deps, notes } = makeDeps({
            selectClient: async () => client,
            selectRemoteFile: async () => "/shared/old.txt"
        })

        await deleteFile(new Map([["default", client]]), deps)

        expect(deleteSpy).toHaveBeenCalledWith("/shared/old.txt")
        expect(notes.at(-1)).toContain("deleted successfully")
    })

    test("creates selected remote directory", async () => {
        const mkdirSpy = mock(async () => undefined)
        const client = makeClient({ createDirectory: mkdirSpy })
        const { deps, notes } = makeDeps({
            selectClient: async () => client,
            selectRemoteDirectory: async () => "/shared/new-dir"
        })

        await createDirectory(new Map([["default", client]]), deps)

        expect(mkdirSpy).toHaveBeenCalledWith("/shared/new-dir")
        expect(notes.at(-1)).toContain("created successfully")
    })
})