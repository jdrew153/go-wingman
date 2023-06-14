package services

import (

	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type CFImageUploaderService struct {
	Client *http.Client
}

func NewCFImageUploaderService(http *http.Client) *CFImageUploaderService {
	
	return &CFImageUploaderService{
		Client: http,
	}
}

func (s *CFImageUploaderService) CreateImageUploadURL() (string, error) {
	godotenv.Load("../cmd/.env")

	CFTACCID := os.Getenv("CF_ACC_IT")
	CFTOKEN := os.Getenv("CF_API_TOKEN")

	req, err := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/accounts/"+CFTACCID+"/images/v2/direct_upload", nil)

	if err != nil {
		println(err.Error())
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+ CFTOKEN)
	req.Header.Add("Content-Type", "application/json")

	res, err := s.Client.Do(req)

	fmt.Println(res)

	if err != nil {
		println(err.Error())
		return "", err
	}
	

	body, err :=  ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
	
	if err != nil {
		return "", err
	}

	return string(body), nil
}