
export async function* streamToGenerator(readableStream: ReadableStream<Uint8Array>): AsyncGenerator<Uint8Array> {
    const reader = readableStream.getReader()
    try {
        while (true) {
            const { done, value } = await reader.read()
            if (done) {
                break
            }
            yield value
        }
    } finally {
        reader.releaseLock()
    }
}