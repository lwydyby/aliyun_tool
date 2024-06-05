package api

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"aliyun/utils"

	"github.com/go-resty/resty/v2"
)

var UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"
var DefaultTimeout = time.Second * 30
var OauthTokenURL = "https://api.nn.ci/alist/ali_open/token"
var GetTokenURL = "https://alist.nn.ci/tool/aliyundrive/request.html"

func NewRestyClient() *resty.Client {
	client := resty.New().
		SetHeader("user-agent", UserAgent).
		SetRetryCount(3).
		SetRetryResetReaders(true).
		SetTimeout(DefaultTimeout).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	return client
}

type ErrResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RefreshToken(token string) (string, string, error) {
	var e ErrResp
	res, err := NewRestyClient().R().
		//ForceContentType("application/json").
		SetBody(map[string]interface{}{
			"client_id":     "",
			"client_secret": "",
			"grant_type":    "refresh_token",
			"refresh_token": token,
		}).
		// SetResult(&resp).
		SetError(&e).
		Post(OauthTokenURL)
	if err != nil {
		return "", "", err
	}
	if e.Code != "" {
		return "", "", fmt.Errorf("failed to refresh token: %s", e.Message)
	}
	refresh, access := utils.Json.Get(res.Body(), "refresh_token").ToString(), utils.Json.Get(res.Body(), "access_token").ToString()
	if refresh == "" {
		return "", "", fmt.Errorf("failed to refresh token: refresh token is empty, resp: %s", res.String())
	}
	curSub, err := getSub(token)
	if err != nil {
		return "", "", err
	}
	newSub, err := getSub(refresh)
	if err != nil {
		return "", "", err
	}
	if curSub != newSub {
		return "", "", errors.New("failed to refresh token: sub not match")
	}
	return refresh, access, nil
}

func getSub(token string) (string, error) {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		return "", errors.New("not a jwt token because of invalid segments")
	}
	bs, err := base64.RawStdEncoding.DecodeString(segments[1])
	if err != nil {
		return "", errors.New("failed to decode jwt token")
	}
	return utils.Json.Get(bs, "sub").ToString(), nil
}
