package main

import (
	"log"
	"net/http"
	"time"

	"github.com/CHH/eventemitter"
	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
	"github.com/gorilla/websocket"
)

const (
	listenPort = ":8080"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

var server = rpc2.NewServer()
var emitter = eventemitter.New()
var subscribedClients = make(map[*rpc2.Client]struct{})

func broadcaster() {
	clock := time.NewTicker(time.Second * 3)

	for t := range clock.C { // send message!
		log.Println("Broadcast sent")
		emitter.Emit("time", t)
	}
}

func eventHandler(t time.Time) {
	for client := range subscribedClients {
		client.Go("Receive", t, nil, nil)
		log.Println("Handled")
	}
}

func Subscribe(client *rpc2.Client, _, _ *struct{}) error {
	log.Println("Subscribing now...")
	subscribedClients[client] = struct{}{}
	return nil
}

func TestFunc(_ *rpc2.Client, t time.Time, out *int64) error {
	*out = t.Unix()
	log.Println("TestFunc called")
	return nil
}

func serveRPC(w http.ResponseWriter, r *http.Request) {
	log.Println("We have a connection!")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	server.ServeCodec(jsonrpc.NewJSONCodec(ws.UnderlyingConn()))
}

func main() {
	go broadcaster() // event creator
	server.Handle("Subscribe", Subscribe)
	server.Handle("TestFunc", TestFunc)
	emitter.On("time", eventHandler)
	server.OnDisconnect(func(client *rpc2.Client) {
		delete(subscribedClients, client)
	})

	http.HandleFunc("/rpc", serveRPC)

	err := http.ListenAndServe(listenPort, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
