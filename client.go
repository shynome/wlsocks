package main

import (
	"fmt"
	"net"
	"time"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/shynome/wlsocks/client"
)

func runClient() {

	config := client.Config{
		Lens: args.lens,
		User: args.user,
		Ices: args.ices,
	}
	client := client.New(config)
	addr := &net.TCPAddr{IP: net.ParseIP(args.addr), Port: args.port}
	l := try.To1(net.ListenTCP("tcp", addr))

	for {
		err := clientServe(client, l)
		if err != nil {
			fmt.Println("err:", err)
		}
		time.Sleep(time.Second)
		fmt.Println("-----------------")
		fmt.Println("retry connect")
	}
}

func clientServe(client *client.Client, l net.Listener) (err error) {
	defer err2.Return(&err)
	try.To(client.Connect())
	fmt.Println("server connected")
	go client.Serve(l)
	<-client.Done()
	fmt.Println("server disconnected")
	return
}
