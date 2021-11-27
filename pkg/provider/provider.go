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
	"errors"
	"strings"

	"github.com/golang-jwt/jwt"
)

type ProviderCreator func(user, password string) (Provider, error)

type Provider interface {
	ListEmails(notNumbers []int) (emails map[int]*Email, err error)
	GetEmail(number int, notNumbers []int) (email *Email, err error)
	GetEmailPayload(number int, notNumbers []int) (payload EmailPayload, err error)
	DeleteEmail(number int) (err error)
}

type S3Bucket struct {
	AWSAccessKeyID     string `json:"awsAccessKeyID,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	Region             string `json:"region,omitempty"`
	Bucket             string `json:"bucket,omitempty"`
	Prefix             string `json:"prefix,omitempty"`
}

type JWTClaims struct {
	jwt.StandardClaims
	Provider string `json:"provider,omitempty"`
	S3Bucket
}

type StaticCredentials struct {
	User     string
	Password string
	S3Bucket *S3Bucket
}

func NewStaticCredentialsProviderCreator(staticCreds StaticCredentials) ProviderCreator {
	return func(user, password string) (Provider, error) {
		if user == staticCreds.User && password == staticCreds.Password {
			if staticCreds.S3Bucket != nil {
				return newS3Provider(*staticCreds.S3Bucket)
			}
			return newNoneProvider()
		}
		return nil, errors.New("credentials do not match user/password")
	}
}

func NewLambdaAuthorizationProviderCreator(authorizationLambda string, s3Bucket S3Bucket) ProviderCreator {
	return func(user, password string) (Provider, error) {
		authorizedBucket, err := CheckAuthorization(user, password, authorizationLambda, &s3Bucket)
		if err != nil {
			return nil, err
		}

		return newS3Provider(*authorizedBucket)
	}
}

func NewJWTProviderCreator(jwtSecret string) ProviderCreator {
	return func(user, password string) (Provider, error) {
		if strings.EqualFold(user, "jwt") {
			claims := JWTClaims{}
			if _, err := jwt.ParseWithClaims(password, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			}); err != nil {
				return nil, err
			}
			switch true {
			case strings.EqualFold(claims.Provider, "none"):
				return newNoneProvider()
			case strings.EqualFold(claims.Provider, "demo"):
				return newNoneProvider(DemoEmail)
			case claims.Provider == "" || strings.EqualFold(claims.Provider, "s3"):
				return newS3Provider(S3Bucket{
					AWSAccessKeyID:     claims.AWSAccessKeyID,
					AWSSecretAccessKey: claims.AWSSecretAccessKey,
					Region:             claims.Region,
					Bucket:             claims.Bucket,
					Prefix:             claims.Prefix,
				})
			}
			return nil, errors.New("provider must be either be '', 'none', 'demo' or 's3'")
		}
		return nil, errors.New("user is != 'jwt'")
	}
}
