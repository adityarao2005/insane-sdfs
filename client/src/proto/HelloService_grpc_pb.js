// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var HelloService_pb = require('./HelloService_pb.js');

function serialize_pb_HelloRequest(arg) {
  if (!(arg instanceof HelloService_pb.HelloRequest)) {
    throw new Error('Expected argument of type pb.HelloRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_HelloRequest(buffer_arg) {
  return HelloService_pb.HelloRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pb_HelloResponse(arg) {
  if (!(arg instanceof HelloService_pb.HelloResponse)) {
    throw new Error('Expected argument of type pb.HelloResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_HelloResponse(buffer_arg) {
  return HelloService_pb.HelloResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// *
// The greeting service definition.
var GreeterService = exports.GreeterService = {
  // *
// @param req The request message containing the name.
// @return The response message containing the greeting.
sayHello: {
    path: '/pb.Greeter/SayHello',
    requestStream: false,
    responseStream: false,
    requestType: HelloService_pb.HelloRequest,
    responseType: HelloService_pb.HelloResponse,
    requestSerialize: serialize_pb_HelloRequest,
    requestDeserialize: deserialize_pb_HelloRequest,
    responseSerialize: serialize_pb_HelloResponse,
    responseDeserialize: deserialize_pb_HelloResponse,
  },
};

exports.GreeterClient = grpc.makeGenericClientConstructor(GreeterService, 'Greeter');
