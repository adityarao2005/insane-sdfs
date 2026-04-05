import { IFileSystemService, getFileSystemProvider } from "../filesystem/filesystem"
import { selectClient, selectLocalFile, selectRemoteDirectory, selectRemoteFile } from "./selectors"
import { streamToGenerator } from "./stream-utils"

interface Writer {
    write(chunk: Uint8Array): number | Promise<number> | Promise<void>
    end(): number | Promise<number> | Promise<void>
}

interface FileHandle {
    stream(): ReadableStream<Uint8Array>
    writer(): Writer
}

export type CliActionDeps = {
    text: (options: { message: string; defaultValue?: string }) => Promise<unknown>
    isCancel: (value: unknown) => boolean
    note: (message: string) => void
    selectClient: typeof selectClient
    selectLocalFile: typeof selectLocalFile
    selectRemoteFile: typeof selectRemoteFile
    selectRemoteDirectory: typeof selectRemoteDirectory
    getFileSystemProvider: typeof getFileSystemProvider
    fileFromPath: (filePath: string) => FileHandle
    streamToGenerator: typeof streamToGenerator
}

export function createDefaultCliActionDeps(
    input: Pick<CliActionDeps, "text" | "isCancel" | "note">
): CliActionDeps {
    return {
        ...input,
        selectClient,
        selectLocalFile,
        selectRemoteFile,
        selectRemoteDirectory,
        getFileSystemProvider,
        fileFromPath: (filePath) => Bun.file(filePath),
        streamToGenerator
    }
}

function asString(value: unknown): string {
    return value === null || value === undefined ? "" : value.toString()
}

function describeError(err: unknown): string {
    return err instanceof Error ? err.message : String(err)
}

function isFileSystemService(value: unknown): value is IFileSystemService {
    return typeof value === "object" && value !== null && "uploadFile" in value
}

export async function connectToFileSystemService(
    clients: Map<string, IFileSystemService>,
    deps: CliActionDeps
) {
    const address = await deps.text({ message: "Enter the address of the file system service (format: {protocol}://{host}:{port}):" })
    if (deps.isCancel(address)) {
        deps.note("Selection canceled.")
        return
    }

    const [protocol, hostPort] = asString(address).split("://")
    if (!protocol || !hostPort) {
        deps.note("Invalid address format. Please use the format: {protocol}://{host}:{port}")
        return
    }

    const fileSystemProvider = deps.getFileSystemProvider(protocol)
    if (!fileSystemProvider) {
        deps.note(`Unsupported protocol: ${protocol}`)
        return
    }

    const [host, port] = hostPort.split(":")
    if (!host || !port) {
        deps.note("Invalid address format. Please use the format: {protocol}://{host}:{port}")
        return
    }

    let client: IFileSystemService
    try {
        client = await fileSystemProvider.getFileSystemService({ host, port: parseInt(port, 10) })
    } catch (err) {
        deps.note(`Failed to connect to file system service at ${asString(address)}: ${describeError(err)}`)
        return
    }

    while (true) {
        const clientKey = await deps.text({ message: "Enter the alias for this client:", defaultValue: `${protocol}://${host}:${port}` })
        if (deps.isCancel(clientKey)) {
            deps.note("Selection canceled.")
            return
        }

        const alias = asString(clientKey)
        if (clients.has(alias)) {
            deps.note("A client with this alias already exists. Please choose a different alias.")
            continue
        }

        clients.set(alias, client)
        return
    }
}

export async function uploadFile(clients: Map<string, IFileSystemService>, deps: CliActionDeps) {
    const client = await deps.selectClient(clients)
    if (deps.isCancel(client)) {
        deps.note("Selection canceled.")
        return
    }

    if (!isFileSystemService(client)) {
        deps.note("No client found")
        return
    }

    const localFile = await deps.selectLocalFile({
        expectExists: true,
        allowsDirectories: false,
        prompt: "Select the file to upload:"
    })
    if (deps.isCancel(localFile)) {
        deps.note("Selection canceled.")
        return
    }
    if (!localFile) {
        deps.note("No local file selected.")
        return
    }

    const remotePath = await deps.selectRemoteFile(client, {
        expectExists: false,
        allowsDirectories: false,
        prompt: "Enter the remote path where the file should be uploaded:"
    })
    if (deps.isCancel(remotePath)) {
        deps.note("Selection canceled.")
        return
    }
    if (!remotePath) {
        deps.note("No remote file path selected.")
        return
    }

    try {
        const file = deps.fileFromPath(asString(localFile))
        const stream = file.stream()
        await client.uploadFile(asString(remotePath), deps.streamToGenerator(stream))
        deps.note(`File uploaded successfully to ${asString(remotePath)}`)
    } catch (err) {
        deps.note(`Failed to upload file: ${describeError(err)}`)
    }
}

export async function downloadFile(clients: Map<string, IFileSystemService>, deps: CliActionDeps) {
    const client = await deps.selectClient(clients)
    if (deps.isCancel(client) || !isFileSystemService(client)) {
        deps.note("Selection canceled.")
        return
    }

    const remoteFile = await deps.selectRemoteFile(client, {
        expectExists: true,
        allowsDirectories: false,
        prompt: "Enter the remote path of the file to download:"
    })
    if (deps.isCancel(remoteFile)) {
        deps.note("Selection canceled.")
        return
    }
    if (!remoteFile) {
        deps.note("No remote file selected.")
        return
    }

    const localFile = await deps.selectLocalFile({
        expectExists: false,
        allowsDirectories: false,
        prompt: "Enter the local path where the file should be downloaded:"
    })
    if (deps.isCancel(localFile)) {
        deps.note("Selection canceled.")
        return
    }
    if (!localFile) {
        deps.note("No local file path selected.")
        return
    }

    try {
        const fileStream = client.downloadFile(asString(remoteFile))
        const writer = deps.fileFromPath(asString(localFile)).writer()

        for await (const chunk of fileStream) {
            await writer.write(chunk)
        }

        await writer.end()
        deps.note(`File downloaded successfully to ${asString(localFile)}`)
    } catch (err) {
        deps.note(`Failed to download file: ${describeError(err)}`)
    }
}

export async function listFiles(clients: Map<string, IFileSystemService>, deps: CliActionDeps) {
    const client = await deps.selectClient(clients)
    if (deps.isCancel(client) || !isFileSystemService(client)) {
        deps.note("Selection canceled.")
        return
    }

    const directory = await deps.selectRemoteDirectory(client)
    if (deps.isCancel(directory)) {
        deps.note("Selection canceled.")
        return
    }
    if (!directory) {
        deps.note("No directory selected.")
        return
    }

    const files = await client.listFiles(asString(directory))
    deps.note(files.join("\n"))
}

export async function deleteFile(clients: Map<string, IFileSystemService>, deps: CliActionDeps) {
    const client = await deps.selectClient(clients)
    if (deps.isCancel(client) || !isFileSystemService(client)) {
        deps.note("Selection canceled.")
        return
    }

    const remoteFile = await deps.selectRemoteFile(client, {
        expectExists: true,
        allowsDirectories: true,
        prompt: "Enter the remote path of the file to delete:"
    })
    if (deps.isCancel(remoteFile)) {
        deps.note("Selection canceled.")
        return
    }
    if (!remoteFile) {
        deps.note("No remote file selected.")
        return
    }

    await client.deleteFile(asString(remoteFile))
    deps.note(`File ${asString(remoteFile)} deleted successfully.`)
}

export async function createDirectory(clients: Map<string, IFileSystemService>, deps: CliActionDeps) {
    const client = await deps.selectClient(clients)
    if (deps.isCancel(client) || !isFileSystemService(client)) {
        deps.note("Selection canceled.")
        return
    }

    const remoteDirectory = await deps.selectRemoteDirectory(client, false)
    if (deps.isCancel(remoteDirectory)) {
        deps.note("Selection canceled.")
        return
    }
    if (!remoteDirectory) {
        deps.note("No remote directory selected.")
        return
    }

    await client.createDirectory(asString(remoteDirectory))
    deps.note(`Directory ${asString(remoteDirectory)} created successfully.`)
}