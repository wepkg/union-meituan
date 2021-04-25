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

	client, err := union.New(&union.Auth{
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
		Type:          "4",
		StartTime:     strconv.FormatInt(st.Unix(), 10),
		EndTime:       strconv.FormatInt(et.Unix(), 10),
		Page:          "1",
		Limit:         "100",
		QueryTimeType: "1",
	}
	resp, err := client.GetOrderList(context.TODO(), params)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", resp.DataList[0])

	// fmt.Println(resp.DataList[0], err)

	p2 := &types.RtNotifyReq{
		Oid:  "4374293237780870",
		Sid:  "test",
		Full: true,
		Type: "4",
	}
	var r2 *types.RtNotifyResp
	err = client.RtNotify(context.TODO(), p2, r2)
	fmt.Println(r2, err)

	p3 := &types.CouponListReq{
		Type:      "4",
		StartTime: "",
		EndTime:   "",
		Page:      "1",
		Limit:     "1",
		Sid:       "test",
	}
	r3, err := client.GetCouponList(context.TODO(), p3)
	fmt.Println(r3, err)

	p4 := &types.GenerateLinkReq{
		ActID:    2,
		Sid:      "union_skd_test",
		LinkType: types.LinkTypeWxa,
	}
	r4, err := client.GenerateLink(context.TODO(), p4)
	fmt.Println(r4, err)

	// http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
	// 	body, _ := ioutil.ReadAll(r.Body)
	// 	fmt.Println(string(body))
	// })
	// http.ListenAndServe(":8080", nil)
}
