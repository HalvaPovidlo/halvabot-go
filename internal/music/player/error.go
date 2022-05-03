package player

type ErrorCode int

const (
	NotConnected ErrorCode = iota + 1
	SongNotFound
	StorageQueryFailed
	QueueEmpty
	ConnectFailed
	NotImplemented
)

var (
	ErrNotConnected       = &Error{code: NotConnected, msg: "player not connected"}
	ErrSongNotFound       = &Error{code: SongNotFound, msg: "song not found"}
	ErrStorageQueryFailed = &Error{code: StorageQueryFailed, msg: "storage query failed"}
	ErrQueueEmpty         = &Error{code: QueueEmpty, msg: "queue is empty. nothing to play"}
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
