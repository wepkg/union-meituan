package union

import "context"

//Endpoint ..
var Endpoint = "runion.meituan.com"

// New ..
func New(opts *Options) (*ClientUnion, error) {
	client, err := NewClient(Endpoint, opts)
	return &ClientUnion{
		Client: client,
	}, err
}

// ClientUnion ..
type ClientUnion struct {
	*Client
}

// OrderList ..
func (c *ClientUnion) OrderList(ctx context.Context) {
	uri := "/api/orderList"
	md := &requestMetadata{}
	c.newRequest(ctx, "POST", md)
}
