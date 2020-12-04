package main

import (
	"context"
	"log"

	"github.com/wepkg/union-meituan"
)

func main() {
	client, err := union.New("www.baidu.com", &union.Options{
		// Creds:  credentials.NewStatic("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
		Secure: true,
	})
	if err != nil {
		log.Fatalln(err)
	}
	client.Test(context.Background())
}
