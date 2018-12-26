package main

import "GoStatistics/server"

func main() {
	_, err := server.InitRpcServer()
	panic(err)
}
