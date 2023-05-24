package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("listening")
		time.Sleep(10 * time.Second)
	}
}
