package main

import "github.com/panjichang1990/tianzong/child"

func main() {
	child.SetPort(5544)
	child.SetGateAddress("127.0.0.1:8888")
	child.Run()
}
