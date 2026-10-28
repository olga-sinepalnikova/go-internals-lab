package main

import "fmt"

type User struct{ Name string }

// Case 1: returning a pointer
func newUser() *User {
	u := User{Name: "Bob"}
	return &u
}

// Case 2: wrapping a value in an interface
func printIt(n int) {
	fmt.Println(n)
}

// Case 3: closure that outlives the function
func counter() func() int {
	c := 0
	return func() int { c++; return c }
}

// Case 4: pointer sent into a channel
func send(ch chan *User) {
	u := User{Name: "Alice"}
	ch <- &u
}

// Case 5: very big slice
func big() int {
	s := make([]int, 100000)
	return len(s)
}

func newUserValue() User {
	u := User{Name: "Bob"}
	return u
}

func sendValue(ch chan User) {
	u := User{Name: "Alice"}
	ch <- u
}

func main() {
	_ = newUser()
	printIt(1)
	_ = counter()
	send(make(chan *User, 1))
	_ = big()

	_ = newUserValue()
	sendValue(make(chan User))
}
