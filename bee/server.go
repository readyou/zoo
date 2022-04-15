package bee

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"git.garena.com/xinlong.wu/zoo/util"
	"go/token"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

type ServiceError string

const DefaultRPCPath = "/rpc"
const DefaultDebugPath = "/rpc/debug"

var DefaultServer = NewServer()

func (err *ServiceError) Error() string {
	return string(*err)
}

// rpc request header
type Request struct {
	Seq           uint64 // unique sequence of a request
	ServiceMethod string // format: "Service.Method"
}

// rpc reponse header
type Response struct {
	Seq           uint64 // echoes that of the request
	ServiceMethod string // echoes that of the request
	Error         string // error message if exists
}

type methodType struct {
	sync.Mutex      // protect callTimes
	callTimes  uint // method call times
	method     reflect.Method
	ArgType    reflect.Type
	ReplyType  reflect.Type
}

type service struct {
	name      string                 // service name
	rcvr      reflect.Value          // receiver of methods for the service
	typ       reflect.Type           // type of the receiver
	methodMap map[string]*methodType // registered methods
}

type ServerCodec interface {
	ReadRequestHeader(request *Request) error
	ReadRequestBody(body any) error
	WriteResponse(response *Response, body any) error
	Close() error
}

type ServerCodecNewFunc func(rwc io.ReadWriteCloser) ServerCodec

type Server struct {
	codecNewFunc ServerCodecNewFunc
	serviceMap   sync.Map // map[string]*service
}

func (server *Server) Register(rcvr any) error {
	return server.register(rcvr, "", false)
}

func (server *Server) RegisterWithName(rcvr any, name string) error {
	return server.register(rcvr, name, true)
}

func (server *Server) register(rcvr any, name string, overrideName bool) error {
	s, err := buildService(rcvr, name, overrideName)
	if err != nil {
		return err
	}

	if _, exists := server.serviceMap.LoadOrStore(s.name, s); exists {
		msg := fmt.Sprintf("rpc.Register: service name exists already %q", s.typ)
		log.Println(msg)
		return errors.New(msg)
	} else {
		log.Printf("rpc.Register: %s, %#v\n", s.name, s)
	}
	return nil
}

func buildService(rcvr any, name string, overrideName bool) (*service, error) {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if overrideName {
		sname = name
	}
	if sname == "" {
		msg := fmt.Sprintf("rpc.Register: no service name for type %q", s.typ)
		log.Println(msg)
		return nil, errors.New(msg)
	}
	if !token.IsExported(sname) && !overrideName {
		msg := fmt.Sprintf("rpc.Register: service %q is not exported", sname)
		log.Println(msg)
		return nil, errors.New(msg)
	}
	s.name = sname

	methodMap := suitableMethod(s.typ)
	if len(methodMap) == 0 {
		methodMap = suitableMethod(reflect.PointerTo(s.typ))
		msg := fmt.Sprintf("rpc.Register: no exported suitable method of %q exists", s.typ)
		if len(methodMap) > 0 {
			msg = fmt.Sprintf("rpc.Register: please use pointer of %q to call register", s.typ)
		}
		log.Println(msg)
		return nil, errors.New(msg)
	}
	s.methodMap = methodMap
	return s, nil
}

func suitableMethod(typ reflect.Type) map[string]*methodType {
	methodMap := make(map[string]*methodType)
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		// method must be exported
		if !method.IsExported() {
			continue
		}
		mname := method.Name
		mtype := method.Type
		if mtype.NumIn() != 3 {
			log.Println("rpc.Register: method needs three ins: receiver, *args, *reply")
			continue
		}
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			log.Printf("rpc.Register: argument type of method %q is not exported: %q\n", mname, argType)
			continue
		}
		replyType := mtype.In(2)
		if !isExportedOrBuiltinType(replyType) {
			log.Printf("rpc.Register: reply type of method %q is not exported: %q\n", mname, replyType)
			continue
		}
		if replyType.Kind() != reflect.Pointer {
			log.Printf("rpc.Register: reply type of method %q should be pointer: %q\n", mname, replyType)
			continue
		}
		if mtype.NumOut() != 1 {
			log.Printf("rpc.Register: method %q should return only one error type\n", mname)
			continue
		}
		retType := mtype.Out(0)
		if retType != errorType {
			log.Printf("rpc.Register: return type of method %q should be error: %q\n", mname, retType)
			continue
		}
		methodMap[mname] = &methodType{
			ArgType:   argType,
			ReplyType: replyType,
			method:    method,
		}
		//log.Printf("rpc.Register: register %q success\n", mname)
	}
	return methodMap
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

func (*service) call(rcvr reflect.Value, mtype *methodType, arg reflect.Value, reply reflect.Value) (err error) {
	defer func() {
		if e := recover(); e != nil {
			message := fmt.Sprintf("%s", e)
			log.Printf("rpc.call: %s\n\n", util.Err.Trace(message))
			err = errors.New(message)
		}
	}()

	method := mtype.method
	values := method.Func.Call([]reflect.Value{rcvr, arg, reply})
	e := values[0].Interface()
	if e != nil {
		err = e.(error)
		return
	}
	return
}

type gobServerCodec struct {
	rwc    io.ReadWriteCloser
	encBuf *bufio.Writer
	enc    *gob.Encoder
	dec    *gob.Decoder
	closed bool
}

func (g *gobServerCodec) ReadRequestHeader(request *Request) error {
	return g.dec.Decode(request)
}

func (g *gobServerCodec) ReadRequestBody(body any) error {
	return g.dec.Decode(body)
}

func (g *gobServerCodec) WriteResponse(response *Response, body any) (err error) {
	defer func() {
		g.encBuf.Flush()
		if err != nil {
			g.Close()
		}
	}()

	err = g.enc.Encode(response)
	if err == nil {
		err = g.enc.Encode(body)
	}
	return err
}

func (g *gobServerCodec) Close() error {
	if g.closed {
		return nil
	}
	g.closed = true
	return g.rwc.Close()
}

// ensure gobServerCodec implements ServerCodec
var _ ServerCodec = (*gobServerCodec)(nil)

func newGobServerCodec(rwc io.ReadWriteCloser) *gobServerCodec {
	writer := bufio.NewWriter(rwc)
	return &gobServerCodec{
		rwc:    rwc,
		encBuf: writer,
		enc:    gob.NewEncoder(writer),
		dec:    gob.NewDecoder(rwc),
	}
}

func NewServer() *Server {
	return &Server{
		codecNewFunc: func(rwc io.ReadWriteCloser) ServerCodec {
			return newGobServerCodec(rwc)
		},
	}
}

func NewServerWithCodec(codecNewFunc ServerCodecNewFunc) *Server {
	return &Server{
		codecNewFunc: codecNewFunc,
	}
}

// should be called after server been inited properly(call NewServer/NewServerWithCodec)
func (server *Server) Start(address string) {
	log.Println("rpc serve at:", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("listen error: %s\n", address)
	}
	server.accept(listener)
}

func (server *Server) StartHTTP(address string) {
	http.Handle(DefaultRPCPath, server)
	http.Handle(DefaultDebugPath, debugHTTP{server})
	if err := http.ListenAndServe(address, server); err != nil {
		log.Fatalf("start http server error: %s\n", err.Error())
	}
	server.Start(address)
}

func (server *Server) accept(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("server accept error: %s\n", err)
			return
		}
		log.Printf("client connected: %s\n", conn.RemoteAddr())
		go server.serveConn(conn)
	}
}

var invalidRequest = struct{}{}

func (server *Server) serveConn(conn io.ReadWriteCloser) {
	sendingMutex := new(sync.Mutex) // one connection should send one request a time
	serverCodec := server.codecNewFunc(conn)
	for {
		svc, mtype, req, argv, replyv, keepReading, err := server.readRequest(serverCodec)
		if err != nil {
			if err != io.EOF {
				log.Println("rpc.readRequest: ", err)
			}
			if !keepReading {
				break
			}
			if req != nil {
				server.sendResponse(sendingMutex, serverCodec, req, invalidRequest, err.Error())
			}
			continue
		}
		go func() {
			errMsg := ""
			if err := svc.call(svc.rcvr, mtype, argv, replyv); err != nil {
				errMsg = err.Error()
			}
			server.sendResponse(sendingMutex, serverCodec, req, replyv.Interface(), errMsg)
		}()
	}
	serverCodec.Close()
}

func (server *Server) readRequest(codec ServerCodec) (svc *service, mtype *methodType, req *Request, argv, replyv reflect.Value, keepReading bool, err error) {
	svc, mtype, req, keepReading, err = server.readRequestHeader(codec)
	if err != nil {
		if keepReading {
			// discard body
			codec.ReadRequestBody(nil)
		}
		return
	}
	argIsValue := false
	if mtype.ArgType.Kind() == reflect.Pointer {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}
	if err = codec.ReadRequestBody(argv.Interface()); err != nil {
		return
	}
	if argIsValue {
		argv = argv.Elem()
	}

	replyv = reflect.New(mtype.ReplyType.Elem())
	switch replyv.Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(mtype.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(mtype.ReplyType.Elem(), 0, 0))
	}

	return
}

func (server *Server) readRequestHeader(codec ServerCodec) (svc *service, mtype *methodType, req *Request, keepReading bool, err error) {
	req = &Request{}
	err = codec.ReadRequestHeader(req)
	if err != nil {
		req = nil
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		}
		err = errors.New("rpc server codec request error: " + err.Error())
		return
	}

	// Read request header success, but can not find correct rpc service, keep reading next request
	keepReading = true
	dot := strings.Index(req.ServiceMethod, ".")
	if dot <= 0 {
		err = errors.New("rpc: service/method request ill-formed: " + req.ServiceMethod)
		return
	}
	serviceName := req.ServiceMethod[:dot]
	methodName := req.ServiceMethod[dot+1:]
	serviceI, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc: can not find service: " + serviceName)
		return
	}

	svc = serviceI.(*service)
	mtype = svc.methodMap[methodName]
	if mtype == nil {
		err = errors.New("rpc: can not find method: " + methodName)
		return
	}
	return
}

func (server *Server) sendResponse(sendingMutex *sync.Mutex, codec ServerCodec, req *Request, reply any, errMsg string) {
	sendingMutex.Lock()
	response := Response{ServiceMethod: req.ServiceMethod, Seq: req.Seq}
	if errMsg != "" {
		response.Error = errMsg
		reply = invalidRequest
	}
	err := codec.WriteResponse(&response, reply)
	if err != nil {
		log.Println("rpc: writing response error:", err)
	}
	sendingMutex.Unlock()
}

// Can connect to RPC service using HTTP CONNECT to rpcPath.
// ServeHTTP implements an http.Handler that answers RPC requests.
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.serveConn(conn)
}
