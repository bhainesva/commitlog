package main

import "fmt"

type UsedStruct struct {
	UsedField, Two string
	UnusedField int
}

type UnusedStruct struct {
	A string
}

func main() {
	h := UsedStruct{UsedField: "Hi"}
	fmt.Println(h)

	//// Shadow unused name to test
	//UnusedStruct := "hi"
	//fmt.Println(UnusedStruct)
}
