package main

import (
	"fmt"
)

func main() {
	race()
	channels()
}

func race() {
	// This function has a race condition.
	// "n" is being modified in two places, possibly at the same time.
	fmt.Println("No synchronization:")
	wait := make(chan struct{})
	n := 0
	go func() {
		// modified here!
		n++
		close(wait)
	}()
	// modified here!
	n++
	<-wait
	fmt.Println(n)
}

func channels() {
	//
	fmt.Println("Using channels:")
	ch := make(chan int)
	go func() {
		n := 0
		// Increment the value
		n++
		// Then pass it to the other goroutine
		ch <- n
	}()
	// Receive the value
	n := <-ch
	// Then modify it
	n++
	fmt.Println(n)
}
