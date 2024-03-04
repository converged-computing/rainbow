from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RegisterRequest(_message.Message):
    __slots__ = ("name", "payload", "subsystem")
    NAME_FIELD_NUMBER: _ClassVar[int]
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    SUBSYSTEM_FIELD_NUMBER: _ClassVar[int]
    name: str
    payload: str
    subsystem: str
    def __init__(self, name: _Optional[str] = ..., payload: _Optional[str] = ..., subsystem: _Optional[str] = ...) -> None: ...

class SatisfyRequest(_message.Message):
    __slots__ = ("payload",)
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    payload: str
    def __init__(self, payload: _Optional[str] = ...) -> None: ...

class SatisfyResponse(_message.Message):
    __slots__ = ("clusters", "status")
    class ResultType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        RESULT_TYPE_UNSPECIFIED: _ClassVar[SatisfyResponse.ResultType]
        RESULT_TYPE_SUCCESS: _ClassVar[SatisfyResponse.ResultType]
        RESULT_TYPE_ERROR: _ClassVar[SatisfyResponse.ResultType]
    RESULT_TYPE_UNSPECIFIED: SatisfyResponse.ResultType
    RESULT_TYPE_SUCCESS: SatisfyResponse.ResultType
    RESULT_TYPE_ERROR: SatisfyResponse.ResultType
    CLUSTERS_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    clusters: _containers.RepeatedScalarFieldContainer[str]
    status: SatisfyResponse.ResultType
    def __init__(self, clusters: _Optional[_Iterable[str]] = ..., status: _Optional[_Union[SatisfyResponse.ResultType, str]] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ("status",)
    class ResultType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        RESULT_TYPE_UNSPECIFIED: _ClassVar[Response.ResultType]
        RESULT_TYPE_SUCCESS: _ClassVar[Response.ResultType]
        RESULT_TYPE_ERROR: _ClassVar[Response.ResultType]
    RESULT_TYPE_UNSPECIFIED: Response.ResultType
    RESULT_TYPE_SUCCESS: Response.ResultType
    RESULT_TYPE_ERROR: Response.ResultType
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: Response.ResultType
    def __init__(self, status: _Optional[_Union[Response.ResultType, str]] = ...) -> None: ...
