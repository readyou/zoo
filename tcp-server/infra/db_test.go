package infra

import (
	"git.garena.com/xinlong.wu/zoo/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDB(t *testing.T) {
	InitDB()
	util.GetQPS(func(n int) {
		for i := 0; i < n; i++ {
			m, err := DB.Exec("select now() as n")
			//log.Printf("%#v, %s\n", m, err)
			assert.Nil(t, err, err)
			assert.NotNil(t, m)
		}
	}, 100000)
	// output:
	// Qps: 31720.414148
}

func TestXDB(t *testing.T) {
	InitXDB()
	util.GetQPS(func(n int) {
		for i := 0; i < n; i++ {
			m, err := XDB.QueryString("select now() as n")
			//log.Printf("%#v, %s\n", m, err)
			assert.Nil(t, err, err)
			assert.NotNil(t, m)
		}
	}, 100000)
	// output:
	// Qps: 23779.067495
}
