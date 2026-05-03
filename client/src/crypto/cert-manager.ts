import { randomUUID } from "crypto"
import { mkdir, readFile, writeFile } from "fs/promises"
import { ChannelCredentials, credentials } from "@grpc/grpc-js"
import { Pkcs10CertificateRequestGenerator } from "@peculiar/x509";

export interface ICertificateManager {
    getClientIdentityCertificate(alias: string): Promise<ChannelCredentials>;

    createCSRForAlias(alias: string): Promise<string>;
}

const crypto = globalThis.crypto

/**
 * The `CertificateManager` class is responsible for managing TLS certificates for secure communication between the client and server. It provides methods to retrieve client identity certificates, create certificate signing requests (CSRs), and validate and add certificates to the client's trust store.
 */
class CertificateManager implements ICertificateManager {

    private certificateDirectory: string;

    constructor(certificateDirectory: string) {
        this.certificateDirectory = certificateDirectory;
    }

    /**
     * Retrieves the client identity certificate for the specified alias.
     * The certificate data is expected to be stored in the following structure:
     * 
     * - server-cas.pem
     * - identities/
     *   - {alias}-cert.pem
     * Server CAs are stored in the `server-cas.pem` file, while client certificates and their corresponding private keys are stored in the `identities` directory, organized by alias.
     * 
     * @param alias the alias of the client being connected to
     * @returns the client TLS config
     */
    async getClientIdentityCertificate(alias: string) {
        // declare the files
        const serverCAs = `${this.certificateDirectory}/server-cas.pem`;
        const certPath = `${this.certificateDirectory}/identities/${alias}-cert.pem`;

        const { privateKey } = await this.getCryptoKeyPair();
        const privateKeyPem = await crypto.subtle.exportKey("pkcs8", privateKey)

        // read the files
        try {
            const certFile = await readFile(certPath);
            const serverCAPool = await readFile(serverCAs);

            // create the TLS credentials
            return credentials.createSsl(serverCAPool, Buffer.from(privateKeyPem), certFile);
        } catch (error: unknown) {
            throw new Error(`Failed to read certificate files for alias ${alias}: ${(error as Error).message}`);
        }
    }

    /**
     * Retrieves the crypto key pair for the client. If the key pair does not exist, it generates a new one and stores it in the certificate directory. The private key is stored in `private.key` and the public key is stored in `public.key`. Both keys are stored in JSON Web Key (JWK) format.
     * @returns the crypto key pair for the client
     */
    private async getCryptoKeyPair(): Promise<CryptoKeyPair> {
        await mkdir(`${this.certificateDirectory}`, { recursive: true });

        const privateKeyPath = `${this.certificateDirectory}/private.key`;
        const publicKeyPath = `${this.certificateDirectory}/public.key`;

        const algorithm = {
            name: "Ed25519",
            namedCurve: "Ed25519"
        };

        try {
            const [privateKeyContent, publicKeyContent] = await Promise.all([
                readFile(privateKeyPath, "utf8"),
                readFile(publicKeyPath, "utf8")
            ]);

            const privateJwk = JSON.parse(privateKeyContent) as JsonWebKey;
            const publicJwk = JSON.parse(publicKeyContent) as JsonWebKey;

            const [privateKey, publicKey] = await Promise.all([
                crypto.subtle.importKey("jwk", privateJwk, algorithm, true, ["sign"]),
                crypto.subtle.importKey("jwk", publicJwk, algorithm, true, ["verify"])
            ]);

            return { privateKey, publicKey };
        } catch (error: unknown) {
            const readError = error as NodeJS.ErrnoException;
            const missingKeyFiles = readError?.code === "ENOENT";
            if (!missingKeyFiles) {
                throw error;
            }
        }

        // create a public and private key for the certificates
        const { privateKey, publicKey } = await crypto.subtle.generateKey(
            algorithm, true,
            ["sign", "verify"]
        );

        const [privateJwk, publicJwk] = await Promise.all([
            crypto.subtle.exportKey("jwk", privateKey),
            crypto.subtle.exportKey("jwk", publicKey)
        ]);

        await Promise.all([
            writeFile(privateKeyPath, JSON.stringify(privateJwk), "utf8"),
            writeFile(publicKeyPath, JSON.stringify(publicJwk), "utf8")
        ]);

        return { privateKey, publicKey };
    }

    /**
     * This will create a public and private key pair and store the the PEM encoded private key in the `identities` folder of the directory for the server alias.
     * 
     * Given both of these we shall create a certificate signing request (CSR) and send it to the server and return the PEM encoded CSR.
     * 
     * @param alias the alias of the client for which to create a CSR
     */
    async createCSRForAlias(alias: string) {
        // create a public and private key for the certificates
        const algorithm =
        {
            name: "Ed25519",
            namedCurve: "Ed25519"
        };

        const { privateKey, publicKey } = await this.getCryptoKeyPair();

        const csrId = randomUUID()
        const certReq = await Pkcs10CertificateRequestGenerator.create({
            name: "CN=" + csrId + ", O=org " + alias,
            keys: { privateKey, publicKey },
            signingAlgorithm: algorithm
        }, crypto)

        const pemEncodedString = certReq.toString("pem");

        // export csr in PEM format to identities/{alias}-csr.pem
        await writeFile(`${this.certificateDirectory}/identities/${alias}-csr.pem`, pemEncodedString, "utf8");

        return pemEncodedString
    }

}

export function createCertificateManager(certificateDirectory: string): ICertificateManager {
    return new CertificateManager(certificateDirectory);
}