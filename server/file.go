package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"aliyun/api"
	"aliyun/config"

	json "github.com/json-iterator/go"
)

type TokenRequest struct {
	Token string `json:"token"`
}

func SubmitTokenHandler(w http.ResponseWriter, req *http.Request) { // 读取请求体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// 将请求体解析到TokenRequest结构体中
	var tokenRequest TokenRequest
	err = json.Unmarshal(body, &tokenRequest)
	if err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if len(tokenRequest.Token) == 0 {
		http.Error(w, "token is empty", http.StatusInternalServerError)
		return
	}
	config.C().Token, config.C().Access, err = api.RefreshToken(tokenRequest.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = config.SaveYaml()
	if err != nil {
		http.Error(w, "save token to yaml failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type BatchFileReqeust struct {
	FilePath string `json:"file_path,omitempty"`
	Name     string `json:"name,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
}

type GetFileNameResponse struct {
	Name string
}

var (
	cacheOnce sync.Once
	cache     map[string][]api.File
	cacheLock sync.Mutex
)

func GetFileHandler(w http.ResponseWriter, req *http.Request) {
	cacheOnce.Do(func() {
		cache = map[string][]api.File{}
	})
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	var request BatchFileReqeust
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if request.FilePath == "" {
		http.Error(w, "file path is empty", http.StatusBadRequest)
		return
	}
	deviceType, filedId := getDeviceTypeAndFileID(request.FilePath)
	files, err := api.GetFiles(context.Background(), filedId, deviceType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// 缓存files,减少查询次数
	cacheLock.Lock()
	defer cacheLock.Unlock()
	cache[filedId] = files
	resp := GetFileNameResponse{}
	if len(files) > 0 {
		resp.Name = files[0].Name
	}
	responseData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error generating JSON response", http.StatusInternalServerError)
		return
	}

	// 设置响应头为JSON格式
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseData)
	if err != nil {
		fmt.Println(err)
	}
}

func BatchRenameHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	var request BatchFileReqeust
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if request.FilePath == "" {
		http.Error(w, "file path is empty", http.StatusBadRequest)
		return
	}
	deviceType, filedId := getDeviceTypeAndFileID(request.FilePath)
	var files []api.File
	var ok bool
	cacheLock.Lock()
	defer cacheLock.Unlock()
	files, ok = cache[filedId]
	if !ok {
		files, err = api.GetFiles(context.Background(), filedId, deviceType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		delete(cache, filedId)
	}
	for i := range files {
		file := files[i]
		newFileName, err := generateNewFileName(file.Name, request.Prefix, request.Name)
		if err != nil {
			continue
		}
		err = api.Rename(newFileName, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type GetDirRequest struct {
	Dir        string `json:"dir,omitempty"`
	DeviceType string `json:"device_type,omitempty"`
}

type GetDirResponse struct {
	Url string `json:"url"`
}

func GetDirHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	var request GetDirRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	if request.Dir == "" || request.DeviceType == "" {
		http.Error(w, "dir is empty or device_type is empty", http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	dirs := strings.Split(request.Dir, "/")
	fileId := "root"
	for i := range dirs {
		files, err := api.GetFiles(ctx, fileId, request.DeviceType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		hasFound := false
		for j := range files {
			if files[i].Type == "file" {
				continue
			}
			if files[j].Name == dirs[i] {
				fileId = files[j].FileId
				hasFound = true
				break
			}
		}
		if !hasFound {
			http.Error(w, "dir not found", http.StatusBadRequest)
			return
		}
	}
	resp := GetDirResponse{
		Url: fmt.Sprintf("%s/%s", request.DeviceType, fileId),
	}
	responseData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error generating JSON response", http.StatusInternalServerError)
		return
	}

	// 设置响应头为JSON格式
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseData)
	if err != nil {
		fmt.Println(err)
	}
}

func getDeviceTypeAndFileID(path string) (string, string) {
	filePaths := strings.Split(path, "/")
	return filePaths[len(filePaths)-2], filePaths[len(filePaths)-1]
}
