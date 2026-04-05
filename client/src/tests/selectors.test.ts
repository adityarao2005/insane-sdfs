import { afterAll, beforeEach, describe, expect, mock, test } from "bun:test"
import type { IFileSystemService } from "../filesystem/filesystem"

const CANCEL = Symbol("cancel")

const textMock = mock(async () => "")
const pathPromptMock = mock(async () => "")
const selectPromptMock = mock(async () => "")
const noteMock = mock((_message: string) => { })

mock.module("@clack/prompts", () => ({
    text: textMock,
    path: pathPromptMock,
    select: selectPromptMock,
    note: noteMock,
    isCancel: (value: unknown) => value === CANCEL
}))

const selectors = await import("../cli/selectors")

type FileMetadata = {
    exists: boolean
    isDirectory: boolean
}

const fileTable = new Map<string, FileMetadata>()

const originalBunFile = Bun.file

    ; (Bun as { file: typeof Bun.file }).file = ((filePath: string) => ({
        async exists() {
            return fileTable.get(filePath)?.exists ?? false
        },
        async stat() {
            const file = fileTable.get(filePath)
            return {
                isDirectory() {
                    return file?.isDirectory ?? false
                }
            }
        }
    })) as typeof Bun.file

afterAll(() => {
    ; (Bun as { file: typeof Bun.file }).file = originalBunFile
})

function makeClient(overrides: Partial<IFileSystemService> = {}): IFileSystemService {
    const base: IFileSystemService = {
        async uploadFile() {
            return
        },
        async *downloadFile() {
            yield new Uint8Array([1])
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

beforeEach(() => {
    textMock.mockReset()
    pathPromptMock.mockReset()
    selectPromptMock.mockReset()
    noteMock.mockReset()
    fileTable.clear()
})

describe("selectLocalFile", () => {
    test("retries until an existing file is chosen", async () => {
        pathPromptMock
            .mockResolvedValueOnce("/tmp/missing.txt")
            .mockResolvedValueOnce("/tmp/found.txt")

        fileTable.set("/tmp/missing.txt", { exists: false, isDirectory: false })
        fileTable.set("/tmp/found.txt", { exists: true, isDirectory: false })

        const selected = await selectors.selectLocalFile({
            expectExists: true,
            allowsDirectories: false,
            prompt: "Select file"
        })

        expect(selected).toBe("/tmp/found.txt")
        const messages = noteMock.mock.calls.map((call) => String(call[0]))
        expect(messages.some((message) => message.includes("Selected file does not exist. Please select a valid file."))).toBe(true)
    })
})

describe("selectRemoteFile", () => {
    test("returns file path when file exists and is not directory", async () => {
        textMock.mockResolvedValueOnce("/remote/file.txt")
        const client = makeClient({
            getFileInfo: async () => ({ size: 10, isDirectory: false, modifiedAt: new Date() })
        })

        const selected = await selectors.selectRemoteFile(client, {
            expectExists: true,
            allowsDirectories: false,
            prompt: "Select remote file"
        })

        expect(selected).toBe("/remote/file.txt")
    })

    test("accepts new file when parent directory exists", async () => {
        textMock.mockResolvedValueOnce("/remote/new.txt")
        const client = makeClient({
            getFileInfo: async (targetPath: string) => {
                if (targetPath === "/remote/new.txt") {
                    throw new Error("not found")
                }
                return { size: 0, isDirectory: true, modifiedAt: new Date() }
            }
        })

        const selected = await selectors.selectRemoteFile(client, {
            expectExists: false,
            allowsDirectories: false,
            prompt: "Select remote file"
        })

        expect(selected).toBe("/remote/new.txt")
    })
})

describe("selectClient", () => {
    test("returns null when no clients are available", async () => {
        const selected = await selectors.selectClient(new Map())
        expect(selected).toBeNull()
    })
})