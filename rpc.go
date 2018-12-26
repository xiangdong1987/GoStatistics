package main

import "./server"

func main() {
	_, err := server.InitRpcServer()
	panic(err)
}
