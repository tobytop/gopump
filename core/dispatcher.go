package core

import (
	"gopump/message"
	"net/http"
	"sync"
	"time"
)

type HandleFunc func(s *Sender)

type Sender struct {
	Reqs    []*http.Request
	Context *message.Context
}

type Dispatcher struct {
	path       string
	mutex      sync.Mutex
	handleFunc HandleFunc
	Acceptors  map[*http.Request]*Acceptor
	sendchan   chan map[*http.Request]*message.Context
}

func (d *Dispatcher) findloop() {
	var messages map[*http.Request]*message.Context
	for {
		<-time.NewTimer(time.Duration(200) * time.Millisecond).C
		messages = d.getMessage()

		if len(messages) > 0 {
			d.sendchan <- messages
			messages = make(map[*http.Request]*message.Context)
		}
	}
}

func (d *Dispatcher) sendloop() {
	//var messages map[*http.Request]*message.Context
	for {
		messages := <-d.sendchan
		for r, v := range messages {
			v.Req = r
			d.handler(v)
		}
	}
}

func (d *Dispatcher) handler(message *message.Context) {
	d.mutex.Lock()
	sender := &Sender{
		Reqs:    []*http.Request{},
		Context: message,
	}
	for k, _ := range d.Acceptors {
		sender.Reqs = append(sender.Reqs, k)
	}
	d.mutex.Unlock()

	d.handleFunc(sender)
	d.sendMessage(sender)
}

func (d *Dispatcher) sendMessage(s *Sender) {
	//d.mutex.Lock()
	//defer d.mutex.Unlock()
	for _, v := range s.Reqs {
		if a, ok := d.Acceptors[v]; ok {
			a.WriteMessage(s.Context)
		}
	}
}

func (d *Dispatcher) getMessage() (msg map[*http.Request]*message.Context) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	msg = make(map[*http.Request]*message.Context)
	for r, a := range d.Acceptors {
		if a.isClosed {
			delete(d.Acceptors, r)
		} else if m, err := a.ReadMessage(); err == nil && m != nil {
			msg[r] = m
		}
	}

	return
}

func NewDispatcher(path string, handleFunc HandleFunc) (d *Dispatcher) {
	d = &Dispatcher{
		path:       path,
		handleFunc: handleFunc,
		Acceptors:  make(map[*http.Request]*Acceptor),
		sendchan:   make(chan map[*http.Request]*message.Context, 1000),
	}

	go d.findloop()
	go d.sendloop()
	return
}
