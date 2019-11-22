package core

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	mutex       sync.Mutex
	mux         *http.ServeMux
	server      *http.Server
	Dispatchers map[string]*Dispatcher
}

var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleConnect(resp http.ResponseWriter, req *http.Request) {
	var (
		err      error
		wsSocket *websocket.Conn
		connID   uint64
		a        *Acceptor
	)
	//websocket握手
	if wsSocket, err = wsUpgrader.Upgrade(resp, req, nil); err != nil {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	connID = uint64(time.Now().Unix())
	a = NewAcceptor(connID, wsSocket)

	s.Dispatchers[req.URL.Path].mutex.Lock()
	defer s.Dispatchers[req.URL.Path].mutex.Unlock()

	s.Dispatchers[req.URL.Path].Acceptors[req] = a
	for r, a := range s.Dispatchers[req.URL.Path].Acceptors {
		if a.isClosed == 1 {
			delete(s.Dispatchers[req.URL.Path].Acceptors, r)
		}
	}
}

//添加路径
func (s *Server) AddRouter(path string, handle HandleFunc) {
	d := NewDispatcher(path, handle)
	s.Dispatchers[path] = d

	s.mux.HandleFunc(path, s.handleConnect)

}

func NewServer() (s *Server) {
	var (
		mux    *http.ServeMux
		server *http.Server
	)
	mux = http.NewServeMux()
	server = &http.Server{
		ReadTimeout:  time.Duration(1000) * time.Millisecond,
		WriteTimeout: time.Duration(1000) * time.Millisecond,
		Handler:      mux,
	}

	s = &Server{
		server:      server,
		mux:         mux,
		Dispatchers: make(map[string]*Dispatcher),
	}
	return
}

func (s *Server) Start(port int) {
	var (
		listener net.Listener
		err      error
	)
	// 监听端口
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(port)); err != nil {
		return
	}

	s.server.Serve(listener)
}
