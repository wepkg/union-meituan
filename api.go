package union

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wepkg/union-meituan/types"
)

const (
	// APIOrderList 订单列表查询(新)
	APIOrderList = "/api/orderList"
	// APIRtNotify 单个订单查询接口(旧)
	APIRtNotify = "/api/rtnotify"
)

// decodeToOrderListResp ..
func decodeToResp(resp *http.Response, result interface{}) error {
	if err := checkResponse(resp); err != nil {
		return err
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(result); err != nil {
		return err
	}
	return nil
}

// GetOrderList 订单列表
func (c *Client) GetOrderList(ctx context.Context, in *types.OrderListReq) (*types.OrderListResp, error) {
	query := url.Values{}
	query.Add("ts", in.Ts)
	query.Add("type", in.Type)
	query.Add("startTime", in.StartTime)
	query.Add("endTime", in.EndTime)
	query.Add("page", in.Page)
	query.Add("limit", in.Limit)
	query.Add("queryTimeType", in.QueryTimeType)
	resp, err := c.get(ctx, c.endpointBase, APIOrderList, query)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)
	out := &types.OrderListResp{}
	return out, decodeToResp(resp, out)
}

// RtNotify 订单列表
func (c *Client) RtNotify(ctx context.Context) (*types.RtNotifyResp, error) {
	query := url.Values{}
	query.Add("oid", "1") //订单id	是
	query.Add("full", "1")
	query.Add("type", "4")
	resp, err := c.get(ctx, c.endpointBase, APIRtNotify, query)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)
	fmt.Println(resp)
	result := &types.RtNotifyResp{}
	return result, decodeToResp(resp, result)
}

// // RtNotify 订单列表
// func (c *Client) RtNotify(ctx context.Context) (*types.OrderListResp, error) {
// 	query := url.Values{}
// 	query.Add("oid", "1") //订单id	是
// 	query.Add("full", "1")
// 	query.Add("type", "4")
// 	resp, err := c.get(ctx, c.endpointBase, APIRtNotify, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer closeResponse(resp)
// 	fmt.Println(resp)
// 	return decodeToOrderListResp(resp)
// }

// // RtNotify 订单列表
// func (c *Client) RtNotify(ctx context.Context) (*types.OrderListResp, error) {
// 	query := url.Values{}
// 	query.Add("oid", "1") //订单id	是
// 	query.Add("full", "1")
// 	query.Add("type", "4")
// 	resp, err := c.get(ctx, c.endpointBase, APIRtNotify, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer closeResponse(resp)
// 	fmt.Println(resp)
// 	return decodeToOrderListResp(resp)
// }
