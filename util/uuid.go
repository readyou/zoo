package util

import (
	"github.com/google/uuid"
	"strings"
)

var UUID *uid = &uid{}

type uid struct {
}

func (*uid) NewString() string {
	str := uuid.NewString()
	str = strings.ReplaceAll(str, "-", "")
	return str
}
