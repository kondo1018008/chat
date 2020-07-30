package main

import (
	"github.com/gorilla/websocket"
	"github.com/kondo1018008/chat/trace"
	"github.com/stretchr/objx"
	"log"
	"net/http"
)

type room struct{
	forward chan *message
	join chan *client
	leave chan *client
	clients map[*client]bool
	tracer trace.Tracer
}



func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("メッセージを受信しました：", msg.Message)
			for client := range r.clients {
				select {
				case client.send <- msg:
					//send message
					r.tracer.Trace(" -- sent to client")
				default:
					//fault send message
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request){
	socket, err := upgrader.Upgrade(w, req,nil)
	if err != nil {
		log.Fatal("ServeHTTP:",err)
		return
	}
	authCookie, err := req.Cookie("auth")
	if err != nil{
		log.Fatal("クッキーの取得に失敗しました：", err)
		return
	}

	client := &client{
		socket: socket,
		send: make(chan *message, messageBufferSize),
		room: r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() {r.leave <- client}()
	go client.write()
	client.read()
}
