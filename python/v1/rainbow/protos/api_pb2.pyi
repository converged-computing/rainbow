from google.protobuf import any_pb2 as _any_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Content(_message.Message):
    __slots__ = ("id", "data", "metadata")
    class MetadataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    id: str
    data: _any_pb2.Any
    metadata: _containers.ScalarMap[str, str]
    def __init__(self, id: _Optional[str] = ..., data: _Optional[_Union[_any_pb2.Any, _Mapping]] = ..., metadata: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Request(_message.Message):
    __slots__ = ("content", "sent")
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    SENT_FIELD_NUMBER: _ClassVar[int]
    content: Content
    sent: _timestamp_pb2.Timestamp
    def __init__(self, content: _Optional[_Union[Content, _Mapping]] = ..., sent: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class RegisterRequest(_message.Message):
    __slots__ = ("name", "secret", "sent")
    NAME_FIELD_NUMBER: _ClassVar[int]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    SENT_FIELD_NUMBER: _ClassVar[int]
    name: str
    secret: str
    sent: _timestamp_pb2.Timestamp
    def __init__(self, name: _Optional[str] = ..., secret: _Optional[str] = ..., sent: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class SubmitJobRequest(_message.Message):
    __slots__ = ("name", "cluster", "token", "nodes", "tasks", "command", "sent")
    NAME_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    NODES_FIELD_NUMBER: _ClassVar[int]
    TASKS_FIELD_NUMBER: _ClassVar[int]
    COMMAND_FIELD_NUMBER: _ClassVar[int]
    SENT_FIELD_NUMBER: _ClassVar[int]
    name: str
    cluster: str
    token: str
    nodes: int
    tasks: int
    command: str
    sent: _timestamp_pb2.Timestamp
    def __init__(self, name: _Optional[str] = ..., cluster: _Optional[str] = ..., token: _Optional[str] = ..., nodes: _Optional[int] = ..., tasks: _Optional[int] = ..., command: _Optional[str] = ..., sent: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ("request_id", "message_count", "messages_processed", "processing_details")
    class ResultType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        RESULT_TYPE_UNSPECIFIED: _ClassVar[Response.ResultType]
        RESULT_TYPE_SUCCESS: _ClassVar[Response.ResultType]
        RESULT_TYPE_ERROR: _ClassVar[Response.ResultType]
    RESULT_TYPE_UNSPECIFIED: Response.ResultType
    RESULT_TYPE_SUCCESS: Response.ResultType
    RESULT_TYPE_ERROR: Response.ResultType
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_COUNT_FIELD_NUMBER: _ClassVar[int]
    MESSAGES_PROCESSED_FIELD_NUMBER: _ClassVar[int]
    PROCESSING_DETAILS_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    message_count: int
    messages_processed: int
    processing_details: str
    def __init__(self, request_id: _Optional[str] = ..., message_count: _Optional[int] = ..., messages_processed: _Optional[int] = ..., processing_details: _Optional[str] = ...) -> None: ...

class RegisterResponse(_message.Message):
    __slots__ = ("request_id", "token", "status")
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
    STATUS_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    token: str
    status: RegisterResponse.ResultType
    def __init__(self, request_id: _Optional[str] = ..., token: _Optional[str] = ..., status: _Optional[_Union[RegisterResponse.ResultType, str]] = ...) -> None: ...

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
