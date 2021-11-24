package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type Authorization struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func CheckAuthorization(user, password, authorizationLambda string, jwtInfo *JWT) (*JWT, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(jwtInfo.Region),
		Credentials: credentials.NewStaticCredentials(jwtInfo.AWSAccessKeyID, jwtInfo.AWSSecretAccessKey, ""),
	})
	if err != nil {
		return nil, err
	}
	svc := lambda.New(sess)
	payload, _ := json.Marshal(Authorization{
		Name:     user,
		Password: password,
	})

	input := &lambda.InvokeInput{
		FunctionName: aws.String(authorizationLambda),
		Payload:      payload,
	}

	result, err := svc.Invoke(input)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	} else {
		response := string(result.Payload)
		if response == "\"OK\"" {
			origJSON, err := json.Marshal(&jwtInfo)
			if err != nil {
				return nil, err
			}

			userJWT := JWT{}
			if err = json.Unmarshal(origJSON, &userJWT); err != nil {
				return nil, err
			}
			prefix := userJWT.Prefix
			if prefix != "" && !strings.HasSuffix(prefix, "/") {
				userJWT.Prefix += "/"
			}
			userJWT.Prefix += user
			return &userJWT, err
		} else {
			return nil, fmt.Errorf("Credentials are not authorized")
		}
	}
}
