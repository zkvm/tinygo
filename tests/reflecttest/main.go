package main

import (
	"fmt"
	"reflect"
)

type String struct {
	Data string
}

func main() {
	p := reflect.New(reflect.TypeOf(String{}))

	// NOTE: uncomment this code and it works
	// reflect.TypeOf((*String)(nil))

	v, ok := p.Interface().(*String)
	if !ok {
		fmt.Println("type assert failed")
		return
	}
	v.Data = "something"
	fmt.Println("type assert success: " + v.Data)
}
