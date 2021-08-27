package union

import (
	"context"
	"net/url"
	"strconv"

	"github.com/wepkg/union-meituan/types"
)

const (
	// APIOrderList 订单列表查询(新)
	APIOrderList = "/api/orderList"
	// APIRtNotify 订单查询接口(旧) 单个订单/订单列表
	APIRtNotify = "/api/rtnotify"
	// APICouponList
	APICouponList = "/api/couponList"
	//APIGenerateLink
	APIGenerateLink = "/generateLink"
	//MeituanWxaAppid 美团小程序appid
	MeituanWxaAppid = "wxde8ac0a21135c07d"
)

// GetMeituanWxaAppid 获取美团微信小程序appid
func (c *Client) GetMeituanWxaAppid() string {
	return MeituanWxaAppid
}

// GetOrderList 订单列表
func (c *Client) GetOrderList(ctx context.Context, in *types.OrderListReq) (*types.OrderListResp, error) {
	query := url.Values{}
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

// GetOrder 获取单个订单
func (c *Client) GetOrder(ctx context.Context, in *types.OrderReq) (*types.OrderResp, error) {
	query := url.Values{}
	query.Add("oid", in.Oid) //订单id	是
	query.Add("type", in.Type)
	query.Add("full", "1")
	resp, err := c.get(ctx, c.endpointBase, APIRtNotify, query)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)
	out := &types.OrderResp{}
	return out, decodeToResp(resp, out)
}

// RtNotify 订单/订单列表 () 由于接口返回参数不固定需要不建议直接使用
func (c *Client) RtNotify(ctx context.Context, in *types.RtNotifyReq, out interface{}) error {
	query := url.Values{}
	query.Add("oid", in.Oid) //订单id	是
	if in.Full {
		query.Add("full", "1")
	} else {
		query.Add("full", "0")
	}
	query.Add("sid", in.Sid)
	query.Add("type", in.Type)
	resp, err := c.get(ctx, c.endpointBase, APIRtNotify, query)
	if err != nil {
		return err
	}
	defer closeResponse(resp)
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))
	return decodeToResp(resp, out)
}

// GetCouponList 领券结果查询
func (c *Client) GetCouponList(ctx context.Context, in *types.CouponListReq) (*types.CouponListResp, error) {
	query := url.Values{}
	query.Add("type", in.Type)
	query.Add("startTime", in.StartTime)
	query.Add("endTime", in.EndTime)
	query.Add("page", in.Page)
	query.Add("limit", in.Limit)
	query.Add("sid", in.Sid)
	resp, err := c.get(ctx, c.endpointBase, APICouponList, query)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))
	out := &types.CouponListResp{}
	return out, decodeToResp(resp, out)
}

const LinkTypeH5 int = 1       //h5链接
const LinkTypeDeepLink int = 2 //deeplink(唤起)链接
const LinkTypeJump int = 3     //中间页唤起链接
const LinkTypeWxa int = 4      //微信小程序唤起路径

// GenerateLink 生成连接
func (c *Client) GenerateLink(ctx context.Context, linkType, actID int, sid string) (string, error) {
	query := url.Values{}
	query.Add("actId", strconv.Itoa(actID))
	query.Add("sid", sid)
	query.Add("linkType", strconv.Itoa(linkType))
	resp, err := c.get(ctx, c.endpointBase, APIGenerateLink, query)
	if err != nil {
		return "", err
	}
	defer closeResponse(resp)
	type generateLinkResp struct {
		Status int    `json:"status"`
		Des    string `json:"des"`
		Data   string `json:"data"`
	}
	out := &generateLinkResp{}
	if err := decodeToResp(resp, out); err != nil {
		return "", err
	}
	if out.Status != 0 {
		return "", APIError{Errno: 400, Errmsg: out.Des}
	}
	return out.Data, nil
}

// CallbackOrder 订单回推接口
func (c *Client) CallbackOrder(ctx context.Context) (*types.CallbackOrder, error) {
	// 数据正常，返回: {"errcode":"0","errmsg":"ok"}
	// 数据错误，返回: {"errcode":"1","errmsg":"err"}
	out := &types.CallbackOrder{}
	// content, err := ioutil.ReadAll(resp.Body)
	// decoder := json.NewDecoder(resp.Body)
	// if err := decoder.Decode(result); err != nil {
	// 	return err
	// }
	return out, nil
}
