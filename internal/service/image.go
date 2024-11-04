package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	cfg "visasolution/internal/config"
	pkgService "visasolution/pkg/service"
)

type ImgurResponse struct {
	Data    ImgurData `json:"data"`
	Success bool      `json:"success"`
	Status  int       `json:"status"`
}

type ImgurData struct {
	Link string `json:"link"`
}

type ImageService struct {
	clientId     string
	clientSecret string
	client       *http.Client
}

func NewImageService(clientId string, clientSecret string) *ImageService {
	return &ImageService{clientId: clientId, clientSecret: clientSecret}
}

func (i *ImageService) ClientInitWithProxy(proxy cfg.Proxy) error {
	transport, err := pkgService.ProxyTransport(proxy.URL())
	if err != nil {
		return fmt.Errorf("new transport with proxy error: %v", err)
	}

	i.client = &http.Client{Transport: transport}

	return nil
}

func (i *ImageService) UploadImage(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	part, err := writer.CreateFormFile("image", imagePath)
	if err != nil {
		return "", fmt.Errorf("create form file error: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("copy file error: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", "https://api.imgur.com/3/image", &reqBody)
	if err != nil {
		return "", fmt.Errorf("create request error: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", i.clientId))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := i.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("client do error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code error: %d", resp.StatusCode)
	}

	var imgurResp ImgurResponse
	err = json.NewDecoder(resp.Body).Decode(&imgurResp)
	if err != nil {
		return "", fmt.Errorf("decode response error: %v", err)
	}

	if !imgurResp.Success {
		return "", fmt.Errorf("imgur response error: %d", imgurResp.Status)
	}

	return imgurResp.Data.Link, nil
}
