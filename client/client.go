package main

import (
	"log"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
	"github.com/gorilla/websocket"
)

const (
	rpcLoc = "ws://localhost:8080/rpc"
)

func Receive(client *rpc2.Client, t time.Time, _ *struct{}) error {
	log.Printf("Time now is: %v", t)
	return nil
}

func main() {
	log.Println("Dialing...")
	ws, _, err := websocket.DefaultDialer.Dial(rpcLoc, nil)
	if err != nil {
		log.Fatalln(err)
	}

	client := rpc2.NewClientWithCodec(jsonrpc.NewJSONCodec(ws.UnderlyingConn()))
	client.Handle("Receive", Receive)

	log.Printf("Subscribing...")
	var unixTime int64
	res := client.Go("TestFunc", time.Now(), &unixTime, nil)
	go func() {
		for range res.Done {
			break
		}
		log.Printf("Phew.. %d", unixTime)
	}()

	client.Go("Subscribe", nil, nil, nil)

	if err != nil {
		log.Printf("calling function failed: %v", err)
	}
	client.Run()
}
