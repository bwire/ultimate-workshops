package main

import (
	"fmt"
	"time"
)

func main() {
	safeGo(func() {
		panic("big panic")
	})
	fmt.Println("Main is still alive")
	time.Sleep(1 * time.Second)
}

func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic caught!")
			}
		}()
		fn()
	}()
}
