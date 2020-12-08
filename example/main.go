package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/wepkg/union-meituan"
	"github.com/wepkg/union-meituan/types"
)

func main() {
	client, err := union.New(&union.TokenAuth{
		Token: "7405cd574aa31c2ddc6bad5113697a93",
	})
	if err != nil {
		log.Fatalln(err)
	}
	timeLayout := "2006-01-02 15:04:05"
	st, _ := time.Parse(timeLayout, "2020-11-01 00:00:00")
	et, _ := time.Parse(timeLayout, "2020-12-31 00:00:00")
	fmt.Println(time.Now().Format(timeLayout))
	params := types.OrderListReq{
		Ts:            strconv.FormatInt(time.Now().Unix(), 10),
		Type:          "4",
		StartTime:     strconv.FormatInt(st.Unix(), 10),
		EndTime:       strconv.FormatInt(et.Unix(), 10),
		Page:          "1",
		Limit:         "100",
		QueryTimeType: "1",
	}
	r1, err := client.GetOrderList(context.TODO(), params)
	fmt.Println(r1, err)

	r2, err := client.RtNotify(context.TODO())
	fmt.Println(r2, err)
}
