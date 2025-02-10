package main

import "fmt"

func foo8(a *int) {
	return
}

func main() {
	data := 10
	f := foo8
	f(&data)
	fmt.Println(data)
}
