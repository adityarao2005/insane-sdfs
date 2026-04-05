import { getFileSystemProvider, getFileSystemServiceProviders, IFileSystemService } from "./filesystem/filesystem"
import { Command } from "commander"
import { color } from "console-log-colors"
import { intro, outro, text, select, spinner, note, path, isCancel } from '@clack/prompts';
import { selectClient, selectLocalFile, selectRemoteDirectory, selectRemoteFile } from "./cmdline/selectors";
import { streamToGenerator } from "./cmdline/stream-utils";

const clients: Map<string, IFileSystemService> = new Map()


const program = new Command()

program.version("1.0.0")
    .description("A CLI for interacting with a remote file system service to upload and download files.")
    .parse(process.argv)

const options = program.opts()

async function connectToFileSystemService() {
    // expects it to be in this format: "{protocol}://{host}:{port}"
    // example: "grpc://localhost:8080"
    const address = await text({ message: "Enter the address of the file system service (format: {protocol}://{host}:{port}):" })
    if (isCancel(address)) {
        note(color.red("Selection canceled."))
        return
    }

    const [protocol, hostPort] = address.toString().split("://")
    if (!protocol || !hostPort) {
        note(color.red("Invalid address format. Please use the format: {protocol}://{host}:{port}"))
        return
    }

    const fileSystemProvider = getFileSystemProvider(protocol)
    if (!fileSystemProvider) {
        note(color.red(`Unsupported protocol: ${protocol}`))
        return
    }

    const [host, port] = hostPort.split(":")
    if (!host || !port) {
        note(color.red("Invalid address format. Please use the format: {protocol}://{host}:{port}"))
        return
    }

    let client: IFileSystemService;
    try {
        client = await fileSystemProvider.getFileSystemService({ host, port: parseInt(port) })
    } catch (err) {
        note(color.red(`Failed to connect to file system service at ${address.toString()}: ${err instanceof Error ? err.message : String(err)}`))
        return
    }


    while (true) {
        const clientKey = await text({ message: "Enter the alias for this client:", defaultValue: `${protocol}://${host}:${port}` })
        if (isCancel(clientKey)) {
            note(color.red("Selection canceled."))
            return
        }

        if (clients.has(clientKey.toString())) {
            note(color.red("A client with this alias already exists. Please choose a different alias."))
        } else {
            clients.set(clientKey.toString(), client)
            break
        }
    }

}


async function uploadFile() {
    const client = await selectClient(clients)
    if (isCancel(client)) {
        note(color.red("Selection canceled."))
        return
    }

    if (!client) {
        note(color.red(`No client found`))
        return
    }
    const localFile = await selectLocalFile({
        expectExists: true,
        allowsDirectories: false,
        prompt: "Select the file to upload:"
    })
    if (isCancel(localFile)) {
        note(color.red("Selection canceled."))
        return
    }
    if (!localFile) {
        note(color.red("No local file selected."))
        return
    }

    const remotePath = await selectRemoteFile(client, {
        expectExists: false,
        allowsDirectories: false,
        prompt: "Enter the remote path where the file should be uploaded:"
    })
    if (isCancel(remotePath)) {
        note(color.red("Selection canceled."))
        return
    }
    if (!remotePath) {
        note(color.red("No remote file path selected."))
        return
    }

    try {
        const file = Bun.file(localFile.toString())
        const stream = file.stream()
        await client.uploadFile(remotePath.toString(), streamToGenerator(stream))
        note(color.green(`File uploaded successfully to ${remotePath.toString()}`))
    } catch (err) {
        note(color.red(`Failed to upload file: ${err instanceof Error ? err.message : String(err)}`))
    }

}

async function downloadFile() {
    const client = await selectClient(clients)
    if (isCancel(client) || !client) {
        note(color.red("Selection canceled."))
        return
    }

    const remoteFile = await selectRemoteFile(client, {
        expectExists: true,
        allowsDirectories: false,
        prompt: "Enter the remote path of the file to download:"
    })
    if (isCancel(remoteFile)) {
        note(color.red("Selection canceled."))
        return
    }
    if (!remoteFile) {
        note(color.red("No remote file selected."))
        return
    }

    const localFile = await selectLocalFile({
        expectExists: false,
        allowsDirectories: false,
        prompt: "Enter the local path where the file should be downloaded:"
    })
    if (isCancel(localFile)) {
        note(color.red("Selection canceled."))
        return
    }
    if (!localFile) {
        note(color.red("No local file path selected."))
        return
    }

    try {
        const fileStream = client.downloadFile(remoteFile.toString())
        const writer = Bun.file(localFile.toString()).writer()

        for await (const chunk of fileStream) {
            await writer.write(chunk)
        }

        await writer.end()
        note(color.green(`File downloaded successfully to ${localFile.toString()}`))
    } catch (err) {
        note(color.red(`Failed to download file: ${err instanceof Error ? err.message : String(err)}`))
    }

}

async function listFiles() {
    const client = await selectClient(clients)
    if (isCancel(client) || !client) {
        note(color.red("Selection canceled."))
        return
    }

    const directory = await selectRemoteDirectory(client)
    if (isCancel(directory)) {
        note(color.red("Selection canceled."))
        return
    }
    if (!directory) {
        note(color.red("No directory selected."))
        return
    }

    const files = await client.listFiles(directory.toString())
    note(color.green(files.join("\n")))

}

async function deleteFile() {
    const client = await selectClient(clients)
    if (isCancel(client) || !client) {
        note(color.red("Selection canceled."))
        return
    }

    const remoteFile = await selectRemoteFile(client, {
        expectExists: true,
        allowsDirectories: true,
        prompt: "Enter the remote path of the file to delete:"
    })
    if (isCancel(remoteFile)) {
        note(color.red("Selection canceled."))
        return
    }
    if (!remoteFile) {
        note(color.red("No remote file selected."))
        return
    }

    await client.deleteFile(remoteFile.toString())
    note(color.green(`File ${remoteFile.toString()} deleted successfully.`))

}

async function createDirectory() {
    const client = await selectClient(clients)
    if (isCancel(client) || !client) {
        note(color.red("Selection canceled."))
        return
    }

    const remoteDirectory = await selectRemoteDirectory(client, false)
    if (isCancel(remoteDirectory)) {
        note(color.red("Selection canceled."))
        return
    }
    if (!remoteDirectory) {
        note(color.red("No remote directory selected."))
        return
    }

    await client.createDirectory(remoteDirectory.toString())
    note(color.green(`Directory ${remoteDirectory.toString()} created successfully.`))

}

async function homeScreen() {
    intro(color.cyan("Welcome to the File System CLI!"))

    outer:
    while (true) {
        note(color.yellow("This CLI allows you to interact with a remote file system service to upload and download files. Here are the connected file clients:"))

        if (clients.size === 0) {
            note(color.red("No clients connected. Please connect to a file system service to get started."))
        } else {
            note(color.green(Array.from(clients.keys()).join("\n"))) // Display connected clients
        }

        type Option = "connect" | "upload" | "download" | "list" | "delete" | "mkdir" | "exit"

        const options: { label: string; value: Option }[] = [
            { label: "Connect to File System Service", value: "connect" },
            { label: "Upload a File", value: "upload" },
            { label: "Download a File", value: "download" },
            { label: "List Files in Directory", value: "list" },
            { label: "Delete a File", value: "delete" },
            { label: "Create a directory", value: "mkdir" },
            { label: "Exit", value: "exit" }
        ]

        const choice = await select({
            message: color.cyan("Please select an option:"),
            options
        })

        if (isCancel(choice)) {
            outro(color.yellow("No option selected. Exiting. Have a nice evening!"))
            break
        }

        switch (choice) {
            case "connect":
                await connectToFileSystemService()
                break;
            case "upload":
                await uploadFile()
                break;
            case "download":
                await downloadFile()
                break;
            case "list":
                await listFiles()
                break;
            case "delete":
                await deleteFile()
                break;
            case "mkdir":
                await createDirectory()
                break;
            case "exit":
                outro(color.yellow(`You selected: ${choice}. Have a nice evening!`))
                break outer
        }

    }


    while (clients.size > 0) {
        const entry = clients.entries().next().value

        if (!entry) {
            break
        }

        using _ = entry[1]
        clients.delete(entry[0])
    }

    process.exit(0)
}

await homeScreen()