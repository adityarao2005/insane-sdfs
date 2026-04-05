// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var FileSystemService_pb = require('./FileSystemService_pb.js');

function serialize_pb_DownloadFileRequest(arg) {
  if (!(arg instanceof FileSystemService_pb.DownloadFileRequest)) {
    throw new Error('Expected argument of type pb.DownloadFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_DownloadFileRequest(buffer_arg) {
  return FileSystemService_pb.DownloadFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pb_DownloadFileResponseFragment(arg) {
  if (!(arg instanceof FileSystemService_pb.DownloadFileResponseFragment)) {
    throw new Error('Expected argument of type pb.DownloadFileResponseFragment');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_DownloadFileResponseFragment(buffer_arg) {
  return FileSystemService_pb.DownloadFileResponseFragment.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pb_FileInfo(arg) {
  if (!(arg instanceof FileSystemService_pb.FileInfo)) {
    throw new Error('Expected argument of type pb.FileInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_FileInfo(buffer_arg) {
  return FileSystemService_pb.FileInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pb_FileInfoRequest(arg) {
  if (!(arg instanceof FileSystemService_pb.FileInfoRequest)) {
    throw new Error('Expected argument of type pb.FileInfoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_FileInfoRequest(buffer_arg) {
  return FileSystemService_pb.FileInfoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pb_PingRequest(arg) {
  if (!(arg instanceof FileSystemService_pb.PingRequest)) {
    throw new Error('Expected argument of type pb.PingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_PingRequest(buffer_arg) {
  return FileSystemService_pb.PingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pb_PongResponse(arg) {
  if (!(arg instanceof FileSystemService_pb.PongResponse)) {
    throw new Error('Expected argument of type pb.PongResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_PongResponse(buffer_arg) {
  return FileSystemService_pb.PongResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pb_SuccessResponse(arg) {
  if (!(arg instanceof FileSystemService_pb.SuccessResponse)) {
    throw new Error('Expected argument of type pb.SuccessResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_SuccessResponse(buffer_arg) {
  return FileSystemService_pb.SuccessResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pb_UploadFileRequestFragment(arg) {
  if (!(arg instanceof FileSystemService_pb.UploadFileRequestFragment)) {
    throw new Error('Expected argument of type pb.UploadFileRequestFragment');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pb_UploadFileRequestFragment(buffer_arg) {
  return FileSystemService_pb.UploadFileRequestFragment.deserializeBinary(new Uint8Array(buffer_arg));
}


// *
// The file system service definition.
var FileSystemService = exports.FileSystemService = {
  // *
// @param req The request message containing the file path and data.
// @return The response message indicating success or failure.
uploadFile: {
    path: '/pb.FileSystem/UploadFile',
    requestStream: true,
    responseStream: false,
    requestType: FileSystemService_pb.UploadFileRequestFragment,
    responseType: FileSystemService_pb.SuccessResponse,
    requestSerialize: serialize_pb_UploadFileRequestFragment,
    requestDeserialize: deserialize_pb_UploadFileRequestFragment,
    responseSerialize: serialize_pb_SuccessResponse,
    responseDeserialize: deserialize_pb_SuccessResponse,
  },
  // *
// @param req The request message containing the file path.
// @return The response message containing the file data.
downloadFile: {
    path: '/pb.FileSystem/DownloadFile',
    requestStream: false,
    responseStream: true,
    requestType: FileSystemService_pb.DownloadFileRequest,
    responseType: FileSystemService_pb.DownloadFileResponseFragment,
    requestSerialize: serialize_pb_DownloadFileRequest,
    requestDeserialize: deserialize_pb_DownloadFileRequest,
    responseSerialize: serialize_pb_DownloadFileResponseFragment,
    responseDeserialize: deserialize_pb_DownloadFileResponseFragment,
  },
  // *
// @param req The request message containing the file path.
// @return The response message indicating success or failure.
deleteFile: {
    path: '/pb.FileSystem/DeleteFile',
    requestStream: false,
    responseStream: false,
    requestType: FileSystemService_pb.FileInfoRequest,
    responseType: FileSystemService_pb.SuccessResponse,
    requestSerialize: serialize_pb_FileInfoRequest,
    requestDeserialize: deserialize_pb_FileInfoRequest,
    responseSerialize: serialize_pb_SuccessResponse,
    responseDeserialize: deserialize_pb_SuccessResponse,
  },
  // *
// @param req The request message containing the file path.
// @return The response message containing the file info.
getFileInfo: {
    path: '/pb.FileSystem/GetFileInfo',
    requestStream: false,
    responseStream: false,
    requestType: FileSystemService_pb.FileInfoRequest,
    responseType: FileSystemService_pb.FileInfo,
    requestSerialize: serialize_pb_FileInfoRequest,
    requestDeserialize: deserialize_pb_FileInfoRequest,
    responseSerialize: serialize_pb_FileInfo,
    responseDeserialize: deserialize_pb_FileInfo,
  },
  // *
// @param req The request message containing the file path.
// @return The response messages containing the file info.
listFiles: {
    path: '/pb.FileSystem/ListFiles',
    requestStream: false,
    responseStream: true,
    requestType: FileSystemService_pb.FileInfoRequest,
    responseType: FileSystemService_pb.FileInfo,
    requestSerialize: serialize_pb_FileInfoRequest,
    requestDeserialize: deserialize_pb_FileInfoRequest,
    responseSerialize: serialize_pb_FileInfo,
    responseDeserialize: deserialize_pb_FileInfo,
  },
  // *
// @param req The request message containing the directory path.
// @return The response message indicating success or failure.
createDirectory: {
    path: '/pb.FileSystem/CreateDirectory',
    requestStream: false,
    responseStream: false,
    requestType: FileSystemService_pb.FileInfoRequest,
    responseType: FileSystemService_pb.SuccessResponse,
    requestSerialize: serialize_pb_FileInfoRequest,
    requestDeserialize: deserialize_pb_FileInfoRequest,
    responseSerialize: serialize_pb_SuccessResponse,
    responseDeserialize: deserialize_pb_SuccessResponse,
  },
  // *
// @param req The request message containing the ping message.
// @return The response message containing the pong message.
ping: {
    path: '/pb.FileSystem/Ping',
    requestStream: false,
    responseStream: false,
    requestType: FileSystemService_pb.PingRequest,
    responseType: FileSystemService_pb.PongResponse,
    requestSerialize: serialize_pb_PingRequest,
    requestDeserialize: deserialize_pb_PingRequest,
    responseSerialize: serialize_pb_PongResponse,
    responseDeserialize: deserialize_pb_PongResponse,
  },
};

exports.FileSystemClient = grpc.makeGenericClientConstructor(FileSystemService, 'FileSystem');
