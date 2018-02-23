package util

type MessageError interface {
	Error() string
	ErrorCode() string
}

func NewMessageError(message string, code string) *messageError {
	return &messageError{
		message: message,
		code:    code,
	}
}

type messageError struct {
	message string
	code    string
}

func (e *messageError) Error() string {
	return e.message
}

func (e *messageError) ErrorCode() string {
	return e.code
}
