package main

import (
	"./server"
)

func main() {
	_, err := server.New()
	panic(err)
}
