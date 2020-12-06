package union

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError type
type APIError struct {
	Code     int
	Response *ErrorResponse
}

// Error method
func (e *APIError) Error() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "sdk: APIError %d ", e.Code)
	if e.Response != nil {
		fmt.Fprintf(&buf, "%s", e.Response.Message)
		for _, d := range e.Response.Details {
			fmt.Fprintf(&buf, "\n[%s] %s", d.Property, d.Message)
		}
	}
	return buf.String()
}

// BasicResponse type
type BasicResponse struct {
	RequestID string
}
type errorResponseDetail struct {
	Message  string `json:"message"`
	Property string `json:"property"`
}

// ErrorResponse type
type ErrorResponse struct {
	Message string                `json:"message"`
	Details []errorResponseDetail `json:"details"`
	// OAuth Errors
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// isSuccess checks if status code is 2xx: The action was successfully received,
// understood, and accepted.
func isSuccess(code int) bool {
	return code/100 == 2
}
func checkResponse(resp *http.Response) error {
	if isSuccess(resp.StatusCode) {
		return nil
	}
	decoder := json.NewDecoder(resp.Body)
	result := ErrorResponse{}
	if err := decoder.Decode(&result); err != nil {
		return &APIError{
			Code: resp.StatusCode,
		}
	}
	return &APIError{
		Code:     resp.StatusCode,
		Response: &result,
	}
}

func decodeToBasicResponse(res *http.Response) (*BasicResponse, error) {
	if err := checkResponse(res); err != nil {
		return nil, err
	}
	fmt.Println(res.Body)
	decoder := json.NewDecoder(res.Body)
	result := BasicResponse{
		RequestID: res.Header.Get("X-Request-Id"),
	}
	if err := decoder.Decode(&result); err != nil {
		if err == io.EOF {
			return &result, nil
		}
		return nil, err
	}
	return &result, nil
}
