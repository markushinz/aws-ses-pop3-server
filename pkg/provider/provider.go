/*
   Copyright 2021 Markus Hinz

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/spf13/viper"
	"gopkg.in/square/go-jose.v2/jwt"
)

type ProviderCreator func(user, password string) (Provider, error)

type Provider interface {
	ListEmails(notNumbers []int) (emails map[int]*Email, err error)
	GetEmail(number int, notNumbers []int) (email *Email, err error)
	GetEmailPayload(number int, notNumbers []int) (payload EmailPayload, err error)
	DeleteEmail(number int) (err error)
}

type JWT struct {
	AWSAccessKeyID     string `json:"awsAccessKeyID,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	Region             string `json:"region,omitempty"`
	Bucket             string `json:"bucket,omitempty"`
	Prefix             string `json:"prefix,omitempty"`
}

type Legacy struct {
	User     string
	Password string
	JWT      *JWT
}

func NewProviderCreator(jwtSecret string, legacy Legacy) ProviderCreator {
	return func(user, password string) (Provider, error) {
		switch {
		case user == legacy.User && password == legacy.Password:
			if legacy.JWT != nil {
				return newAWSS3Provider(*legacy.JWT)
			}
			return newNoneProvider()
		case strings.EqualFold(user, "jwt"):
			token, err := jwt.ParseSigned(password)
			if err != nil {
				return nil, err
			}
			decoded := JWT{}
			if err := token.Claims([]byte(jwtSecret), &decoded); err != nil {
				return nil, err
			}
			return newAWSS3Provider(decoded)
		default:
			if viper.IsSet("authorization-lambda") && viper.IsSet("aws-s3-region") {
				ok := CheckAuthorization(user, password, viper.GetString("authorization-lambda"), viper.GetString("aws-s3-region"))
				if ok {
					if legacy.JWT != nil {
						return newAWSS3Provider(*legacy.JWT)
					}
					return newNoneProvider()
				}
			}
			return nil, fmt.Errorf("Credentials do not match user/password nor are a jwt")
		}
	}
}

type Authorization struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func CheckAuthorization(user, password, authorizationLambdaName, region string) bool {
	svc := lambda.New(session.New(), aws.NewConfig().WithRegion(region))
	payload, _ := json.Marshal(Authorization{
		Name:     user,
		Password: password,
	})

	input := &lambda.InvokeInput{
		FunctionName: aws.String(authorizationLambdaName),
		Payload:      payload,
	}

	result, err := svc.Invoke(input)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		response := string(result.Payload)
		authenticated := response == "\"OK\""
		return authenticated
	}

	return false
}
