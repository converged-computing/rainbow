from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RegisterRequest(_message.Message):
    __slots__ = ("name", "secret", "nodes", "subsystem", "sent")
    NAME_FIELD_NUMBER: _ClassVar[int]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    NODES_FIELD_NUMBER: _ClassVar[int]
    SUBSYSTEM_FIELD_NUMBER: _ClassVar[int]
    SENT_FIELD_NUMBER: _ClassVar[int]
    name: str
    secret: str
    nodes: str
    subsystem: str
    sent: _timestamp_pb2.Timestamp
    def __init__(self, name: _Optional[str] = ..., secret: _Optional[str] = ..., nodes: _Optional[str] = ..., subsystem: _Optional[str] = ..., sent: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class SubmitJobRequest(_message.Message):
    __slots__ = ("name", "clusters", "jobspec", "sent")
    class Cluster(_message.Message):
        __slots__ = ("name", "token")
        NAME_FIELD_NUMBER: _ClassVar[int]
        TOKEN_FIELD_NUMBER: _ClassVar[int]
        name: str
        token: str
        def __init__(self, name: _Optional[str] = ..., token: _Optional[str] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    CLUSTERS_FIELD_NUMBER: _ClassVar[int]
    JOBSPEC_FIELD_NUMBER: _ClassVar[int]
    SENT_FIELD_NUMBER: _ClassVar[int]
    name: str
    clusters: _containers.RepeatedCompositeFieldContainer[SubmitJobRequest.Cluster]
    jobspec: str
    sent: _timestamp_pb2.Timestamp
    def __init__(self, name: _Optional[str] = ..., clusters: _Optional[_Iterable[_Union[SubmitJobRequest.Cluster, _Mapping]]] = ..., jobspec: _Optional[str] = ..., sent: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class ReceiveJobsRequest(_message.Message):
    __slots__ = ("cluster", "secret", "maxJobs", "sent")
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    MAXJOBS_FIELD_NUMBER: _ClassVar[int]
    SENT_FIELD_NUMBER: _ClassVar[int]
    cluster: str
    secret: str
    maxJobs: int
    sent: _timestamp_pb2.Timestamp
    def __init__(self, cluster: _Optional[str] = ..., secret: _Optional[str] = ..., maxJobs: _Optional[int] = ..., sent: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class AcceptJobsRequest(_message.Message):
    __slots__ = ("cluster", "secret", "jobids", "sent")
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    JOBIDS_FIELD_NUMBER: _ClassVar[int]
    SENT_FIELD_NUMBER: _ClassVar[int]
    cluster: str
    secret: str
    jobids: _containers.RepeatedScalarFieldContainer[int]
    sent: _timestamp_pb2.Timestamp
    def __init__(self, cluster: _Optional[str] = ..., secret: _Optional[str] = ..., jobids: _Optional[_Iterable[int]] = ..., sent: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class RegisterResponse(_message.Message):
    __slots__ = ("request_id", "token", "secret", "status")
    class ResultType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        REGISTER_UNSPECIFIED: _ClassVar[RegisterResponse.ResultType]
        REGISTER_SUCCESS: _ClassVar[RegisterResponse.ResultType]
        REGISTER_ERROR: _ClassVar[RegisterResponse.ResultType]
        REGISTER_DENIED: _ClassVar[RegisterResponse.ResultType]
        REGISTER_EXISTS: _ClassVar[RegisterResponse.ResultType]
    REGISTER_UNSPECIFIED: RegisterResponse.ResultType
    REGISTER_SUCCESS: RegisterResponse.ResultType
    REGISTER_ERROR: RegisterResponse.ResultType
    REGISTER_DENIED: RegisterResponse.ResultType
    REGISTER_EXISTS: RegisterResponse.ResultType
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    token: str
    secret: str
    status: RegisterResponse.ResultType
    def __init__(self, request_id: _Optional[str] = ..., token: _Optional[str] = ..., secret: _Optional[str] = ..., status: _Optional[_Union[RegisterResponse.ResultType, str]] = ...) -> None: ...

class SubmitJobResponse(_message.Message):
    __slots__ = ("request_id", "jobid", "status")
    class ResultType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        SUBMIT_UNSPECIFIED: _ClassVar[SubmitJobResponse.ResultType]
        SUBMIT_SUCCESS: _ClassVar[SubmitJobResponse.ResultType]
        SUBMIT_ERROR: _ClassVar[SubmitJobResponse.ResultType]
        SUBMIT_DENIED: _ClassVar[SubmitJobResponse.ResultType]
    SUBMIT_UNSPECIFIED: SubmitJobResponse.ResultType
    SUBMIT_SUCCESS: SubmitJobResponse.ResultType
    SUBMIT_ERROR: SubmitJobResponse.ResultType
    SUBMIT_DENIED: SubmitJobResponse.ResultType
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    JOBID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    jobid: int
    status: SubmitJobResponse.ResultType
    def __init__(self, request_id: _Optional[str] = ..., jobid: _Optional[int] = ..., status: _Optional[_Union[SubmitJobResponse.ResultType, str]] = ...) -> None: ...

class ReceiveJobsResponse(_message.Message):
    __slots__ = ("request_id", "jobs", "status")
    class ResultType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        REQUEST_JOBS_NORESULTS: _ClassVar[ReceiveJobsResponse.ResultType]
        REQUEST_JOBS_SUCCESS: _ClassVar[ReceiveJobsResponse.ResultType]
        REQUEST_JOBS_ERROR: _ClassVar[ReceiveJobsResponse.ResultType]
        REQUEST_JOBS_DENIED: _ClassVar[ReceiveJobsResponse.ResultType]
    REQUEST_JOBS_NORESULTS: ReceiveJobsResponse.ResultType
    REQUEST_JOBS_SUCCESS: ReceiveJobsResponse.ResultType
    REQUEST_JOBS_ERROR: ReceiveJobsResponse.ResultType
    REQUEST_JOBS_DENIED: ReceiveJobsResponse.ResultType
    class JobsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: int
        value: str
        def __init__(self, key: _Optional[int] = ..., value: _Optional[str] = ...) -> None: ...
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    JOBS_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    jobs: _containers.ScalarMap[int, str]
    status: ReceiveJobsResponse.ResultType
    def __init__(self, request_id: _Optional[str] = ..., jobs: _Optional[_Mapping[int, str]] = ..., status: _Optional[_Union[ReceiveJobsResponse.ResultType, str]] = ...) -> None: ...

class AcceptJobsResponse(_message.Message):
    __slots__ = ("status",)
    class ResultType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        RESULT_TYPE_UNSPECIFIED: _ClassVar[AcceptJobsResponse.ResultType]
        RESULT_TYPE_PARTIAL: _ClassVar[AcceptJobsResponse.ResultType]
        RESULT_TYPE_SUCCESS: _ClassVar[AcceptJobsResponse.ResultType]
        RESULT_TYPE_ERROR: _ClassVar[AcceptJobsResponse.ResultType]
    RESULT_TYPE_UNSPECIFIED: AcceptJobsResponse.ResultType
    RESULT_TYPE_PARTIAL: AcceptJobsResponse.ResultType
    RESULT_TYPE_SUCCESS: AcceptJobsResponse.ResultType
    RESULT_TYPE_ERROR: AcceptJobsResponse.ResultType
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: AcceptJobsResponse.ResultType
    def __init__(self, status: _Optional[_Union[AcceptJobsResponse.ResultType, str]] = ...) -> None: ...
