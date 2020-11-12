package main

import "fmt"

type Receiver struct {

}

func (r *Receiver) Receive(line []byte) {



	fmt.Println(string(line))
}