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
	CONNECTION_ERROR  string = "Erro de conexão"
	UUID_REQUIRED     string = "UUID é obrigatório"
	ROLE_REQUIRED     string = "Função é obrigatória"
	PHONE_REQUIRED    string = "Telefone é obrigatório"
	PASSWORD_REQUIRED string = "Senha é obrigatória"
	NAME_REQUIRED     string = "Nome é obrigatório"
	CATEGORY_REQUIRED string = "Categoria é obrigatória"
	ITEM_REQUIRED     string = "Item é obrigatório"
	PRICE_REQUIRED    string = "Preço é obrigatório"

	INVALID_UUID             string = "UUID inválido"
	SKU_REQUIRED                    = "SKU é obrigatório"
	QUANTITY_REQUIRED               = "Quantidade é obrigatória"
	STOCK_CANNOT_BE_NEGATIVE        = "Estoque não pode ser negativo"

	AUTH_ERROR              string = "Erro de autenticação"
	INTERNAL_ERROR          string = "Erro interno"
	INVALID_PASSWORD        string = "Senha inválida"
	INVALID_CREDENTIALS     string = "Usuario ou senha incorretos"
	DISABLED_USER           string = "Usuário desabilitado"
	VERIFICATION_ERROR      string = "Usuário não verificado"
	USER_ALREADY_EXISTS     string = "Usuário já existe"
	STOCK_ALREADY_EXISTS    string = "Estoque já existe"
	WHATSAPP_ALREADY_EXISTS string = "Whatsapp já existe"
	DATABASE_ERROR          string = "Erro de banco de dados"

	USER_NOT_FOUND string = "Usuário não encontrado"
	NO_USERS_FOUND string = "Nenhum usuário encontrado"

	WHATSAPP_NOT_FOUND string = "Whatsapp não encontrado"
	NO_WHATSAPP_FOUND  string = "Nenhum whatsapp encontrado"

	UNAUTHORIZED string = "Não autorizado"

	PERMISSION_DENIED string = "Permissão negada"
	NO_ROWS_FOUND     string = "Não encontrado"
	NO_ROWS_AFFECTED  string = "Não Alterado"
)
