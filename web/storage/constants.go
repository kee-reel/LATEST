package storage

type TokenData struct {
	IP     string             `json:"ip"`
	Email  string             `json:"email"`
	Extra  *map[string]string `json:"extra"`
	UserId int                `json:"-"`
}

type TokenType int

const (
	RegisterToken TokenType = 1
	VerifyToken             = 2
	AccessToken             = 3
	RestoreToken            = 4
	SuspendToken            = 5
)

type TokenError int

const (
	NoError      TokenError = 0
	TokenUnknown            = 1
	TokenExists             = 2
	EmailTaken              = 3
	EmailUnknown            = 4
	WrongIP                 = 5
)
