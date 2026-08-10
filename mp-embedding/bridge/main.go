// main.go
package main

/*
#include <stdio.h>
*/

import "C"
import "fmt"

//export say_hello
func say_hello(name string) {
	fmt.Println("Hello from Go!", name)
}

func main() {}
