import { describe, expect, test } from "bun:test"
import { streamToGenerator } from "../cli/stream-utils"

describe("streamToGenerator", () => {
    test("yields chunks in order", async () => {
        const chunks = [new Uint8Array([1]), new Uint8Array([2, 3])]
        const stream = new ReadableStream<Uint8Array>({
            start(controller) {
                for (const chunk of chunks) {
                    controller.enqueue(chunk)
                }
                controller.close()
            }
        })

        const collected: Uint8Array[] = []
        for await (const chunk of streamToGenerator(stream)) {
            collected.push(chunk)
        }

        expect(collected).toEqual(chunks)
    })

    test("releases the reader lock", async () => {
        const stream = new ReadableStream<Uint8Array>({
            start(controller) {
                controller.enqueue(new Uint8Array([7]))
                controller.close()
            }
        })

        for await (const _ of streamToGenerator(stream)) {
            // consume stream fully
        }

        const reader = stream.getReader()
        reader.releaseLock()
        expect(true).toBe(true)
    })
})