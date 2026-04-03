
import { ChannelCredentials } from "@grpc/grpc-js"
import { GreeterClient } from "./proto/HelloService_grpc_pb"
import { HelloRequest } from "./proto/HelloService_pb"

const address = process.env.GRPC_SERVER_ADDRESS || "localhost:8080"
const name = process.env.NAME || "Aditya"


const credentials = ChannelCredentials.createInsecure()
const client = new GreeterClient(address, credentials)
const helloRequest = new HelloRequest()
helloRequest.setName(name)

const shutdown = (exitCode: number) => {
    client.close()
    console.log("Client closed.")
    setTimeout(() => process.exit(exitCode), 0)
}

client.sayHello(helloRequest, (err, response) => {
    if (err) {
        console.error("Error:", err)
        shutdown(1)
        return
    } else {
        console.log("Response:", response.getMessage())
    }

    console.log("Closing client...")
    shutdown(0)

})
