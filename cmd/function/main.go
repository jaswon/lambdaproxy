package function

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Address string `json:"address"`
	Port    string `json:"string"`
	Key     string `json:"key"`
	User    string `json:"user"`
}

func HandleRequest(req Request) error {

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
