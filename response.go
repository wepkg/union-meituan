package union

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// APIError type
type APIError struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("APIError: %v [%v]", e.Errmsg, e.Errno)
}

// decodeToOrderListResp ..
func decodeToResp(resp *http.Response, result interface{}) error {
	if resp.StatusCode/100 != 2 { //2xx
		// decoder := json.NewDecoder(resp.Body)
		// if err := decoder.Decode(&result); err != nil {
		// 	return &APIError{
		// 		Code: resp.StatusCode,
		// 	}
		// }
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return &APIError{
			Errno:  resp.StatusCode,
			Errmsg: string(content),
		}
	}
	if resp.ContentLength == 0 {
		return &APIError{
			Errno:  resp.StatusCode,
			Errmsg: "Content Empty",
		}
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(result); err != nil {
		return err
	}
	return nil
}
