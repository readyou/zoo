package util

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestEncryptUtil_EncryptPassword(t *testing.T) {
	password := UUID.NewString()
	hash1 := Encrypt.EncryptPassword(password)
	hash2 := Encrypt.EncryptPassword(password)
	assert.Equal(t, hash1, hash2)
}

func TestEncryptUtil_RandStr(t *testing.T) {
	n := 10000000
	m := make(map[string]bool)
	t1 := time.Now()
	for i := 0; i < n; i++ {
		str := Encrypt.RandStr()
		m[str] = true
	}
	t2 := time.Now()
	cost := t2.Sub(t1).Seconds()
	log.Printf("size: %d, time: %fs, qps: %f", len(m), cost, float64(n)/cost)
}
