package main

import (
	"fmt"
)

//import fmt "fmt" // Package implementing formatted I/O.

type node struct {
	ip string
	port int
	networkID int
}

var k int = 20
var alpha int =3

func main() {
	var routingTable list[map[int]node]
	 var bucket map[int]node
	 bucket = make(map[int]node)
	 bucket[158] =node{"10.0.0.0",80, 158}
	fmt.Println(bucket[158].ip)
}