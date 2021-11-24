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
	"fmt"
	"strings"

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
	User                string
	Password            string
	AuthorizationLambda string
	JWT                 *JWT
}

func NewProviderCreator(jwtSecret string, legacy Legacy) ProviderCreator {
	return func(user, password string) (Provider, error) {
		switch {
		case legacy.AuthorizationLambda != "":
			userJWT, err := CheckAuthorization(user, password, legacy.AuthorizationLambda, legacy.JWT)
			if err != nil {
				return nil, err
			}
			return newAWSS3Provider(*userJWT)
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
			return nil, fmt.Errorf("Credentials do not match user/password nor are a jwt")
		}
	}
}
