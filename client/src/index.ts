import { getFileSystemProvider, getFileSystemServiceProviders, IFileSystemService } from "./filesystem/filesystem"
import { Command } from "commander"
import { color } from "console-log-colors"
import { intro, outro, text, select, spinner, note } from '@clack/prompts';

const clients: IFileSystemService[] = []

const program = new Command()

program.version("1.0.0")
    .description("A CLI for interacting with a remote file system service to upload and download files.")
    .parse(process.argv)

async function connectToFileSystemService() {
    // expects it to be in this format: "{protocol}://{host}:{port}"
    // example: "grpc://localhost:8080"
    const address = await text({ message: "Enter the address of the file system service (format: {protocol}://{host}:{port}):" })
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
        clients.push(client)
    } catch (err) {
        note(color.red(`Failed to connect to file system service at ${address.toString()}: ${err instanceof Error ? err.message : String(err)}`))
        return
    }

}

async function uploadFile() {

}

async function downloadFile() {

}

async function listFiles() {

}

async function deleteFile() {

}

async function homeScreen() {
    intro(color.cyan("Welcome to the File System CLI!"))

    outer:
    while (true) {
        note(color.yellow("This CLI allows you to interact with a remote file system service to upload and download files. Here are the connected file clients:"))

        if (clients.length === 0) {
            note(color.red("No clients connected. Please connect to a file system service to get started."))
        } else {
            note(color.green(clients.map((_, index) => `Client ${index + 1}`).join("\n"))) // Display connected clients
        }

        type Option = "connect" | "upload" | "download" | "list" | "delete" | "exit"

        const options: { label: string; value: Option }[] = [
            { label: "Connect to File System Service", value: "connect" },
            { label: "Upload a File", value: "upload" },
            { label: "Download a File", value: "download" },
            { label: "List Files in Directory", value: "list" },
            { label: "Delete a File", value: "delete" },
            { label: "Exit", value: "exit" }
        ]

        const choice = await select({
            message: color.cyan("Please select an option:"),
            options
        })

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
            case "exit":
                outro(color.yellow(`You selected: ${choice}. Have a nice evening!`))
                break outer
        }

    }


    while (clients.length > 0) {
        const client = clients.pop()

        client?.[Symbol.dispose]()
    }

    process.exit(0)
}

async function main() {
    const options = program.opts()

    await homeScreen()
}

await main()