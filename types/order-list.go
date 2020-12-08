package types

// OrderListReq ..
type OrderListReq struct {
	Ts            string `json:"ts"`        //请求时刻10位时间戳(秒级)，有效期60s	是
	Type          string `json:""`          //查询订单类型 0团购订单 2酒店订单 4外卖订单 5话费订单 6闪购订单
	StartTime     string `json:"startTime"` //查询起始时间10位时间戳，以下单时间为准
	EndTime       string `json:"endTime"`   //查询起始时间10位时间戳，以下单时间为准
	Page          string `json:"page"`      //分页参数，起始值从1开始
	Limit         string `json:"limit"`     //每页显示数据条数，最大值为100
	QueryTimeType string `json:""`          //查询时间类型，枚举值 1 按订单支付时间查询 2 按订单发生修改时间查询
}

// OrderListResp ..
type OrderListResp struct {
	OrderList
	Total int64
}

// OrderList ..
type OrderList struct {
	OrderId      string `json:"orderid"`      // 订单id	是
	PayTime      string `json:"paytime"`      // 订单支付时间，10位时间戳	是
	PayPrice     string `json:"payprice"`     // 订单用户实际支付金额	是
	Profit       string `json:"profit"`       // 订单预估返佣金额	是
	Sid          string `json:"sid"`          // 订单对应的推广位sid	是
	Appkey       string `json:"appkey"`       // 订单对应的appkey，外卖、话费、闪购订单会返回该字段	否
	SmsTitle     string `json:"smstitle"`     // 订单标题	是
	RefundPrice  string `json:"refundprice"`  // 订单实际退款金额，外卖、话费、闪购订单若发生退款会返回该字段	否
	RefundTime   string `json:"refundtime"`   // 订单退款时间，10位时间戳，外卖、话费、闪购订单若发生退款会返回该字段	否
	RefundProfit string `json:"refundprofit"` // 订单需要扣除的返佣金额，外卖、话费、闪购订单若发生退款会返回该字段	否
	Status       int64  `json:"status"`       // 订单状态，外卖、话费、闪购订单会返回该字段 1 已付款 8 已完成 9 已退款或风控 否
}

// RtNotifyResp ..
type RtNotifyResp struct {
}
