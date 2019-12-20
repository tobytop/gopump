package core

import (
	"sync"
	"sync/atomic"

	"github.com/tobytop/gopump/common"
	"github.com/tobytop/gopump/message"

	"github.com/gorilla/websocket"
)

type Acceptor struct {
	mutex     sync.Mutex
	wsSocket  *websocket.Conn
	connID    uint64
	inChan    chan *message.Context
	outChan   chan *message.Context
	closeChan chan byte
	isClosed  int32
}

func NewAcceptor(connID uint64, wsSocket *websocket.Conn) (acceptor *Acceptor) {
	acceptor = &Acceptor{
		wsSocket:  wsSocket,
		connID:    connID,
		inChan:    make(chan *message.Context, 1000),
		outChan:   make(chan *message.Context, 1000),
		closeChan: make(chan byte),
		isClosed:  0,
	}
	go acceptor.readloop()
	go acceptor.writeloop()
	return
}

func (a *Acceptor) readloop() {
	var (
		msgType int
		msgData []byte
		context *message.Context
		err     error
	)
	for {
		if msgType, msgData, err = a.wsSocket.ReadMessage(); err != nil {
			goto ERR
		}
		context = &message.Context{
			Data:     msgData,
			DataType: msgType,
		}
		select {
		case a.inChan <- context:
		case <-a.closeChan:
			goto CLOSED
		}
	}
ERR:
	a.close()
CLOSED:
	close(a.inChan)
}

func (a *Acceptor) writeloop() {
	var (
		message *message.Context
		err     error
	)
	for {
		select {
		case message = <-a.outChan:
			if err = a.wsSocket.WriteMessage(message.DataType, message.Data); err != nil {
				goto ERR
			}
		case <-a.closeChan:
			goto CLOSED
		}
	}
ERR:
	a.close()
CLOSED:
	close(a.outChan)
	close(a.closeChan)
}

func (a *Acceptor) close() {
	a.wsSocket.Close()

	a.closeChan <- byte(1)

	if a.isClosed == 0 {
		atomic.StoreInt32(&a.isClosed, 1)
	}
}

func (a *Acceptor) ReadMessage() (message *message.Context, err error) {
	select {
	case message = <-a.inChan:
	case <-a.closeChan:
		err = common.ERR_CONNECTION_LOSS
	default:
		err = common.ERR_RECEIVE_MESSAGE_FULL
	}
	return
}

func (a *Acceptor) WriteMessage(message *message.Context) (err error) {
	// a.mutex.Lock()
	// defer a.mutex.Unlock()

	select {
	case a.outChan <- message:
	case <-a.closeChan:
		err = common.ERR_CONNECTION_LOSS
	default:
		err = common.ERR_SEND_MESSAGE_FULL
	}
	return
}
