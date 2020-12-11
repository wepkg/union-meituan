package union

import (
	"fmt"
	"net/url"
)

// NewActivity origURL 为美团活动原始url
func NewActivity(origURL string) *Activity {
	return &Activity{OriginalURL: origURL}
}

// Activity ...
type Activity struct {
	OriginalURL string
}

// BuildH5URL ..
func (a Activity) BuildH5URL(appkey, sid string) string {
	jumpBase := "https://runion.meituan.com/url?key=%v&url=%v&sid=%v"
	origURL, _ := url.Parse(a.OriginalURL)
	origURL.RawQuery = origURL.RawQuery + fmt.Sprintf("appkey=%v:%v", appkey, sid)
	strURL := origURL.String()
	escapeURL := url.QueryEscape(strURL)
	jumpURL := fmt.Sprintf(jumpBase, appkey, escapeURL, sid)
	return jumpURL
}

// BuildDeepLink ..
func (a Activity) BuildDeepLink(appkey, sid string) string {
	jumpBase := "imeituan://www.meituan.com?web?url=%v&lch=%v"
	strURL := a.BuildH5URL(appkey, sid)
	escapeURL := url.QueryEscape(strURL)
	lch := fmt.Sprintf("cps:waimai:%v:%v:%v", 3, appkey, sid)
	jumpURL := fmt.Sprintf(jumpBase, escapeURL, lch)
	return jumpURL
}

// BuildH5Evoke ..
func (a Activity) BuildH5Evoke(appkey, sid string) string {
	jumpBase := "https://w.dianping.com/cube/evoke/increase/meituan.html?lch=%v&url=%v"
	strURL := a.BuildH5URL(appkey, sid)
	escapeURL := url.QueryEscape(strURL)
	lch := fmt.Sprintf("cps:waimai:%v:%v:%v", 1, appkey, sid)
	jumpURL := fmt.Sprintf(jumpBase, lch, escapeURL)
	return jumpURL
}

// BuildWechatMiniapp ..
func (a Activity) BuildWechatMiniapp(appkey, sid string, needLogin bool) string {
	jumpBase := "/index/pages/h5/h5?weburl=%v"
	strURL := a.BuildH5URL(appkey, sid)
	escapeURL := url.QueryEscape(strURL)
	jumpURL := fmt.Sprintf(jumpBase, escapeURL)
	if needLogin == false {
		jumpURL = jumpURL + "&f_token=1&f_userId=1"
	}
	return jumpURL
}
