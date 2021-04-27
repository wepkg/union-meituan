package types

import (
	"time"
)

// OrderType ..
type OrderType string

// GroupOrder 团购订单
const GroupOrder OrderType = "0"

// HotelOrder 酒店订单
const HotelOrder OrderType = "2"

// TakeawayOrder 外卖订单
const TakeawayOrder OrderType = "4"

// PhoneOrder 话费订单
const PhoneOrder OrderType = "5"

// FlashOrder 闪购订单
const FlashOrder OrderType = "6"

// Order ..
type Order struct {
	OrderID  string `json:"orderid"`  // 订单id
	SmsTitle string `json:"smstitle"` // 订单标题

	PayTime  string `json:"paytime"`  // 订单支付时间，10位时间戳
	PayPrice string `json:"payprice"` // 订单用户实际支付金额
	Profit   string `json:"profit"`   // 订单预估返佣金额

	Status       int    `json:"status"`       // 订单状态，外卖、话费、闪购订单会返回该字段 1 已付款 8 已完成 9 已退款或风控 否
	Appkey       string `json:"appkey"`       // 订单对应的appkey，外卖、话费、闪购订单会返回该字段	否
	RefundPrice  string `json:"refundprice"`  // 订单实际退款金额，外卖、话费、闪购订单若发生退款会返回该字段	否
	RefundTime   string `json:"refundtime"`   // 订单退款时间，10位时间戳，外卖、话费、闪购订单若发生退款会返回该字段	否
	RefundProfit string `json:"refundprofit"` // 订单需要扣除的返佣金额，外卖、话费、闪购订单若发生退款会返回该字段	否

	Uid string `json:"uid"` // 联盟媒体id
	Sid string `json:"sid"` // 订单对应的推广位sid

	Total  string `json:"total"`  // 订单总金额
	Direct string `json:"direct"` // 订单用户实付金额

	Quantity int       `json:"quantity"` // 订单包含的团购券数量(团购订单该字段有意义，其它订单不用考虑该字段)
	DealID   string    `json:"dealid"`
	ModTime  time.Time `json:"modtime"`
}

// Coupon ..
type Coupon struct {
	Sequence string    `json:"sequence"` // 核销序列号
	OrderID  string    `json:"orderid"`  // 订单id
	Price    string    `json:"price"`    // 核销实际支付金额
	Profit   string    `json:"profit"`   // 核销实际返佣金额
	UseTime  time.Time `json:"usetime"`  // 核销时间
}

// Refund ..
type Refund struct {
	OrderID    string    `json:"orderid"`         // 订单id
	Quantity   int       `json:"quantity,string"` // 退款笔数
	RefundTime time.Time `json:"refundtime"`      // 退款时间，10位时间戳
	Money      string    `json:"money"`           // 实际退款金额
}

// OrderListReq ..
type OrderListReq struct {
	Type          string `json:""`          //查询订单类型 0团购订单 2酒店订单 4外卖订单 5话费订单 6闪购订单
	StartTime     string `json:"startTime"` //查询起始时间10位时间戳，以下单时间为准
	EndTime       string `json:"endTime"`   //查询起始时间10位时间戳，以下单时间为准
	Page          string `json:"page"`      //分页参数，起始值从1开始
	Limit         string `json:"limit"`     //每页显示数据条数，最大值为100
	QueryTimeType string `json:""`          //查询时间类型，枚举值 1 按订单支付时间查询 2 按订单发生修改时间查询
}

// OrderListResp ..
type OrderListResp struct {
	DataList []Order `json:"dataList"`
	Total    int64   `json:"total"`
}

// OrderReq ..
type OrderReq struct {
	Oid  string // 订单id 单个订单查询必传
	Type string // 查询订单类型 0 团购订单 2 酒店订单 4 外卖订单 5 话费订单 6 闪购订单 是
}

// OrderResp ..
type OrderResp struct {
	Order  Order    `json:"order"`
	Coupon []Coupon `json:"coupon"` // 	否，根据full值决定是否回传
	Refund Refund   `json:"refund"` // 否，根据full值决定是否回传
}

// RtNotifyReq ..
type RtNotifyReq struct {
	Oid  string // 订单id 单个订单查询必传
	Sid  string // 推广位sid 否
	Full bool   // 是否返回完整订单信息(即是否包含返佣、退款信息)	否，需要完整信息full=1
	Type string // 查询订单类型 0 团购订单 2 酒店订单 4 外卖订单 5 话费订单 6 闪购订单 是
}

// RtNotifyResp ..
type RtNotifyResp OrderResp

// CouponListReq ..
type CouponListReq struct {
	Type      string `json:"type"`      //查询订单类型 0团购订单 2酒店订单 4外卖订单 5话费订单 6闪购订单
	StartTime string `json:"startTime"` //查询起始时间10位时间戳，以下单时间为准
	EndTime   string `json:"endTime"`   //查询起始时间10位时间戳，以下单时间为准
	Page      string `json:"page"`      //分页参数，起始值从1开始
	Limit     string `json:"limit"`     //每页显示数据条数，最大值为100
	Sid       string `json:"sid"`       //推广位id
}

// CouponListResp ..
type CouponListResp struct {
	DataList []CouponItem `json:"dataList"`
	Total    int64        `json:"total"`
}

// CouponItem ..
type CouponItem struct {
	Appkey      string    `json:"appKey"`      // 媒体appkey
	Sid         string    `json:"sid"`         // 推广位sid
	CouponTime  time.Time `json:"couponTime"`  // 领券日期yyyy-MM-dd HH:mm:ss
	Money       string    `json:"money"`       // 券优惠金额	是
	MinUseMoney string    `json:"minUseMoney"` // 用券门槛金额	是
	CouponName  string    `json:"couponName"`  // 券名称	是
	CouponType  string    `json:"couponType"`  // 券类型	是
	CouponCode  string    `json:"couponCode"`  // 券唯一标识	是
	BeginTime   string    `json:"beginTime"`   // 	券生效起始时间，10位时间戳	是
	EndTime     string    `json:"endTime"`     // 券生效截止时间，10位时间戳	是
}

// CallbackOrder ..
type CallbackOrder struct {
	OrderID  string `json:"orderid"`  // 订单id
	SmsTitle string `json:"smstitle"` // 订单标题
	DealID   string `json:"dealid"`   // 订单所含商品id	是

	PayTime string `json:"paytime"` // 订单支付时间，10位时间戳
	// PayPrice string `json:"payprice"` // 订单用户实际支付金额
	// Profit   string `json:"profit"`   // 订单预估返佣金额

	Status int `json:"status"` // 订单状态，外卖、话费、闪购订单会返回该字段 1 已付款 8 已完成 9 已退款或风控 否
	// Appkey       string `json:"appkey"`       // 订单对应的appkey，外卖、话费、闪购订单会返回该字段	否
	// RefundPrice  string `json:"refundprice"`  // 订单实际退款金额，外卖、话费、闪购订单若发生退款会返回该字段	否
	// RefundTime   string `json:"refundtime"`   // 订单退款时间，10位时间戳，外卖、话费、闪购订单若发生退款会返回该字段	否
	// RefundProfit string `json:"refundprofit"` // 订单需要扣除的返佣金额，外卖、话费、闪购订单若发生退款会返回该字段	否

	Uid string `json:"uid"` // 联盟媒体id
	Sid string `json:"sid"` // 订单对应的推广位sid

	Total  string `json:"total"`  // 订单总金额
	Direct string `json:"direct"` // 订单用户实付金额

	Quantity int `json:"quantity"` // 订单数量	是

	// ModTime  time.Time `json:"modtime"`

	Type      string    `json:""` //查询订单类型 0团购订单 2酒店订单 4外卖订单 5话费订单 6闪购订单
	OrderTime time.Time `json:"ordertime"`
	Ratio     string    `json:"ratio"` //订单返佣比例，cps活动的订单会返回该字段	是
	Sign      string    `json:"sign"`
}

type GenerateLinkResp struct {
	Status int    `json:"status"`
	Des    string `json:"des"`
	Data   string `json:"data"`
}
