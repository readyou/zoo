package bee

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClient_Call_timeout(t *testing.T) {
	StartServer()
	timeout := time.Second * 3
	client := NewClientWithOption(ClientOption{
		ConnectTimeout:  time.Second * 10,
		ResponseTimeout: timeout,
	})
	err := client.Start(address)
	assert.Nil(t, err, err)
	defer client.Close()

	arg := Args{A: 1, B: 2, Delay: timeout + 10}
	reply := new(Reply)
	err = client.Call("Calculator.Add", arg, &reply)
	assert.NotNil(t, err, err)
	assert.Equal(t, ErrResponseTimeout, err)

	arg = Args{A: 1, B: 0}
	reply = new(Reply)
	err = client.Call("Calculator.Div", arg, &reply)
	assert.NotNil(t, err, "should throw error")

	arg = Args{A: 1, B: 2}
	reply = new(Reply)
	err = client.Call("Calculator.Add", arg, &reply)
	assert.Nil(t, err, err)
	assert.Equal(t, arg.A+arg.B, reply.C)
}
