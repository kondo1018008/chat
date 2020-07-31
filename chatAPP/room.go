package main

import (
	"github.com/gorilla/websocket"
	"github.com/kondo1018008/chat/trace"
	"github.com/stretchr/objx"
	"log"
	"net/http"
	"os"
)

type room struct{
	forward chan *message
	join chan *client
	leave chan *client
	clients map[*client]bool
	tracer trace.Tracer
}


//Javaでいうところのコンストラクタ。値の初期化をしている。
func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
		tracer: trace.New(os.Stdout),//本ではmain.goで初期化しているが、オリジナルでここで初期化している。
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("新しいクライアントが参加しました")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("クライアントが退出しました")
		case msg := <-r.forward:
			r.tracer.Trace("メッセージを受信しました：", msg.Message)
			for client := range r.clients {
				select {
				case client.send <- msg:
					//send message
					r.tracer.Trace(" -- クライアントに送信しました")
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
	//http通信をwebsocketにアップグレードする
	socket, err := upgrader.Upgrade(w, req,nil)
	if err != nil {
		log.Fatal("ServeHTTP:",err)
		return
	}
	//クッキーの取得
	authCookie, err := req.Cookie("auth")
	if err != nil{
		log.Fatal("クッキーの取得に失敗しました：", err)
		return
	}
	//クライアントインスタンスの生成（厳密にはインスタンスではない）
	client := &client{
		socket: socket,
		send: make(chan *message, messageBufferSize),
		room: r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() {r.leave <- client}()//遅延処理
	go client.write()
	client.read()
}
