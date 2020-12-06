package main

import (
	"context"
	"fmt"
	"log"

	"github.com/wepkg/union-meituan"
	"github.com/wepkg/union-meituan/types"
)

func main() {
	client, err := union.New(nil)
	if err != nil {
		log.Fatalln(err)
	}
	params := types.OrderListReq{
		Ts:   "123123",
		Type: "1",
		
	}
	ret, err := client.GetOrderList(context.TODO(), params)
	fmt.Println(ret, err)
}
