package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	ticker := time.Tick(time.Millisecond * 100)
	for range ticker {
		n := rand.Intn(100)
		if n%10 == 0 {
			panic("error")
		}

		fmt.Println("this is dummy")

	}
}
