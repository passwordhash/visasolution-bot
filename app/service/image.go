package service

type ImageService struct {
	clientId     string
	clientSecret string
}

func (i ImageService) UploadImage(url string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func NewImageService(clientId string, clientSecret string) *ImageService {
	return &ImageService{clientId: clientId, clientSecret: clientSecret}
}
