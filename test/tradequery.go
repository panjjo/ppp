package main

import (
	"fmt"
	"log"
	"net/rpc"

	"gopkg.in/mgo.v2/bson"

	"github.com/panjjo/ppp"
)

func main() {
	client, err := rpc.DialHTTP("tcp", ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	result := ppp.TradeListResult{}
	err = client.Call("Statement.List", ppp.ListRequest{
		Query: bson.M{},
		Sort:  "uptime",
	}, &result)
	fmt.Println(result)
	if err != nil {
		fmt.Println(err)
	}
}
