package main

import (
	"github.com/panjichang1990/tianzong/gate"
)

func main() {
	gate.SetAuthAddress("127.0.0.1:5055")
	gate.Run()

}
