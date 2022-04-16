package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

var Encrypt = &encryptUtil{}

type encryptUtil struct {
}

func (*encryptUtil) RandStr() string {
	byteList := make([]byte, 32)
	_, err := rand.Read(byteList)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(byteList)
}

func (*encryptUtil) EncryptPassword(password string) string {
	salt := "@#k8L!%Z"
	password = password + salt
	h := md5.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}
