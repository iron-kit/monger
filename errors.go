package monger

type MongerQueryError struct {
	message string
}

func (err *MongerQueryError) Error() string {
	return err.message
}

type NotFoundError struct {
	*MongerQueryError
}

type DuplicateDocumentError struct {
	*MongerQueryError
}

type ValidationError struct {
	*MongerQueryError
	Errors []error
}

type InvalidIdError struct {
	*MongerQueryError
}

type NotInitDocumentError struct {
	*MongerQueryError
}

type InvalidParamsError struct {
	*MongerQueryError
}

func NewError(msg string) *MongerQueryError {
	return &MongerQueryError{msg}
}
