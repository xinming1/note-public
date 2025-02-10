package main

import "fmt"

func foo9(a []string) {
	return
}

func main() {
	s := []string{"aceld"}
	foo9(s)
	fmt.Println(s)
}
