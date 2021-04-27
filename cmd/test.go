package main

import "fmt"

type Hello struct {
	Sub string
	Unused int
}

type Lonely struct {
	A string
}

func Wow() {
	fmt.Println("hello")
	h := Hello{Sub: "Hi"}
	fmt.Println(h.Sub)
}
