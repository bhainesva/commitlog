package main

import fmt "fmt"

type UsedType struct {
	UsedField, UnusedField string
	SecondUnusedField      int
}

type UnusedType struct {
	A string
}

func UsedFunc(u UsedType) {
	type UnusedNestedType struct{}
	fmt.Println(u)
}

func main() {
	h := UsedType{UsedField: "Hi"}
	UsedFunc(h)

	//// Testing with shadowed name
	UnusedStruct := "hi"
	fmt.Println(UnusedStruct)
}
