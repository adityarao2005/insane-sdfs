import "reflect-metadata"
import { afterEach, describe, expect, test } from "bun:test"
import { mkdtemp, readFile, rm } from "fs/promises"
import { tmpdir } from "os"
import { join } from "path"
import { createCertificateManager } from "../crypto/cert-manager"

const tempDirs: string[] = []

afterEach(async () => {
    await Promise.all(
        tempDirs.splice(0).map(async (dir) => {
            await rm(dir, { recursive: true, force: true })
        })
    )
})

async function makeTempDir(): Promise<string> {
    const dir = await mkdtemp(join(tmpdir(), "insane-sdfs-cert-manager-"))
    tempDirs.push(dir)
    return dir
}

describe("CertificateManager.createCSRForAlias", () => {
    test("returns CSR in PEM format", async () => {
        const certDir = await makeTempDir()
        const manager = createCertificateManager(certDir)

        const pem = await manager.createCSRForAlias("alpha")

        expect(typeof pem).toBe("string")
        expect(pem).toContain("-----BEGIN CERTIFICATE REQUEST-----")
        expect(pem).toContain("-----END CERTIFICATE REQUEST-----")
    })

    test("creates key files once and reuses them on subsequent calls", async () => {
        const certDir = await makeTempDir()
        const manager = createCertificateManager(certDir)

        await manager.createCSRForAlias("alpha")

        const privateKeyPath = join(certDir, "private.key")
        const publicKeyPath = join(certDir, "public.key")

        const privateKeyFirst = await readFile(privateKeyPath, "utf8")
        const publicKeyFirst = await readFile(publicKeyPath, "utf8")

        await manager.createCSRForAlias("alpha")

        const privateKeySecond = await readFile(privateKeyPath, "utf8")
        const publicKeySecond = await readFile(publicKeyPath, "utf8")

        expect(privateKeyFirst).toBe(privateKeySecond)
        expect(publicKeyFirst).toBe(publicKeySecond)
    })
})
