package union

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"hash"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	md5simd "github.com/minio/md5-simd"
	"github.com/minio/sha256-simd"
)

func trimEtag(etag string) string {
	etag = strings.TrimPrefix(etag, "\"")
	return strings.TrimSuffix(etag, "\"")
}

// xmlDecoder provide decoded value in xml.
func xmlDecoder(body io.Reader, v interface{}) error {
	d := xml.NewDecoder(body)
	return d.Decode(v)
}

// sum256 calculate sha256sum for an input byte array, returns hex encoded.
func sum256Hex(data []byte) string {
	hash := newSHA256Hasher()
	defer hash.Close()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// sumMD5Base64 calculate md5sum for an input byte array, returns base64 encoded.
func sumMD5Base64(data []byte) string {
	hash := newMd5Hasher()
	defer hash.Close()
	hash.Write(data)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// getEndpointURL - construct a new endpoint.
func getEndpointURL(endpoint string, secure bool) (*url.URL, error) {
	if strings.Contains(endpoint, ":") {
		host, _, err := net.SplitHostPort(endpoint)
		if err != nil {
			return nil, err
		}
	}
	// If secure is false, use 'http' scheme.
	scheme := "https"
	if !secure {
		scheme = "http"
	}

	// Construct a secured endpoint URL.
	endpointURLStr := scheme + "://" + endpoint
	endpointURL, err := url.Parse(endpointURLStr)
	if err != nil {
		return nil, err
	}

	// Validate incoming endpoint URL.
	if err := isValidEndpointURL(*endpointURL); err != nil {
		return nil, err
	}
	return endpointURL, nil
}

// closeResponse close non nil response with any response Body.
// convenient wrapper to drain any remaining data on response body.
//
// Subsequently this allows golang http RoundTripper
// to re-use the same connection for future requests.
func closeResponse(resp *http.Response) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if resp != nil && resp.Body != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		//  - http://stackoverflow.com/a/17961593/4465767
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
}

var (
	// Hex encoded string of nil sha256sum bytes.
	emptySHA256Hex = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// Sentinel URL is the default url value which is invalid.
	sentinelURL = url.URL{}
)

// Verify if input endpoint URL is valid.
func isValidEndpointURL(endpointURL url.URL) error {
	if endpointURL == sentinelURL {
		return errInvalidArgument("Endpoint url cannot be empty.")
	}
	if endpointURL.Path != "/" && endpointURL.Path != "" {
		return errInvalidArgument("Endpoint url cannot have fully qualified paths.")
	}
	return nil
}

// Verify if input expires value is valid.
func isValidExpiry(expires time.Duration) error {
	expireSeconds := int64(expires / time.Second)
	if expireSeconds < 1 {
		return errInvalidArgument("Expires cannot be lesser than 1 second.")
	}
	if expireSeconds > 604800 {
		return errInvalidArgument("Expires cannot be greater than 7 days.")
	}
	return nil
}

// Extract only necessary metadata header key/values by
// filtering them out with a list of custom header keys.
func extractObjMetadata(header http.Header) http.Header {
	preserveKeys := []string{
		"Content-Type",
		"Cache-Control",
		"Content-Encoding",
		"Content-Language",
		"Content-Disposition",
		// Add new headers to be preserved.
		// if you add new headers here, please extend
		// PutObjectOptions{} to preserve them
		// upon upload as well.
	}
	filteredHeader := make(http.Header)
	for k, v := range header {
		var found bool
		for _, prefix := range preserveKeys {
			if !strings.HasPrefix(k, prefix) {
				continue
			}
			found = true
			break
		}
		if found {
			filteredHeader[k] = v
		}
	}
	return filteredHeader
}

var readFull = func(r io.Reader, buf []byte) (n int, err error) {
	// ReadFull reads exactly len(buf) bytes from r into buf.
	// It returns the number of bytes copied and an error if
	// fewer bytes were read. The error is EOF only if no bytes
	// were read. If an EOF happens after reading some but not
	// all the bytes, ReadFull returns ErrUnexpectedEOF.
	// On return, n == len(buf) if and only if err == nil.
	// If r returns an error having read at least len(buf) bytes,
	// the error is dropped.
	for n < len(buf) && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		// Some spurious io.Reader's return
		// io.ErrUnexpectedEOF when nn == 0
		// this behavior is undocumented
		// so we are on purpose not using io.ReadFull
		// implementation because this can lead
		// to custom handling, to avoid that
		// we simply modify the original io.ReadFull
		// implementation to avoid this issue.
		// io.ErrUnexpectedEOF with nn == 0 really
		// means that io.EOF
		if err == io.ErrUnexpectedEOF && nn == 0 {
			err = io.EOF
		}
		n += nn
	}
	if n >= len(buf) {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}

// regCred matches credential string in HTTP header
var regCred = regexp.MustCompile("Credential=([A-Z0-9]+)/")

// regCred matches signature string in HTTP header
var regSign = regexp.MustCompile("Signature=([[0-9a-f]+)")

// Redact out signature value from authorization string.
func redactSignature(origAuth string) string {
	return "Auth **REDACTED**:**REDACTED**"
	// // Strip out accessKeyID from:
	// // Credential=<access-key-id>/<date>/<aws-region>/<aws-service>/aws4_request
	// newAuth := regCred.ReplaceAllString(origAuth, "Credential=**REDACTED**/")

	// // Strip out 256-bit signature from: Signature=<256-bit signature>
	// return regSign.ReplaceAllString(newAuth, "Signature=**REDACTED**")
}

var supportedHeaders = []string{
	"content-type",
	"cache-control",
	"content-encoding",
	"content-disposition",
	"content-language",
	"expires",
	// Add more supported headers here.
}

// isStandardHeader returns true if header is a supported header and not a custom header
func isStandardHeader(headerKey string) bool {
	key := strings.ToLower(headerKey)
	for _, header := range supportedHeaders {
		if strings.ToLower(header) == key {
			return true
		}
	}
	return false
}

var md5Pool = sync.Pool{New: func() interface{} { return md5.New() }}
var sha256Pool = sync.Pool{New: func() interface{} { return sha256.New() }}

func newMd5Hasher() md5simd.Hasher {
	return hashWrapper{Hash: md5Pool.New().(hash.Hash), isMD5: true}
}

func newSHA256Hasher() md5simd.Hasher {
	return hashWrapper{Hash: sha256Pool.New().(hash.Hash), isSHA256: true}
}

// hashWrapper implements the md5simd.Hasher interface.
type hashWrapper struct {
	hash.Hash
	isMD5    bool
	isSHA256 bool
}

// Close will put the hasher back into the pool.
func (m hashWrapper) Close() {
	if m.isMD5 && m.Hash != nil {
		m.Reset()
		md5Pool.Put(m.Hash)
	}
	if m.isSHA256 && m.Hash != nil {
		m.Reset()
		sha256Pool.Put(m.Hash)
	}
	m.Hash = nil
}

// Expects ascii encoded strings - from output of urlEncodePath
func percentEncodeSlash(s string) string {
	return strings.Replace(s, "/", "%2F", -1)
}

// QueryEncode - encodes query values in their URL encoded form. In
// addition to the percent encoding performed by urlEncodePath() used
// here, it also percent encodes '/' (forward slash)
func queryEncode(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := percentEncodeSlash(EncodePath(k)) + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(percentEncodeSlash(EncodePath(v)))
		}
	}
	return buf.String()
}

// if object matches reserved string, no need to encode them
var reservedObjectNames = regexp.MustCompile("^[a-zA-Z0-9-_.~/]+$")

// EncodePath encode the strings from UTF-8 byte representations to HTML hex escape sequences
//
// This is necessary since regular url.Parse() and url.Encode() functions do not support UTF-8
// non english characters cannot be parsed due to the nature in which url.Encode() is written
//
// This function on the other hand is a direct replacement for url.Encode() technique to support
// pretty much every UTF-8 character.
func EncodePath(pathName string) string {
	if reservedObjectNames.MatchString(pathName) {
		return pathName
	}
	var encodedPathname strings.Builder
	for _, s := range pathName {
		if 'A' <= s && s <= 'Z' || 'a' <= s && s <= 'z' || '0' <= s && s <= '9' { // §2.3 Unreserved characters (mark)
			encodedPathname.WriteRune(s)
			continue
		}
		switch s {
		case '-', '_', '.', '~', '/': // §2.3 Unreserved characters (mark)
			encodedPathname.WriteRune(s)
			continue
		default:
			len := utf8.RuneLen(s)
			if len < 0 {
				// if utf8 cannot convert return the same string as is
				return pathName
			}
			u := make([]byte, len)
			utf8.EncodeRune(u, s)
			for _, r := range u {
				hex := hex.EncodeToString([]byte{r})
				encodedPathname.WriteString("%" + strings.ToUpper(hex))
			}
		}
	}
	return encodedPathname.String()
}
