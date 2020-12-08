package union

import (
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

func checkResponse(resp *http.Response) error {
	if resp.StatusCode/100 == 2 { //2xx
		return nil
	}
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
