package main

import "ldh/Sock"

func main() {
	svr := &Sock.Server{
		KeepAlive:        300,           // seconds
		ConnectTimeout:   2,             // seconds
	}

	// Listen and serve connections at localhost:1883
	svr.ListenAndServe("tcp://:8888")
}
