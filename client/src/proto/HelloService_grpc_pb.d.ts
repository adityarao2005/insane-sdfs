// package: pb
// file: HelloService.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as HelloService_pb from "./HelloService_pb";

interface IGreeterService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    sayHello: IGreeterService_ISayHello;
}

interface IGreeterService_ISayHello extends grpc.MethodDefinition<HelloService_pb.HelloRequest, HelloService_pb.HelloResponse> {
    path: "/pb.Greeter/SayHello";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<HelloService_pb.HelloRequest>;
    requestDeserialize: grpc.deserialize<HelloService_pb.HelloRequest>;
    responseSerialize: grpc.serialize<HelloService_pb.HelloResponse>;
    responseDeserialize: grpc.deserialize<HelloService_pb.HelloResponse>;
}

export const GreeterService: IGreeterService;

export interface IGreeterServer extends grpc.UntypedServiceImplementation {
    sayHello: grpc.handleUnaryCall<HelloService_pb.HelloRequest, HelloService_pb.HelloResponse>;
}

export interface IGreeterClient {
    sayHello(request: HelloService_pb.HelloRequest, callback: (error: grpc.ServiceError | null, response: HelloService_pb.HelloResponse) => void): grpc.ClientUnaryCall;
    sayHello(request: HelloService_pb.HelloRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: HelloService_pb.HelloResponse) => void): grpc.ClientUnaryCall;
    sayHello(request: HelloService_pb.HelloRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: HelloService_pb.HelloResponse) => void): grpc.ClientUnaryCall;
}

export class GreeterClient extends grpc.Client implements IGreeterClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public sayHello(request: HelloService_pb.HelloRequest, callback: (error: grpc.ServiceError | null, response: HelloService_pb.HelloResponse) => void): grpc.ClientUnaryCall;
    public sayHello(request: HelloService_pb.HelloRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: HelloService_pb.HelloResponse) => void): grpc.ClientUnaryCall;
    public sayHello(request: HelloService_pb.HelloRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: HelloService_pb.HelloResponse) => void): grpc.ClientUnaryCall;
}
