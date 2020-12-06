package union

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// APIEndpoint constants
const (
	version         = "v0.0.1"
	APIEndpointBase = "https://runion.meituan.com"
)

// Auth ..
type Auth interface{}

// Client type
type Client struct {
	auth         Auth
	signer       func(auth Auth, vals url.Values) string
	endpointBase *url.URL     // default APIEndpointBase
	httpClient   *http.Client // default http.DefaultClient
	retryKeyID   string       // X-Retry-Key allows you to safely retry API requests without duplicating messages
}

// ClientOption type
type ClientOption func(*Client) error

// New returns a new bot client instance.
func New(auth Auth, options ...ClientOption) (*Client, error) {
	c := &Client{
		auth:       auth,
		httpClient: http.DefaultClient,
	}
	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, err
		}
	}
	if c.endpointBase == nil {
		u, err := url.ParseRequestURI(APIEndpointBase)
		if err != nil {
			return nil, err
		}
		c.endpointBase = u
	}
	c.signer = makeSign
	return c, nil
}

// WithHTTPClient function
func WithHTTPClient(c *http.Client) ClientOption {
	return func(client *Client) error {
		client.httpClient = c
		return nil
	}
}

// WithEndpointBase function
func WithEndpointBase(endpointBase string) ClientOption {
	return func(client *Client) error {
		u, err := url.ParseRequestURI(endpointBase)
		if err != nil {
			return err
		}
		client.endpointBase = u
		return nil
	}
}

func (client *Client) url(base *url.URL, endpoint string) string {
	u := *base
	u.Path = path.Join(u.Path, endpoint)
	return u.String()
}

func (client *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// req.Header.Set("Authorization", "Bearer "+client.channelToken)
	req.Header.Set("User-Agent", "Go-Client/"+version)
	if len(client.retryKeyID) > 0 {
		req.Header.Set("X-Retry-Key", client.retryKeyID)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	return client.httpClient.Do(req)

}

func (client *Client) get(ctx context.Context, base *url.URL, endpoint string, query url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, client.url(base, endpoint), nil)
	if err != nil {
		return nil, err
	}
	if query != nil {
		// signer
		if client.signer != nil {
			query.Add("sign", client.signer(client.auth, query))
		}
		req.URL.RawQuery = query.Encode()
	}
	return client.do(ctx, req)
}

func (client *Client) post(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, client.url(client.endpointBase, endpoint), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return client.do(ctx, req)
}

func (client *Client) postform(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", client.url(client.endpointBase, endpoint), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return client.do(ctx, req)
}

func (client *Client) put(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, client.url(client.endpointBase, endpoint), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return client.do(ctx, req)
}

func (client *Client) delete(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, client.url(client.endpointBase, endpoint), nil)
	if err != nil {
		return nil, err
	}
	return client.do(ctx, req)
}

func (client *Client) setRetryKey(retryKey string) {
	client.retryKeyID = retryKey
}

func closeResponse(res *http.Response) error {
	defer res.Body.Close()
	_, err := io.Copy(ioutil.Discard, res.Body)
	return err
}

func makeSign(auth Auth, vals url.Values) string {
	var str string
	for k, v := range vals {
		str += k + v[0]
	}
	encoder := md5.New()
	encoder.Write([]byte(str))
	encodeBytes := encoder.Sum(nil)
	return hex.EncodeToString(encodeBytes)
}
