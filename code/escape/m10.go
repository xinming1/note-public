package main

func main() {
	ch := make(chan []string)

	s := []string{"aceld"}

	go func() {
		ch <- s
	}()
}
