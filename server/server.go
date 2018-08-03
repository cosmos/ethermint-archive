package server

import (
	"github.com/gorilla/rpc"
	"sync"
	"net/http"
	"github.com/pkg/errors"
	"strings"
	"fmt"
	"reflect"
	"unicode/utf8"
	"unicode"
)

var (
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfRequest = reflect.TypeOf((*http.Request)(nil)).Elem()
)

type BeforeFilterInfo struct {
	Method   string
	Request  *http.Request
	Response http.ResponseWriter
}

type AfterFilterInfo struct {
	Method   string
	Error    error
	Request  *http.Request
	Response http.ResponseWriter
}

type BeforeFilter func(b *BeforeFilterInfo) bool
type AfterFilter func(a *AfterFilterInfo) bool

type handlerSpec struct {
	needsHTTP bool
	handler   reflect.Value
	argsType  reflect.Type
	replyType reflect.Type
}

type Server struct {
	codecs        map[string]rpc.Codec
	handlers      map[string]*handlerSpec
	beforeFilters []BeforeFilter
	afterFilters  []AfterFilter
	handlerMtx    sync.Mutex
}

func NewServer() *Server {
	return &Server{
		codecs:        make(map[string]rpc.Codec),
		handlers:      make(map[string]*handlerSpec),
		beforeFilters: make([]BeforeFilter, 0),
		afterFilters:  make([]AfterFilter, 0),
	}
}

// Adds a new codec to the server. Identical to Gorilla's implementation.
func (s *Server) RegisterCodec(codec rpc.Codec, contentType string) {
	s.codecs[strings.ToLower(contentType)] = codec
}

// Adds a new handler to the server.
//
// Handlers must expose an args and a reply argument of pointer types to exported
// structs or builtins. There is an optional first argument of type *http.Request.
func (s *Server) RegisterHandler(name string, handler interface{}) error {
	return s.registerHandler(name, handler)
}


// Appends a before filter to the list of available before filters. Before filters are
// functions that are run prior to execution of the requested RPC method but after the
// RPC request is validated. A return value of false from a return filter will cancel
// execution of the RPC method. Useful for things like authentication and CORS header
// checking.
func (s *Server) AddBeforeFilter(before BeforeFilter) {
	s.beforeFilters = append(s.beforeFilters, before)
}


// Appends an after filter to the list of available after filters. After filters are
// functions that are run after execution of the requests RPC method but prior to sending
// the response. A return value of false from an after filter will cancel returning a response.
func (s *Server) AddAfterFilter(after AfterFilter) {
	s.afterFilters = append(s.afterFilters, after)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, 405, "rpc: POST method required, received "+r.Method)
		return
	}

	contentType := r.Header.Get("Content-Type")
	idx := strings.Index(contentType, ";")
	if idx != -1 {
		contentType = contentType[:idx]
	}

	var codec rpc.Codec
	if contentType == "" && len(s.codecs) == 1 {
		for _, c := range s.codecs {
			codec = c
		}
	} else if codec = s.codecs[strings.ToLower(contentType)]; codec == nil {
		s.writeError(w, 415, "rpc: unrecognized Content-Type: "+contentType)
		return
	}

	codecReq := codec.NewRequest(r)
	hdlrName, codecErr := codecReq.Method()
	if codecErr != nil {
		s.writeError(w, 400, codecErr.Error())
		return
	}

	hdlr, ok := s.handlers[hdlrName]
	if !ok {
		s.writeError(w, 400, "rpc: can't find method "+hdlrName)
		return
	}

	args := reflect.New(hdlr.argsType)
	if errRead := codecReq.ReadRequest(args.Interface()); errRead != nil {
		s.writeError(w, 400, errRead.Error())
	}

	bfi := &BeforeFilterInfo{
		Method:   hdlrName,
		Request:  r,
		Response: w,
	}
	for _, bf := range s.beforeFilters {
		if !bf(bfi) {
			return
		}
	}

	reply := reflect.New(hdlr.replyType)
	var errValue []reflect.Value
	if hdlr.needsHTTP {
		errValue = hdlr.handler.Call([]reflect.Value{
			reflect.ValueOf(r),
			args,
			reply,
		})
	} else {
		errValue = hdlr.handler.Call([]reflect.Value{
			args,
			reply,
		})
	}

	var errResult error
	errInter := errValue[0].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}

	afi := &AfterFilterInfo{
		Method:   hdlrName,
		Request:  r,
		Response: w,
		Error:    errResult,
	}

	for _, af := range s.afterFilters {
		if !af(afi) {
			return
		}
	}

	if errWrite := codecReq.WriteResponse(w, reply.Interface(), errResult); errWrite != nil {
		s.writeError(w, 400, errWrite.Error())
	}
}

func (s *Server) registerHandler(name string, handler interface{}) error {
	s.handlerMtx.Lock()
	defer s.handlerMtx.Unlock()

	if _, exists := s.handlers[name]; exists {
		return errors.New("handler already registered for method name " + name)
	}

	hdlrType := reflect.TypeOf(handler)
	if hdlrType.Kind() != reflect.Func {
		return errors.New("handler must be a function")
	}

	var needsHTTP bool
	var argOffset int
	numIn := hdlrType.NumIn()
	if numIn == 2 {
		needsHTTP = false
		argOffset = 0
	} else if numIn == 3 {
		if hdlrType.In(0) != typeOfRequest {
			return errors.New("handler first argument must be *http.Request")
		}

		needsHTTP = true
		argOffset = 1
	} else {
		return errors.New("handler must have either two or three arguments")
	}

	args := hdlrType.In(argOffset)
	if args.Kind() != reflect.Ptr || !isExportedOrBuiltin(args) {
		return errors.New("args must be a pointer to an exported struct or builtin")
	}
	reply := hdlrType.In(1 + argOffset)
	if reply.Kind() != reflect.Ptr || !isExportedOrBuiltin(reply) {
		return errors.New("reply must be a pointer to an exported struct or builtin")
	}

	if hdlrType.NumOut() != 1 || hdlrType.Out(0) != typeOfError {
		return errors.New("handler must have one return value of type error")
	}

	s.handlers[name] = &handlerSpec{
		needsHTTP: needsHTTP,
		handler:   reflect.ValueOf(handler),
		argsType:  args.Elem(),
		replyType: reply.Elem(),
	}
	return nil
}

func (s *Server) writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, msg)
}

// isExported returns true of a string is an exported (upper case) name.
func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// isExportedOrBuiltin returns true if a type is exported or a builtin.
func isExportedOrBuiltin(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}
