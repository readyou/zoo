package bee

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	server         *Server
	serverAddr     string
	httpServerAddr string
	once, httpOnce sync.Once
)
var address = "127.0.0.1:5678"

type calculator struct {
}

type Calculator struct {
}

type Args struct {
	Delay time.Duration
	A, B  int
}

type args struct {
	A, B int
}

type Reply struct {
	C int
}
type reply struct {
	C int
}

func (*Calculator) Add(arg Args, reply *Reply) error {
	if arg.Delay.Seconds() > 0 {
		time.Sleep(arg.Delay)
	}
	reply.C = arg.A + arg.B
	return nil
}

func (*Calculator) Div(arg Args, reply *Reply) error {
	reply.C = arg.A / arg.B
	return nil
}

func (*Calculator) methodNotExported(arg Args, reply *Reply) error {
	return nil
}

func (*Calculator) ReturnArgNumberIncorrect(arg Args, reply *Reply) (int, error) {
	return 0, nil
}

func (*Calculator) ReturnIsNil(arg Args, reply *Reply) {
}

func (*Calculator) ReturnTypeNotError(arg Args, reply *Reply) int {
	return 0
}

func (*Calculator) ArgNumberIncorrect(arg Args, reply *Reply, extra int) error {
	return nil
}

func (*Calculator) ArgTypeNotExported(arg args, reply *Reply) error {
	return nil
}
func (*Calculator) ReplyTypeNotExported(arg Args, reply *reply) error {
	return nil
}
func (*Calculator) ReplyTypeNotPointer(arg Args, reply Reply) error {
	return nil
}

func TestServer_Register(t *testing.T) {
	server := NewServer()
	err := server.Register(new(calculator))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not exported"))

	rcvr := Calculator{}
	err = server.Register(&rcvr)
	assert.Nil(t, err, err)
	serviceMap := server.serviceMap

	s, success := serviceMap.Load("Calculator")
	assert.True(t, success)
	methodMap := s.(*service).methodMap
	assert.Len(t, methodMap, 2)
	for k, _ := range methodMap {
		assert.True(t, k == "Add" || k == "Div")
	}

	// not a pointer
	err = server.RegisterWithName(rcvr, "Calculator2")
	assert.NotNil(t, err, err)
	assert.Contains(t, err.Error(), "pointer")
}

func TestService_Call(t *testing.T) {
	s, err := buildService(new(Calculator), "", false)
	assert.Nil(t, err, err)
	mtype := s.methodMap["Add"]
	arg := Args{A: 1, B: 2}
	var reply Reply
	s.call(s.rcvr, mtype, reflect.ValueOf(arg), reflect.ValueOf(&reply))
	assert.Equal(t, 3, reply.C)
}

func TestGobClientCodec_WriteRequest(t *testing.T) {
	fileName := "test.txt"
	file1, err := os.Create(fileName)
	assert.Nilf(t, err, "open file1 error: %s", err)
	defer func() { file1.Close() }()

	gobCodec := newGobClientCodec(file1)
	// encode
	serviceName := "ServiceMethod.TestMethod"
	request := &Request{
		Seq:           1,
		ServiceMethod: serviceName,
	}
	response := &Response{
		Seq:           1,
		ServiceMethod: serviceName,
		Error:         "test error",
	}
	err = gobCodec.WriteRequest(request, response)
	assert.Nil(t, err, err)
	file1.Close()

	// decode
	file2, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	assert.Nilf(t, err, "open file1 error: %s", err)
	gobCodec = newGobClientCodec(file2)

	response2 := &Response{}
	err = gobCodec.ReadResponseHeader(response2)
	assert.Nil(t, err, err)
	assert.Equal(t, request.ServiceMethod, response2.ServiceMethod)

	body := &Response{}
	err = gobCodec.ReadResponseBody(body)
	assert.Nilf(t, err, "ReadBody error: %s", err)

	log.Println(request, response2, body)
	assert.EqualValues(t, response, body)

	file2.Close()
	file2, err = os.OpenFile(fileName, os.O_RDONLY, 0)
	assert.Nilf(t, err, "open file1 error: %s", err)
	gobCodec = newGobClientCodec(file2)

	err = gobCodec.ReadResponseHeader(response2)
	assert.Nil(t, err, err)
	assert.Equal(t, request.ServiceMethod, response2.ServiceMethod)

	// read to nil
	err = gobCodec.ReadResponseBody(nil)
	assert.Nil(t, err, err)

	file2.Close()
	os.Remove(fileName)
}

func TestGobServerCodec_ReadWriteOK(t *testing.T) {
	fileName := "test.txt"
	file1, err := os.Create(fileName)
	assert.Nilf(t, err, "open file1 error: %s", err)
	defer func() { file1.Close() }()

	gobCodec := newGobServerCodec(file1)
	// encode
	serviceName := "ServiceMethod.TestMethod"
	request := &Request{
		Seq:           1,
		ServiceMethod: serviceName,
	}
	response := &Response{
		Seq:           1,
		ServiceMethod: serviceName,
		Error:         "test error",
	}
	err = gobCodec.WriteResponse(response, response)
	assert.Nil(t, err, err)
	file1.Close()

	// decode
	file2, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	assert.Nilf(t, err, "open file1 error: %s", err)
	gobCodec = newGobServerCodec(file2)

	request2 := &Request{}
	err = gobCodec.ReadRequestHeader(request2)
	assert.Nil(t, err, err)
	assert.EqualValues(t, request, request2)

	body := &Response{}
	err = gobCodec.ReadRequestBody(body)
	assert.Nilf(t, err, "ReadBody error: %s", err)

	log.Println(request, request2, body)
	assert.EqualValues(t, response, body)

	file2.Close()
	file2, err = os.OpenFile(fileName, os.O_RDONLY, 0)
	assert.Nilf(t, err, "open file1 error: %s", err)
	gobCodec = newGobServerCodec(file2)

	err = gobCodec.ReadRequestHeader(request2)
	assert.Nil(t, err, err)
	assert.EqualValues(t, request, request2)

	// read to nil
	err = gobCodec.ReadRequestBody(nil)
	assert.Nil(t, err, err)

	file2.Close()
	os.Remove(fileName)
}

func StartServer() {
	once.Do(func() {
		server = NewServer()
		err := server.Register(new(Calculator))
		if err != nil {
			log.Fatalf("server.Register error: %s", err)
			return
		}

		go server.Start(address)
	})
}

func TestRPC(t *testing.T) {
	StartServer()
	client := NewClient()
	err := client.Dial(address)
	assert.Nil(t, err, err)
	defer client.Close()

	arg := Args{A: 1, B: 2}
	start := time.Now()
	n := 500000
	for i := 0; i < n; i++ {
		reply := new(Reply)
		err = client.Call("Calculator.Add", arg, &reply)
		if err != nil {
			t.Fatalf("rpc error: %s", err.Error())
		}
		if reply.C != 3 {
			t.Fatalf("rpc result incorrect, expect=%d, actual=%d\n", 3, *reply)
		}
	}
	end := time.Now()
	seconds := end.Sub(start).Seconds()
	log.Printf("Qps: %f\n", float64(n)/seconds)
}

func BenchmarkRPC(b *testing.B) {
	log.Println("BenchmarkRPC started")
	StartServer()

	client := NewClient()
	err := client.Dial(address)
	assert.Nil(b, err, err)
	defer client.Close()

	arg := Args{A: 1, B: 2}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reply := new(Reply)
			err = client.Call("Calculator.Add", arg, &reply)
			if err != nil {
				b.Fatalf("rpc error: %s", err.Error())
			}
			if reply.C != 3 {
				b.Fatalf("rpc result incorrect, expect=%d, actual=%d\n", 3, *reply)
			}
		}
	})
}
