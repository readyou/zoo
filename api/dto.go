package api

// requests

type RegisterReq struct {
	Username string
	Password string
}

type RefreshTokenReq struct {
	Token        string
	RefreshToken string
}

type EchoTokenReq struct {
	Token string
	Sleep int64
	UseDB bool
}

type LoginReq struct {
	Username string
	Password string
}

type ValidateTokenReq struct {
	Token string
}

type LogoutReq struct {
	Token string
}

type UpdateProfileReq struct {
	Token    string
	Nickname string
	Avatar   string
}

type GetProfileReq struct {
	Token string
}

// responses

type ProfileResp struct {
	Id       int64
	Username string
	Nickname string
	Avatar   string
}

type TokenResp struct {
	Username          string
	Token             string
	ExpireTime        int64
	RefreshToken      string
	RefreshExpireTime int64
}
