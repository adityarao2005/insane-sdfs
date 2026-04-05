import { IFileSystemService } from "./filesystem.service"
import { Command } from "commander"
import { color } from "console-log-colors"
import { intro, outro, text, select, spinner, note } from '@clack/prompts';

const address = process.env.GRPC_SERVER_ADDRESS || "localhost:8080"

const clients: IFileSystemService[] = []

const program = new Command()

program.version("1.0.0")
    .description("A CLI for interacting with a remote file system service to upload and download files.")
    .parse(process.argv)

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
                const address = await text({ message: color.yellow(`You selected: ${choice}. Enter the file system address to connect to (e.g grpc://localhost:8080):`) })

                note(color.green(`Connecting to file system service at ${address.toString()}...`))

                break;
            case "upload":
                break;
            case "download":

                break;
            case "list":
                break;
            case "delete":
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
}

async function main() {
    const options = program.opts()

    await homeScreen()
}

await main()