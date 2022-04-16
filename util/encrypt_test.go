package util

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"log"
	"sync"
	"testing"
	"time"
)

// 性能太差，不考虑使用
func encryptPassword(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hashed)
}

// 性能太差，不考虑使用
func isPasswordMatch(hashedPassword, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return false
	}
	return true
}

func TestEncryptUtil_EncryptPassword_old(t *testing.T) {
	password := UUID.NewString()
	GetQPS(func(n int) {
		for i := 0; i < n; i++ {
			hash := encryptPassword(password)
			//log.Println(hash)
			assert.True(t, isPasswordMatch(hash, password))
		}
	}, 10000)
	// Output:
	// Qps: 420
}

func TestEncryptUtil_EncryptPassword_benchmark_go(t *testing.T) {
	password := UUID.NewString()
	encryptedPassword := Encrypt.EncryptPassword(password)
	log.Printf("%s: %s\n", password, encryptedPassword)
	GetQPS(func(n int) {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				assert.Equal(t, encryptedPassword, Encrypt.EncryptPassword(password))
				wg.Done()
			}()
		}
		wg.Wait()
	}, 10000000)
	// 性能很好
	// Output:
	// Qps: 970123.854415
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
