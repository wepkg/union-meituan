package union

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"

	md5simd "github.com/minio/md5-simd"
	"github.com/wepkg/union-meituan/credentials"
)

// Client ..
type Client struct {
	endpointURL *url.URL
	// // Holds various credential providers.
	credsProvider *credentials.Credentials
	// // Custom signerType value overrides all credentials.
	overrideSignerType credentials.SignatureType

	// User supplied.
	appInfo struct {
		appName    string
		appVersion string
	}

	// Indicate whether we are using https or not
	secure bool

	// Needs allocation.
	httpClient *http.Client
	// bucketLocCache *bucketLocationCache

	// Advanced functionality.
	isTraceEnabled  bool
	traceErrorsOnly bool
	traceOutput     io.Writer

	// Random seed.
	random *rand.Rand

	// Factory for MD5 hash functions.
	md5Hasher    func() md5simd.Hasher
	sha256Hasher func() md5simd.Hasher
}

// Options for New method
type Options struct {
	Creds     *credentials.Credentials
	Secure    bool
	Transport http.RoundTripper

	// Custom hash routines. Leave nil to use standard.
	CustomMD5    func() md5simd.Hasher
	CustomSHA256 func() md5simd.Hasher
}

// Global constants.
const (
	libraryName    = "go-client"
	libraryVersion = "v1.0.0"
)

// User Agent should always following the below style.
// Please open an issue to discuss any new changes here.
//
//       AgentName (OS; ARCH) LIB/VER APP/VER
const (
	libraryUserAgentPrefix = "union-meituan (" + runtime.GOOS + "; " + runtime.GOARCH + ") "
	libraryUserAgent       = libraryUserAgentPrefix + libraryName + "/" + libraryVersion
)

// NewClient - instantiate client with options
func NewClient(endpoint string, opts *Options) (*Client, error) {
	if opts == nil {
		return nil, errors.New("no options provided")
	}
	clnt, err := privateNew(endpoint, opts)
	if err != nil {
		return nil, err
	}
	return clnt, nil
}

// EndpointURL returns the URL of the endpoint.
func (c *Client) EndpointURL() *url.URL {
	endpoint := *c.endpointURL // copy to prevent callers from modifying internal state
	return &endpoint
}

// lockedRandSource provides protected rand source, implements rand.Source interface.
type lockedRandSource struct {
	lk  sync.Mutex
	src rand.Source
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func (r *lockedRandSource) Int63() (n int64) {
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

// Seed uses the provided seed value to initialize the generator to a
// deterministic state.
func (r *lockedRandSource) Seed(seed int64) {
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}

// Redirect requests by re signing the request.
func (c *Client) redirectHeaders(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return errors.New("stopped after 5 redirects")
	}
	if len(via) == 0 {
		return nil
	}
	// lastRequest := via[len(via)-1]
	// var reAuth bool
	// for attr, val := range lastRequest.Header {
	// 	// if hosts do not match do not copy Authorization header
	// 	if attr == "Authorization" && req.Host != lastRequest.Host {
	// 		reAuth = true
	// 		continue
	// 	}
	// 	if _, ok := req.Header[attr]; !ok {
	// 		req.Header[attr] = val
	// 	}
	// }

	*c.endpointURL = *req.URL

	// value, err := c.credsProvider.Get()
	// if err != nil {
	// 	return err
	// }
	// var (
	// 	accessKeyID     = value.AccessKeyID
	// 	secretAccessKey = value.SecretAccessKey
	// 	sessionToken    = value.SessionToken
	// )

	// if reAuth {
	// 	signer.SignV4(*req, accessKeyID, secretAccessKey, sessionToken, *c.endpointURL)
	// }
	return nil
}

func privateNew(endpoint string, opts *Options) (*Client, error) {
	// construct endpoint.
	endpointURL, err := getEndpointURL(endpoint, opts.Secure)
	if err != nil {
		return nil, err
	}

	// Initialize cookies to preserve server sent cookies if any and replay
	// them upon each request.
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	// instantiate new Client.
	clnt := new(Client)

	// Save the credentials.
	clnt.credsProvider = opts.Creds

	// Remember whether we are using https or not
	clnt.secure = opts.Secure

	// Save endpoint URL, user agent for future uses.
	clnt.endpointURL = endpointURL

	transport := opts.Transport
	if transport == nil {
		transport, err = DefaultTransport(opts.Secure)
		if err != nil {
			return nil, err
		}
	}

	// Instantiate http client and bucket location cache.
	clnt.httpClient = &http.Client{
		Jar:           jar,
		Transport:     transport,
		CheckRedirect: clnt.redirectHeaders,
	}

	// Introduce a new locked random seed.
	clnt.random = rand.New(&lockedRandSource{src: rand.NewSource(time.Now().UTC().UnixNano())})

	// Add default md5 hasher.
	clnt.md5Hasher = opts.CustomMD5
	clnt.sha256Hasher = opts.CustomSHA256
	if clnt.md5Hasher == nil {
		clnt.md5Hasher = newMd5Hasher
	}
	if clnt.sha256Hasher == nil {
		clnt.sha256Hasher = newSHA256Hasher
	}
	return clnt, nil
}

// SetAppInfo - add application details to user agent.
func (c *Client) SetAppInfo(appName string, appVersion string) {
	// if app name and version not set, we do not set a new user agent.
	if appName != "" && appVersion != "" {
		c.appInfo.appName = appName
		c.appInfo.appVersion = appVersion
	}
}

// TraceOn - enable HTTP tracing.
func (c *Client) TraceOn(outputStream io.Writer) {
	// if outputStream is nil then default to os.Stdout.
	if outputStream == nil {
		outputStream = os.Stdout
	}
	// Sets a new output stream.
	c.traceOutput = outputStream

	// Enable tracing.
	c.isTraceEnabled = true
}

// TraceErrorsOnlyOn - same as TraceOn, but only errors will be traced.
func (c *Client) TraceErrorsOnlyOn(outputStream io.Writer) {
	c.TraceOn(outputStream)
	c.traceErrorsOnly = true
}

// TraceErrorsOnlyOff - Turns off the errors only tracing and everything will be traced after this call.
// If all tracing needs to be turned off, call TraceOff().
func (c *Client) TraceErrorsOnlyOff() {
	c.traceErrorsOnly = false
}

// TraceOff - disable HTTP tracing.
func (c *Client) TraceOff() {
	// Disable tracing.
	c.isTraceEnabled = false
	c.traceErrorsOnly = false
}

// Hash materials provides relevant initialized hash algo writers
// based on the expected signature type.
func (c *Client) hashMaterials(isMd5Requested bool) (hashAlgos map[string]md5simd.Hasher, hashSums map[string][]byte) {
	hashSums = make(map[string][]byte)
	hashAlgos = make(map[string]md5simd.Hasher)
	if c.secure {
		hashAlgos["md5"] = c.md5Hasher()
	} else {
		hashAlgos["sha256"] = c.sha256Hasher()
	}
	if isMd5Requested {
		hashAlgos["md5"] = c.md5Hasher()
	}
	return hashAlgos, hashSums
}

// requestMetadata - is container for all the values to make a request.
type requestMetadata struct {
	queryValues  url.Values
	customHeader http.Header
	expires      int64

	// Generated by our internal code.
	contentBody   io.Reader
	contentLength int64
}

// dumpHTTP - dump HTTP request and response.
func (c Client) dumpHTTP(req *http.Request, resp *http.Response) error {
	// Starts http dump.
	_, err := fmt.Fprintln(c.traceOutput, "---------START-HTTP---------")
	if err != nil {
		return err
	}

	// Filter out Signature field from Authorization header.
	origAuth := req.Header.Get("Authorization")
	if origAuth != "" {
		req.Header.Set("Authorization", redactSignature(origAuth))
	}

	// Only display request header.
	reqTrace, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		return err
	}

	// Write request to trace output.
	_, err = fmt.Fprint(c.traceOutput, string(reqTrace))
	if err != nil {
		return err
	}

	// Only display response header.
	var respTrace []byte

	// For errors we make sure to dump response body as well.
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusPartialContent &&
		resp.StatusCode != http.StatusNoContent {
		respTrace, err = httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}
	} else {
		respTrace, err = httputil.DumpResponse(resp, false)
		if err != nil {
			return err
		}
	}

	// Write response to trace output.
	_, err = fmt.Fprint(c.traceOutput, strings.TrimSuffix(string(respTrace), "\r\n"))
	if err != nil {
		return err
	}

	// Ends the http dump.
	_, err = fmt.Fprintln(c.traceOutput, "---------END-HTTP---------")
	if err != nil {
		return err
	}

	// Returns success.
	return nil
}

// do - execute http request.
func (c Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Handle this specifically for now until future Golang versions fix this issue properly.
		if urlErr, ok := err.(*url.Error); ok {
			if strings.Contains(urlErr.Err.Error(), "EOF") {
				return nil, &url.Error{
					Op:  urlErr.Op,
					URL: urlErr.URL,
					Err: errors.New("Connection closed by foreign host " + urlErr.URL + ". Retry again."),
				}
			}
		}
		return nil, err
	}
	// Response cannot be non-nil, report error if thats the case.
	if resp == nil {
		return nil, fmt.Errorf("Response is empty. ")
	}
	// If trace is enabled, dump http request and response,
	// except when the traceErrorsOnly enabled and the response's status code is ok
	if c.isTraceEnabled && !(c.traceErrorsOnly && resp.StatusCode == http.StatusOK) {
		err = c.dumpHTTP(req, resp)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// List of success status.
var successStatus = []int{
	http.StatusOK,
	http.StatusNoContent,
	http.StatusPartialContent,
}

// executeMethod - instantiates a given method, and retries the
// request upon any error up to maxRetries attempts in a binomially
// delayed manner using a standard back off algorithm.
func (c Client) executeMethod(ctx context.Context, method string, metadata requestMetadata) (res *http.Response, err error) {
	var retryable bool       // Indicates if request can be retried.
	var bodySeeker io.Seeker // Extracted seeker from io.Reader.
	var reqRetry = MaxRetry  // Indicates how many times we can retry the request

	if metadata.contentBody != nil {
		// Check if body is seekable then it is retryable.
		bodySeeker, retryable = metadata.contentBody.(io.Seeker)
		switch bodySeeker {
		case os.Stdin, os.Stdout, os.Stderr:
			retryable = false
		}
		// Retry only when reader is seekable
		if !retryable {
			reqRetry = 1
		}

		// Figure out if the body can be closed - if yes
		// we will definitely close it upon the function
		// return.
		bodyCloser, ok := metadata.contentBody.(io.Closer)
		if ok {
			defer bodyCloser.Close()
		}
	}

	// Create cancel context to control 'newRetryTimer' go routine.
	retryCtx, cancel := context.WithCancel(ctx)

	// Indicate to our routine to exit cleanly upon return.
	defer cancel()

	// Blank indentifier is kept here on purpose since 'range' without
	// blank identifiers is only supported since go1.4
	// https://golang.org/doc/go1.4#forrange.
	for range c.newRetryTimer(retryCtx, reqRetry, DefaultRetryUnit, DefaultRetryCap, MaxJitter) {
		// Retry executes the following function body if request has an
		// error until maxRetries have been exhausted, retry attempts are
		// performed after waiting for a given period of time in a
		// binomial fashion.
		if retryable {
			// Seek back to beginning for each attempt.
			if _, err = bodySeeker.Seek(0, 0); err != nil {
				// If seek failed, no need to retry.
				return nil, err
			}
		}

		// Instantiate a new request.
		var req *http.Request
		req, err = c.newRequest(ctx, method, metadata)
		if err != nil {
			// errResponse := ToErrorResponse(err)
			// if isS3CodeRetryable(errResponse.Code) {
			// 	continue // Retry.
			// }
			return nil, err
		}

		// Initiate the request.
		res, err = c.do(req)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			// Retry the request
			continue
		}

		// For any known successful http status, return quickly.
		for _, httpStatus := range successStatus {
			if httpStatus == res.StatusCode {
				return res, nil
			}
		}

		// Read the body to be saved later.
		errBodyBytes, err := ioutil.ReadAll(res.Body)
		// res.Body should be closed
		// closeResponse(res)
		res.Body.Close()
		if err != nil {
			return nil, err
		}

		// Save the body.
		errBodySeeker := bytes.NewReader(errBodyBytes)
		res.Body = ioutil.NopCloser(errBodySeeker)

		// Save the body back again.
		errBodySeeker.Seek(0, 0) // Seek back to starting point.
		res.Body = ioutil.NopCloser(errBodySeeker)

		// Verify if http status code is retryable.
		if isHTTPStatusRetryable(res.StatusCode) {
			continue // Retry.
		}

		// For all other cases break out of the retry loop.
		break
	}

	// Return an error when retry is canceled or deadlined
	if e := retryCtx.Err(); e != nil {
		return nil, e
	}

	return res, err
}

// newRequest - instantiate a new HTTP request for a given method.
func (c Client) newRequest(ctx context.Context, method string, metadata requestMetadata) (req *http.Request, err error) {
	// If no method is supplied default to 'POST'.
	if method == "" {
		method = http.MethodPost
	}
	targetURL, err := c.makeTargetURL(metadata.queryValues)
	if err != nil {
		return nil, err
	}

	// Initialize a new HTTP request for the method.
	req, err = http.NewRequestWithContext(ctx, method, targetURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Get credentials from the configured credentials provider.
	// value, err := c.credsProvider.Get()
	// if err != nil {
	// 	return nil, err
	// }

	// var (
	// 	signerType      = value.SignerType
	// 	accessKeyID     = value.AccessKeyID
	// 	secretAccessKey = value.SecretAccessKey
	// 	sessionToken    = value.SessionToken
	// )

	// Set 'User-Agent' header for the request.
	c.setUserAgent(req)

	// Set all headers.
	for k, v := range metadata.customHeader {
		req.Header.Set(k, v[0])
	}

	// Go net/http notoriously closes the request body.
	// - The request Body, if non-nil, will be closed by the underlying Transport, even on errors.
	// This can cause underlying *os.File seekers to fail, avoid that
	// by making sure to wrap the closer as a nop.
	if metadata.contentLength == 0 {
		req.Body = nil
	} else {
		req.Body = ioutil.NopCloser(metadata.contentBody)
	}

	// Set incoming content-length.
	req.ContentLength = metadata.contentLength
	if req.ContentLength <= -1 {
		// For unknown content length, we upload using transfer-encoding: chunked.
		req.TransferEncoding = []string{"chunked"}
	}
	// For anonymous requests just return.
	// if signerType.IsAnonymous() {
	// 	return req, nil
	// }
	// Add signature authorization header.
	// req = signer.SignV4(*req, accessKeyID, secretAccessKey, sessionToken)
	// Return request.
	return req, nil
}

// set User agent.
func (c Client) setUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", libraryUserAgent)
	if c.appInfo.appName != "" && c.appInfo.appVersion != "" {
		req.Header.Set("User-Agent", libraryUserAgent+" "+c.appInfo.appName+"/"+c.appInfo.appVersion)
	}
}

// makeTargetURL make a new target url.
func (c Client) makeTargetURL(queryValues url.Values) (*url.URL, error) {
	host := c.endpointURL.Host
	// Save scheme.
	scheme := c.endpointURL.Scheme

	// Strip port 80 and 443 so we won't send these ports in Host header.
	// The reason is that browsers and curl automatically remove :80 and :443
	// with the generated presigned urls, then a signature mismatch error.
	if h, p, err := net.SplitHostPort(host); err == nil {
		if scheme == "http" && p == "80" || scheme == "https" && p == "443" {
			host = h
		}
	}

	urlStr := scheme + "://" + host + "/"

	// If there are any query values, add them to the end.
	if len(queryValues) > 0 {
		urlStr = urlStr + "?" + queryEncode(queryValues)
	}

	return url.Parse(urlStr)
}
