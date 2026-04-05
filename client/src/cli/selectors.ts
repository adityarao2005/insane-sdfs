import { isCancel, note, path as pathPrompt, select, text } from "@clack/prompts"
import { color } from "console-log-colors"
import { IFileSystemService } from "../filesystem/filesystem"
import path from "path"

type SelectOption = {
    expectExists?: boolean
    allowsDirectories?: boolean
    prompt: string
}

export async function selectLocalFile(options: SelectOption) {
    const { expectExists = true, allowsDirectories = false } = options;

    while (true) {

        const filePath = expectExists ? await pathPrompt({ message: options.prompt, directory: false }) : await text({ message: options.prompt })
        if (isCancel(filePath)) {
            note(color.red("Selection canceled."))
            return filePath
        }

        if (!filePath) {
            note(color.red("No file selected."))
            return
        }

        const file = Bun.file(filePath.toString())

        const fileExists = await file.exists()
        if (expectExists && !fileExists) {
            note(color.red("Selected file does not exist. Please select a valid file."))
            continue;
        } else if (fileExists) {

            const stat = await file.stat()

            if (stat.isDirectory() && !allowsDirectories) {
                note(color.red("Selected path is a directory. Please select a file."))
                continue;
            }

            return filePath
        } else if (!expectExists) {
            const parentDir = Bun.file(path.dirname(filePath.toString()))

            try {
                await parentDir.stat()
            } catch (err) {
                note(color.red(`Parent directory of the selected file does not exist: ${path.dirname(filePath.toString())}. Please select a valid path.`))
                continue;
            }

            return filePath
        }
    }
}

export async function selectRemoteFile(client: IFileSystemService, options: SelectOption) {
    const { expectExists = true, allowsDirectories = false } = options;

    while (true) {
        const filePath = await text({ message: options.prompt })
        if (isCancel(filePath)) {
            note(color.red("Selection canceled."))
            return filePath
        }

        if (!filePath) {
            note(color.red("No file path entered."))
            return
        }

        try {
            const fileInfo = await client.getFileInfo(filePath.toString())

            if (fileInfo.isDirectory && !allowsDirectories) {
                note(color.red("Selected path is a directory. Please select a file."))
                continue;
            }

            return filePath
        } catch (err) {
            if (expectExists) {
                note(color.red(`Failed to get file info for ${filePath.toString()}: ${err instanceof Error ? err.message : String(err)}`))
                continue;
            }

            try {
                // if the call succeeds, it means the parent directory exists, which is good enough for our use case since we will be creating the file during upload
                await client.getFileInfo(path.dirname(filePath.toString()));
                return filePath
            } catch (err) {
                note(color.red(`Failed to get file info for ${filePath.toString()}: ${err instanceof Error ? err.message : String(err)}`))
                continue;
            }
        }
    }
}

export async function selectRemoteDirectory(client: IFileSystemService, expectExists: boolean = true) {

    while (true) {
        const filePath = await text({ message: "Enter the path of the directory to select from the remote file system:" })
        if (isCancel(filePath)) {
            note(color.red("Selection canceled."))
            return filePath
        }

        if (!filePath) {
            note(color.red("No file path entered."))
            return
        }

        try {
            const fileInfo = await client.getFileInfo(filePath.toString())

            if (fileInfo.isDirectory) {
                return filePath
            }

            note(color.red("Selected path is a file. Please select a directory."))
        } catch (err) {
            if (expectExists) {
                note(color.red(`Failed to get file info for ${filePath.toString()}: ${err instanceof Error ? err.message : String(err)}`))
                continue;
            }

            try {
                // if the call succeeds, it means the parent directory exists, which is good enough for our use case since we will be creating the file during upload
                await client.getFileInfo(path.dirname(filePath.toString()));
                return filePath
            } catch (err) {
                note(color.red(`Failed to get file info for ${filePath.toString()}: ${err instanceof Error ? err.message : String(err)}`))
                continue;
            }
        }
    }
}

export async function selectClient(clients: Map<string, IFileSystemService>) {
    while (true) {
        const opts = Array.from(clients.keys())
        if (opts.length === 0) {
            return null
        }

        const alias = await select({ message: "Enter the alias of the client to use:", options: opts.map((v): { label: string; value: string } => ({ label: v, value: v })) })
        if (isCancel(alias)) {
            note(color.red("Selection canceled."))
            return alias
        }

        if (!alias) {
            note(color.red("No alias entered."))
            return
        }

        const client = clients.get(alias.toString())
        if (!client) {
            note(color.red(`No client found with alias: ${alias.toString()}`))
            continue
        }

        return client
    }
}