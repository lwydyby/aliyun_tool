package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"aliyun/config"
	"aliyun/utils"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var base = "https://openapi.alipan.com"
var rateLimiter = rate.NewLimiter(4, 1)

type Json map[string]interface{}

type Files struct {
	Items      []File `json:"items"`
	NextMarker string `json:"next_marker"`
}

type File struct {
	DriveId       string    `json:"drive_id"`
	FileId        string    `json:"file_id"`
	ParentFileId  string    `json:"parent_file_id"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	FileExtension string    `json:"file_extension"`
	ContentHash   string    `json:"content_hash"`
	Category      string    `json:"category"`
	Type          string    `json:"type"`
	Thumbnail     string    `json:"thumbnail"`
	Url           string    `json:"url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// create only
	FileName string `json:"file_name"`
}

func list(ctx context.Context, data Json) (*Files, error) {
	err := rateLimiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}
	var resp Files
	_, err = request("/adrive/v1.0/openFile/list", http.MethodPost, func(req *resty.Request) {
		req.SetBody(data).SetResult(&resp)
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetFiles(ctx context.Context, fileId string, driverType string) ([]File, error) {
	marker := "first"
	res := make([]File, 0)
	for marker != "" {
		if marker == "first" {
			marker = ""
		}
		data := Json{
			"drive_id":        getDriveID(driverType),
			"limit":           200,
			"marker":          marker,
			"order_by":        "created_at",
			"order_direction": "ASC",
			"parent_file_id":  fileId,
			// "category":              "",
			// "type":                  "",
			// "video_thumbnail_time":  120000,
			// "video_thumbnail_width": 480,
			// "image_thumbnail_width": 480,
		}
		resp, err := list(ctx, data)
		if err != nil {
			return nil, err
		}
		marker = resp.NextMarker
		res = append(res, resp.Items...)
	}
	return res, nil
}

func Rename(name string, f File) error {
	err := rateLimiter.Wait(context.Background())
	if err != nil {
		return err
	}
	var newFile File
	_, err = request("/adrive/v1.0/openFile/update", http.MethodPost, func(req *resty.Request) {
		req.SetBody(Json{
			"drive_id": f.DriveId,
			"file_id":  f.FileId,
			"name":     name,
		}).SetResult(&newFile)
	})
	return err
}

func getDriveID(driveType string) string {
	if config.C().DriveType == driveType && config.C().DriveID != "" {
		return config.C().DriveID
	}
	err := rateLimiter.Wait(context.Background())
	if err != nil {
		return ""
	}
	res, err := request("/adrive/v1.0/user/getDriveInfo", http.MethodPost, nil)
	if err != nil {
		panic(err)
	}
	config.C().DriveID = utils.Json.Get(res, config.C().DriveType+"_drive_id").ToString()
	config.C().DriveType = driveType
	config.SaveYaml()
	return config.C().DriveID
}

type ReqCallback func(req *resty.Request)

func request(uri, method string, callback ReqCallback, retry ...bool) ([]byte, error) {
	b, err, _ := requestReturnErrResp(uri, method, callback, retry...)
	return b, err
}

func requestReturnErrResp(uri, method string, callback ReqCallback, retry ...bool) ([]byte, error, *ErrResp) {
	req := NewRestyClient().R()
	// TODO check whether access_token is expired
	req.SetHeader("Authorization", "Bearer "+config.C().Access)
	if method == http.MethodPost {
		req.SetHeader("Content-Type", "application/json")
	}
	if callback != nil {
		callback(req)
	}
	var e ErrResp
	req.SetError(&e)
	res, err := req.Execute(method, base+uri)
	if err != nil {
		if res != nil {
			log.Errorf("[aliyundrive_open] request error: %s", res.String())
		}
		return nil, err, nil
	}
	isRetry := len(retry) > 0 && retry[0]
	if e.Code != "" {
		if !isRetry && (SliceContains([]string{"AccessTokenInvalid", "AccessTokenExpired", "I400JD"}, e.Code) || config.C().Access == "") {
			config.C().Token, config.C().Access, err = RefreshToken(config.C().Token)
			if err != nil {
				return nil, err, &e
			}
			config.SaveYaml()
			return requestReturnErrResp(uri, method, callback, true)
		}
		return nil, fmt.Errorf("%s:%s", e.Code, e.Message), &e
	}
	return res.Body(), nil, nil
}

// SliceContains check if slice contains element
func SliceContains[T comparable](arr []T, v T) bool {
	for _, vv := range arr {
		if vv == v {
			return true
		}
	}
	return false
}
