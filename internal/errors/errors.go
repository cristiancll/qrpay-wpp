package errors

type Error struct {
	Code       int64
	Type       string
	Message    string
	Stacktrace []string
}

func (e *Error) Error() string {
	return ""
}

const (
	UUID_REQUIRED    string = "UUID é obrigatório"
	DATABASE_ERROR   string = "Erro de banco de dados"
	NO_ROWS_FOUND    string = "Não encontrado"
	NO_ROWS_AFFECTED string = "Não Alterado"
)
