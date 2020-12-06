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
	APIOrderList = "/api/orderList"
)

// GetOrderList 订单列表
func (c *Client) GetOrderList(ctx context.Context, params types.OrderListReq) (*types.OrderListResp, error) {
	query := url.Values{}
	query.Add("ts", params.Ts)
	query.Add("type", params.Type)
	query.Add("startTime", params.StartTime)
	query.Add("endTime", params.EndTime)
	query.Add("page", params.Page)
	query.Add("limit", params.Limit)
	query.Add("queryTimeType", params.QueryTimeType)
	resp, err := c.get(ctx, c.endpointBase, APIOrderList, query)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)
	fmt.Println(resp)
	return decodeToOrderListResp(resp)
}

// decodeToOrderListResp ..
func decodeToOrderListResp(resp *http.Response) (*types.OrderListResp, error) {
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	result := &types.OrderListResp{}
	if err := decoder.Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}
