package bee

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var ErrShutDown = errors.New("connection is shut down")
var ErrResponseTimeout = errors.New("get response timeout")

// Can connect to RPC service using HTTP CONNECT to rpcPath.
var connected = "200 Connected to Go RPC"

type Call struct {
	Seq           uint64
	ServiceMethod string     // fmt: "Service.Method".
	Args          any        // The argument to the function (*struct).
	Reply         any        // The reply from the function (*struct).
	Error         error      // After completion, the error status.
	Done          chan *Call // Receives *Call when Go is completed.
}

func (call *Call) done() {
	select {
	case call.Done <- call:
	// ok
	default:
		// Won't block here, it's caller's responsibility to make sure the channel has enough buffer
		log.Println("Discard response due to not enough capacity of Done chan")
	}
}

type ClientCodecNewFunc func(rwc io.ReadWriteCloser) ClientCodec

type Client struct {
	codecNewFunc ClientCodecNewFunc
	codec        ClientCodec
	reqMutex     sync.Mutex       // protects request
	request      Request          // one client should send one request at a time
	mutex        sync.Mutex       // protects following
	pending      map[uint64]*Call // calls whose response is not received yet
	seq          uint64           // current request seq
	isClosing    bool             // user has called Close
	isShutdown   bool             // server has told us to stop
	option       ClientOption
	isInited     bool
}

type ClientOption struct {
	ConnectTimeout  time.Duration // timeout for connect
	ResponseTimeout time.Duration // timeout for waiting response
}

type ClientCodec interface {
	ReadResponseHeader(response *Response) error
	ReadResponseBody(body any) error
	WriteRequest(req *Request, body any) error
	Close() error
}

type gobClientCodec struct {
	rwc    io.ReadWriteCloser
	encBuf *bufio.Writer
	enc    *gob.Encoder
	dec    *gob.Decoder
}

func (g *gobClientCodec) ReadResponseHeader(response *Response) error {
	return g.dec.Decode(response)
}

func (g *gobClientCodec) ReadResponseBody(body any) error {
	return g.dec.Decode(body)
}

func (g *gobClientCodec) WriteRequest(req *Request, body any) (err error) {
	if err = g.enc.Encode(req); err != nil {
		return
	}
	if err = g.enc.Encode(body); err != nil {
		return
	}
	return g.encBuf.Flush()
}

func (g *gobClientCodec) Close() error {
	return g.rwc.Close()
}

// ensure gobClientCodec implements ClientCodec
var _ ClientCodec = (*gobClientCodec)(nil)

func newGobClientCodec(rwc io.ReadWriteCloser) ClientCodec {
	writer := bufio.NewWriter(rwc)
	return &gobClientCodec{
		rwc:    rwc,
		encBuf: writer,
		enc:    gob.NewEncoder(writer),
		dec:    gob.NewDecoder(rwc),
	}
}

func NewClient() *Client {
	return NewClientWithOption(ClientOption{
		ConnectTimeout:  time.Second * 60,
		ResponseTimeout: time.Second * 60,
	})
}

func NewClientWithOption(option ClientOption) *Client {
	return NewClientWithCodec(option, func(rwc io.ReadWriteCloser) ClientCodec {
		return newGobClientCodec(rwc)
	})
}

func NewClientWithCodec(option ClientOption, codecNewFunc ClientCodecNewFunc) *Client {
	client := Client{
		codecNewFunc: codecNewFunc,
		pending:      make(map[uint64]*Call),
		option:       option,
		isInited:     true,
	}
	return &client
}

func (client *Client) DialHTTP(addr string) (err error) {
	if !client.isInited {
		msg := "rpc client: client not inited, please call NewClient/NewClientWithCodec to create client"
		log.Println(msg)
		return errors.New(msg)
	}
	conn, err := net.DialTimeout("tcp", addr, client.option.ConnectTimeout)
	if err != nil {
		log.Fatalln("rpc client: connecting server error:", err)
		return
	}
	io.WriteString(conn, "CONNECT "+DefaultRPCPath+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil {
		return
	}
	if resp.Status != connected {
		err = errors.New("unexpected HTTP response: " + resp.Status)
		conn.Close()
		return
	}

	client.codec = client.codecNewFunc(conn)
	log.Printf("rpc.Client connected: %s\n", addr)
	go client.run()
	return
}

func (client *Client) Dial(addr string) (err error) {

	if !client.isInited {
		msg := "rpc client: client not inited, please call NewClient/NewClientWithCodec to create client"
		log.Println(msg)
		return errors.New(msg)
	}
	conn, err := net.DialTimeout("tcp", addr, client.option.ConnectTimeout)
	if err != nil {
		log.Fatalln("rpc client: connecting server error:", err)
		return
	}
	client.codec = client.codecNewFunc(conn)
	log.Printf("rpc.Client connected: %s\n", addr)
	go client.run()
	return nil
}

func (client *Client) run() {
	var err error
	for err == nil {
		response := Response{}
		err = client.codec.ReadResponseHeader(&response)
		if err != nil {
			break
		}
		client.mutex.Lock()
		call := client.pending[response.Seq]
		delete(client.pending, response.Seq)
		client.mutex.Unlock()
		switch {
		case call == nil:
			// call not found, ignore response body
			//log.Printf("rpc client: call not found: %v\n", response)
			// if read response success, continue(err == nil)
			err = client.codec.ReadResponseBody(nil)
			if err != nil {
				err = errors.New("rpc client: reading error body: " + err.Error())
			}
		case response.Error != "":
			call.Error = ServerError(response.Error)
			err = client.codec.ReadResponseBody(nil)
			if err != nil {
				err = errors.New("rpc client: reading error body: " + err.Error())
			}
			call.done()
		default:
			err = client.codec.ReadResponseBody(call.Reply)
			if err != nil {
				call.Error = errors.New("rpc client: reading body: " + err.Error())
			}
			call.done()
		}
	}
	// close the client
	client.reqMutex.Lock()
	client.mutex.Lock()
	client.isShutdown = true
	closing := client.isClosing
	if err == io.EOF {
		if closing {
			err = ErrShutDown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	// make pending calls done
	for _, v := range client.pending {
		v.Error = err
		v.done()
	}
	client.mutex.Unlock()
	client.reqMutex.Unlock()
	if err != io.EOF && !closing {
		log.Println("rpc client protocol error", err)
	}
}

func (client *Client) Close() error {
	client.mutex.Lock()
	if client.isClosing {
		client.mutex.Lock()
		return ErrShutDown
	}
	client.isClosing = true
	client.mutex.Unlock()
	if client.codec != nil {
		return client.codec.Close()
	}
	return nil
}

func (client *Client) Call(serviceMethod string, args any, reply any) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	timer := time.NewTimer(client.option.ResponseTimeout)
	select {
	case <-timer.C:
		client.mutex.Lock()
		req := call.Seq
		delete(client.pending, req)
		client.mutex.Unlock()
		call.Error = ErrResponseTimeout
		call.done()
	case <-call.Done:
		// ok
	}
	return call.Error
}

func (client *Client) Go(serviceMethod string, args any, reply any, done chan *Call) *Call {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10)
	}
	call.Done = done

	// send request should be sync: make sure request sent success.
	client.send(call)
	return call
}

func (client *Client) send(call *Call) {
	if client.codec == nil {
		log.Println("please call client.Dial(addr) first")
		return
	}
	client.reqMutex.Lock()
	defer client.reqMutex.Unlock()

	client.mutex.Lock()
	// client closed
	if client.isClosing || client.isShutdown {
		client.mutex.Unlock()
		call.Error = ErrShutDown
		call.done()
		return
	}
	seq := client.seq
	client.seq += 1
	client.request.Seq = seq
	client.request.ServiceMethod = call.ServiceMethod
	call.Seq = seq
	client.pending[seq] = call
	client.mutex.Unlock()

	request := &client.request
	if err := client.codec.WriteRequest(request, call.Args); err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		if call != nil {
			call.Error = err
			call.done()
		}
	} else {
		//log.Printf("rpc.Client: send request success: %#v\n", client.request)
	}
}
