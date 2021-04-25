package simple

import "fmt"

type Wow struct {
	Hi    int    // Comment
	Bye   int    // Comment
	Hello string // Comment
}

// Print prints the wow
func Print(w Wow) {
	fmt.Println("Hi,: ", w.Hi)
}

// Check checks the wow
func Check(w Wow) bool {
	fmt.Println(w.Hello)

	// The magic
	return w.Bye > 29
}
