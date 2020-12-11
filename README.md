# union-meituan

## 美团联盟接口

- 订单列表查询(新)
- 单个订单查询接口(旧)
- 订单列表查询(旧)
- 领券结果查询
- 订单回推接口

- 构造分享链接

### 开始：
引入包
```golang
import "github.com/wepkg/union-meituan"
```
初始化
```golang
    client, err := union.New(&union.Auth{
		Appkey: *Appkey,
		Secret: *Secret,
	})
	if err != nil {
		log.Fatalln(err)
	}
```

### 查询订单列表
```golang
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
fmt.Printf("%+v", resp.DataList)
```
### 查询单个订单

```golang
params := &types.OrderReq{
    Oid: "1000000001",
    Type:          "4",
}
resp, err := client.GetOrder(context.TODO(), params)
if err != nil {
    fmt.Println(err)
}
fmt.Printf("%+v", resp)
```

### 领券结果查询

```golang
params := &types.OrderReq{
    Oid: "1000000001",
    Type:          "4",
}
resp, err := client.GetOrder(context.TODO(), params)
if err != nil {
    fmt.Println(err)
}
fmt.Printf("%+v", resp)
```

### 构造分享链接

```golang
act := union.NewActivity("<原始活动链接>")
// 构造h5分享页链接
shareH5Url := act.BuildH5URL("<appkey>","<sid>")
//构造deeplink链接
shareDkUrl := act.BuildDeepLink("<appkey>","<sid>")
//构造h5唤起app链接
shareH5eUrl := act.BuildH5Evoke("<appkey>","<sid>")
//构造小程序分享链接
shareWxappUrl := act.BuildWechatMiniapp("<appkey>","<sid>")

```
