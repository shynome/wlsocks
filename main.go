package main

import (
	"flag"
	"fmt"

	"github.com/armon/go-socks5"
	"github.com/lainio/err2/try"
)

func main() {
	flag.Parse()
	if args.pass != "" {
		runServer()
		return
	}
	if args.user != "" {
		if args.port == 0 {
			args.port = 1080
		}
		runClient()
		return
	}
	server := try.To1(socks5.New(&socks5.Config{}))
	try.To(server.ListenAndServe("tcp", fmt.Sprintf("%s:%v", args.addr, args.port)))
}
