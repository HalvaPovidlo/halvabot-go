package player

type ErrorCode int

const (
	NotConnected ErrorCode = iota + 1
	SongNotFound
	StorageQueryFailed
	ConnectFailed
	NotImplemented
)

var (
	ErrNotConnected       = &Error{code: NotConnected, msg: "player not connected"}
	ErrSongNotFound       = &Error{code: SongNotFound, msg: "song not found"}
	ErrStorageQueryFailed = &Error{code: StorageQueryFailed, msg: "storage query failed"}
	ErrConnectFailed      = &Error{code: ConnectFailed, msg: "connection failed"}
	ErrNotImplemented     = &Error{code: NotImplemented, msg: "method not implemented"}
)

type Error struct {
	code ErrorCode
	msg  string
}

func (e *Error) Error() string {
	return e.msg
}

func (e *Error) Wrap(msg string) *Error {
	e.msg = msg
	return e
}
