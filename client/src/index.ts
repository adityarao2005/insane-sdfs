import { IFileSystemService } from "./filesystem/filesystem"
import { Command } from "commander"
import { color } from "console-log-colors"
import { intro, outro, text, select, note, isCancel } from '@clack/prompts';
import {
    connectToFileSystemService,
    createDefaultCliActionDeps,
    createDirectory,
    deleteFile,
    downloadFile,
    listFiles,
    uploadFile
} from "./cli/actions";

const clients: Map<string, IFileSystemService> = new Map()

const program = new Command()

program.version("1.0.0")
    .description("A CLI for interacting with a remote file system service to upload and download files.")
    .parse(process.argv)

program.opts()

const actionDeps = createDefaultCliActionDeps({
    text,
    isCancel,
    note: (message) => note(message)
})

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
                await connectToFileSystemService(clients, actionDeps)
                break;
            case "upload":
                await uploadFile(clients, actionDeps)
                break;
            case "download":
                await downloadFile(clients, actionDeps)
                break;
            case "list":
                await listFiles(clients, actionDeps)
                break;
            case "delete":
                await deleteFile(clients, actionDeps)
                break;
            case "mkdir":
                await createDirectory(clients, actionDeps)
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