package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/wepkg/union-meituan"
	"github.com/wepkg/union-meituan/types"
)

func main() {
	Appkey := flag.String("appkey", "", "Appkey")
	Secret := flag.String("secret", "", "Secret")
	flag.Parse()

	client, err := union.New(union.Auth{
		Appkey: *Appkey,
		Secret: *Secret,
	})
	if err != nil {
		log.Fatalln(err)
	}
	timeLayout := "2006-01-02 15:04:05"
	st, _ := time.Parse(timeLayout, "2020-11-01 00:00:00")
	et, _ := time.Parse(timeLayout, "2020-12-31 00:00:00")
	params := &types.OrderListReq{
		Ts:            strconv.FormatInt(time.Now().Unix(), 10),
		Type:          "4",
		StartTime:     strconv.FormatInt(st.Unix(), 10),
		EndTime:       strconv.FormatInt(et.Unix(), 10),
		Page:          "1",
		Limit:         "100",
		QueryTimeType: "1",
	}
	resp, err := client.GetOrderList(context.TODO(), params)
	fmt.Println(err)
	fmt.Println(resp)
	// fmt.Println(resp.DataList[0], err)

	// r2, err := client.RtNotify(context.TODO())
	// fmt.Println(r2, err)
}
