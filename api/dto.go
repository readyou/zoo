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

type LoginReq struct {
	Username string
	Password string
}

type LogoutReq struct {
	Token string
}

type UpdateProfileReq struct {
	Token    string
	Nickname string
	Avatar   string
}

type UploadImgReq struct {
	Token    string
	ByteList []byte
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