# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: rainbow.proto
# Protobuf Python Version: 4.25.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder

# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2

DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(
    b'\n\rrainbow.proto\x12\x1e\x63onvergedcomputing.org.grpc.v1\x1a\x1fgoogle/protobuf/timestamp.proto"{\n\x0fRegisterRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x0e\n\x06secret\x18\x02 \x01(\t\x12\r\n\x05nodes\x18\x03 \x01(\t\x12\x11\n\tsubsystem\x18\x04 \x01(\t\x12(\n\x04sent\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.Timestamp"?\n\rDeleteRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x0e\n\x06secret\x18\x02 \x01(\t\x12\x10\n\x08subsytem\x18\x03 \x01(\t"\xb8\x01\n\x0e\x44\x65leteResponse\x12I\n\x06status\x18\x01 \x01(\x0e\x32\x39.convergedcomputing.org.grpc.v1.DeleteResponse.ResultType"[\n\nResultType\x12\x12\n\x0e\x44\x45LETE_SUCCESS\x10\x00\x12\x10\n\x0c\x44\x45LETE_ERROR\x10\x01\x12\x11\n\rDELETE_DENIED\x10\x02\x12\x14\n\x10\x44\x45LETE_NO_EXISTS\x10\x03"F\n\x12UpdateStateRequest\x12\x0f\n\x07\x63luster\x18\x01 \x01(\t\x12\x0e\n\x06secret\x18\x02 \x01(\t\x12\x0f\n\x07payload\x18\x03 \x01(\t"\xdd\x01\n\x13UpdateStateResponse\x12N\n\x06status\x18\x01 \x01(\x0e\x32>.convergedcomputing.org.grpc.v1.UpdateStateResponse.ResultType"v\n\nResultType\x12\x1c\n\x18UPDATE_STATE_UNSPECIFIED\x10\x00\x12\x18\n\x14UPDATE_STATE_PARTIAL\x10\x01\x12\x18\n\x14UPDATE_STATE_SUCCESS\x10\x02\x12\x16\n\x12UPDATE_STATE_ERROR\x10\x03"\x92\x03\n\x10SubmitJobRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\x12J\n\x08\x63lusters\x18\x02 \x03(\x0b\x32\x38.convergedcomputing.org.grpc.v1.SubmitJobRequest.Cluster\x12\x0f\n\x07jobspec\x18\x03 \x01(\t\x12\x18\n\x10select_algorithm\x18\x04 \x01(\t\x12[\n\x0eselect_options\x18\x05 \x03(\x0b\x32\x43.convergedcomputing.org.grpc.v1.SubmitJobRequest.SelectOptionsEntry\x12\x14\n\x0csatisfy_only\x18\x06 \x01(\x08\x12(\n\x04sent\x18\x07 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x1a\x34\n\x12SelectOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\x1a&\n\x07\x43luster\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\r\n\x05token\x18\x02 \x01(\t"p\n\x12ReceiveJobsRequest\x12\x0f\n\x07\x63luster\x18\x01 \x01(\t\x12\x0e\n\x06secret\x18\x02 \x01(\t\x12\x0f\n\x07maxJobs\x18\x03 \x01(\x05\x12(\n\x04sent\x18\x07 \x01(\x0b\x32\x1a.google.protobuf.Timestamp"n\n\x11\x41\x63\x63\x65ptJobsRequest\x12\x0f\n\x07\x63luster\x18\x01 \x01(\t\x12\x0e\n\x06secret\x18\x02 \x01(\t\x12\x0e\n\x06jobids\x18\x03 \x03(\x05\x12(\n\x04sent\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.Timestamp"\x8e\x02\n\x10RegisterResponse\x12\x12\n\nrequest_id\x18\x01 \x01(\t\x12\r\n\x05token\x18\x02 \x01(\t\x12\x0e\n\x06secret\x18\x03 \x01(\t\x12K\n\x06status\x18\x04 \x01(\x0e\x32;.convergedcomputing.org.grpc.v1.RegisterResponse.ResultType"z\n\nResultType\x12\x18\n\x14REGISTER_UNSPECIFIED\x10\x00\x12\x14\n\x10REGISTER_SUCCESS\x10\x01\x12\x12\n\x0eREGISTER_ERROR\x10\x02\x12\x13\n\x0fREGISTER_DENIED\x10\x03\x12\x13\n\x0fREGISTER_EXISTS\x10\x04"\x86\x02\n\x11SubmitJobResponse\x12\x12\n\nrequest_id\x18\x01 \x01(\t\x12\r\n\x05jobid\x18\x02 \x01(\x05\x12\x0f\n\x07\x63luster\x18\x03 \x01(\t\x12L\n\x06status\x18\x04 \x01(\x0e\x32<.convergedcomputing.org.grpc.v1.SubmitJobResponse.ResultType\x12\x10\n\x08\x63lusters\x18\x05 \x03(\t"]\n\nResultType\x12\x16\n\x12SUBMIT_UNSPECIFIED\x10\x00\x12\x12\n\x0eSUBMIT_SUCCESS\x10\x01\x12\x10\n\x0cSUBMIT_ERROR\x10\x02\x12\x11\n\rSUBMIT_DENIED\x10\x03"\xe8\x02\n\x13ReceiveJobsResponse\x12\x12\n\nrequest_id\x18\x01 \x01(\t\x12K\n\x04jobs\x18\x02 \x03(\x0b\x32=.convergedcomputing.org.grpc.v1.ReceiveJobsResponse.JobsEntry\x12N\n\x06status\x18\x03 \x01(\x0e\x32>.convergedcomputing.org.grpc.v1.ReceiveJobsResponse.ResultType\x1a+\n\tJobsEntry\x12\x0b\n\x03key\x18\x01 \x01(\x05\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01"s\n\nResultType\x12\x1a\n\x16REQUEST_JOBS_NORESULTS\x10\x00\x12\x18\n\x14REQUEST_JOBS_SUCCESS\x10\x01\x12\x16\n\x12REQUEST_JOBS_ERROR\x10\x02\x12\x17\n\x13REQUEST_JOBS_DENIED\x10\x03"\xd7\x01\n\x12\x41\x63\x63\x65ptJobsResponse\x12M\n\x06status\x18\x01 \x01(\x0e\x32=.convergedcomputing.org.grpc.v1.AcceptJobsResponse.ResultType"r\n\nResultType\x12\x1b\n\x17RESULT_TYPE_UNSPECIFIED\x10\x00\x12\x17\n\x13RESULT_TYPE_PARTIAL\x10\x01\x12\x17\n\x13RESULT_TYPE_SUCCESS\x10\x02\x12\x15\n\x11RESULT_TYPE_ERROR\x10\x03\x32\xb1\x07\n\x10RainbowScheduler\x12m\n\x08Register\x12/.convergedcomputing.org.grpc.v1.RegisterRequest\x1a\x30.convergedcomputing.org.grpc.v1.RegisterResponse\x12k\n\x06\x44\x65lete\x12/.convergedcomputing.org.grpc.v1.RegisterRequest\x1a\x30.convergedcomputing.org.grpc.v1.RegisterResponse\x12v\n\x11RegisterSubsystem\x12/.convergedcomputing.org.grpc.v1.RegisterRequest\x1a\x30.convergedcomputing.org.grpc.v1.RegisterResponse\x12r\n\x0f\x44\x65leteSubsystem\x12-.convergedcomputing.org.grpc.v1.DeleteRequest\x1a\x30.convergedcomputing.org.grpc.v1.RegisterResponse\x12p\n\tSubmitJob\x12\x30.convergedcomputing.org.grpc.v1.SubmitJobRequest\x1a\x31.convergedcomputing.org.grpc.v1.SubmitJobResponse\x12v\n\x0bUpdateState\x12\x32.convergedcomputing.org.grpc.v1.UpdateStateRequest\x1a\x33.convergedcomputing.org.grpc.v1.UpdateStateResponse\x12v\n\x0bReceiveJobs\x12\x32.convergedcomputing.org.grpc.v1.ReceiveJobsRequest\x1a\x33.convergedcomputing.org.grpc.v1.ReceiveJobsResponse\x12s\n\nAcceptJobs\x12\x31.convergedcomputing.org.grpc.v1.AcceptJobsRequest\x1a\x32.convergedcomputing.org.grpc.v1.AcceptJobsResponseB3Z1github.com/converged-computing/rainbow/pkg/api/v1b\x06proto3'
)

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, "rainbow_pb2", _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
    _globals["DESCRIPTOR"]._options = None
    _globals[
        "DESCRIPTOR"
    ]._serialized_options = b"Z1github.com/converged-computing/rainbow/pkg/api/v1"
    _globals["_SUBMITJOBREQUEST_SELECTOPTIONSENTRY"]._options = None
    _globals["_SUBMITJOBREQUEST_SELECTOPTIONSENTRY"]._serialized_options = b"8\001"
    _globals["_RECEIVEJOBSRESPONSE_JOBSENTRY"]._options = None
    _globals["_RECEIVEJOBSRESPONSE_JOBSENTRY"]._serialized_options = b"8\001"
    _globals["_REGISTERREQUEST"]._serialized_start = 82
    _globals["_REGISTERREQUEST"]._serialized_end = 205
    _globals["_DELETEREQUEST"]._serialized_start = 207
    _globals["_DELETEREQUEST"]._serialized_end = 270
    _globals["_DELETERESPONSE"]._serialized_start = 273
    _globals["_DELETERESPONSE"]._serialized_end = 457
    _globals["_DELETERESPONSE_RESULTTYPE"]._serialized_start = 366
    _globals["_DELETERESPONSE_RESULTTYPE"]._serialized_end = 457
    _globals["_UPDATESTATEREQUEST"]._serialized_start = 459
    _globals["_UPDATESTATEREQUEST"]._serialized_end = 529
    _globals["_UPDATESTATERESPONSE"]._serialized_start = 532
    _globals["_UPDATESTATERESPONSE"]._serialized_end = 753
    _globals["_UPDATESTATERESPONSE_RESULTTYPE"]._serialized_start = 635
    _globals["_UPDATESTATERESPONSE_RESULTTYPE"]._serialized_end = 753
    _globals["_SUBMITJOBREQUEST"]._serialized_start = 756
    _globals["_SUBMITJOBREQUEST"]._serialized_end = 1158
    _globals["_SUBMITJOBREQUEST_SELECTOPTIONSENTRY"]._serialized_start = 1066
    _globals["_SUBMITJOBREQUEST_SELECTOPTIONSENTRY"]._serialized_end = 1118
    _globals["_SUBMITJOBREQUEST_CLUSTER"]._serialized_start = 1120
    _globals["_SUBMITJOBREQUEST_CLUSTER"]._serialized_end = 1158
    _globals["_RECEIVEJOBSREQUEST"]._serialized_start = 1160
    _globals["_RECEIVEJOBSREQUEST"]._serialized_end = 1272
    _globals["_ACCEPTJOBSREQUEST"]._serialized_start = 1274
    _globals["_ACCEPTJOBSREQUEST"]._serialized_end = 1384
    _globals["_REGISTERRESPONSE"]._serialized_start = 1387
    _globals["_REGISTERRESPONSE"]._serialized_end = 1657
    _globals["_REGISTERRESPONSE_RESULTTYPE"]._serialized_start = 1535
    _globals["_REGISTERRESPONSE_RESULTTYPE"]._serialized_end = 1657
    _globals["_SUBMITJOBRESPONSE"]._serialized_start = 1660
    _globals["_SUBMITJOBRESPONSE"]._serialized_end = 1922
    _globals["_SUBMITJOBRESPONSE_RESULTTYPE"]._serialized_start = 1829
    _globals["_SUBMITJOBRESPONSE_RESULTTYPE"]._serialized_end = 1922
    _globals["_RECEIVEJOBSRESPONSE"]._serialized_start = 1925
    _globals["_RECEIVEJOBSRESPONSE"]._serialized_end = 2285
    _globals["_RECEIVEJOBSRESPONSE_JOBSENTRY"]._serialized_start = 2125
    _globals["_RECEIVEJOBSRESPONSE_JOBSENTRY"]._serialized_end = 2168
    _globals["_RECEIVEJOBSRESPONSE_RESULTTYPE"]._serialized_start = 2170
    _globals["_RECEIVEJOBSRESPONSE_RESULTTYPE"]._serialized_end = 2285
    _globals["_ACCEPTJOBSRESPONSE"]._serialized_start = 2288
    _globals["_ACCEPTJOBSRESPONSE"]._serialized_end = 2503
    _globals["_ACCEPTJOBSRESPONSE_RESULTTYPE"]._serialized_start = 2389
    _globals["_ACCEPTJOBSRESPONSE_RESULTTYPE"]._serialized_end = 2503
    _globals["_RAINBOWSCHEDULER"]._serialized_start = 2506
    _globals["_RAINBOWSCHEDULER"]._serialized_end = 3451
# @@protoc_insertion_point(module_scope)
