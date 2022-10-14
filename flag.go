package main

import "flag"

var args struct {
	lens string
	user string
	pass string
	ices string

	addr string
	port int
}

func init() {
	flag.StringVar(&args.lens, "lens", "https://lens.slive.fun/", "lens it is for passing msgs")
	flag.StringVar(&args.user, "user", "", "required")
	flag.StringVar(&args.pass, "pass", "", "pass, required when this run as server ")

	flag.StringVar(&args.ices, "ices", "", "file or http(s) link to get ice servers")

	flag.StringVar(&args.addr, "addr", "", "")
	flag.IntVar(&args.port, "port", 0, "the default value is 1080 when run as client")
}
