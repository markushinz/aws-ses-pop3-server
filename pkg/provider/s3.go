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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type s3Cache struct {
	emails map[int]*Email
}

type s3Provider struct {
	bucket     string
	prefix     string
	client     s3iface.S3API
	downloader s3manageriface.DownloaderAPI
	cache      *s3Cache
}

var _ Provider = &s3Provider{}

func newS3Provider(jwt S3Bucket) (provider *s3Provider, err error) {
	client, downloader, err := initClientAndDownloader(jwt.AWSAccessKeyID, jwt.AWSSecretAccessKey, jwt.Region)
	if err != nil {
		return nil, err
	}
	prefix := jwt.Prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return &s3Provider{
		bucket:     jwt.Bucket,
		prefix:     prefix,
		client:     client,
		downloader: downloader,
	}, nil
}

func initClientAndDownloader(awsAccessKeyID, awsSecretAccessKey, region string) (client *s3.S3, downloader *s3manager.Downloader, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
	})
	if err != nil {
		return nil, nil, err
	}
	return s3.New(sess), s3manager.NewDownloader(sess), nil
}

func (provider *s3Provider) initCache() (err error) {
	res, err := provider.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(provider.bucket),
		Prefix: aws.String(provider.prefix),
	})
	if err != nil {
		return err
	}
	provider.cache = &s3Cache{
		emails: make(map[int]*Email),
	}
	for index, item := range res.Contents {
		provider.cache.emails[index+1] = &Email{
			ID:   strings.TrimPrefix(*item.Key, provider.prefix),
			Size: *item.Size,
		}
	}
	return nil
}

func (provider *s3Provider) ListEmails(notNumbers []int) (emails map[int]*Email, err error) {
	if provider.cache == nil {
		err := provider.initCache()
		if err != nil {
			return nil, err
		}
	}
	emails = make(map[int]*Email)
	for number, email := range provider.cache.emails {
		skip := false
		for _, notNumnber := range notNumbers {
			if number == notNumnber {
				skip = true
				break
			}
		}
		if !skip {
			emails[number] = email
		}
	}
	return emails, nil
}

func (provider *s3Provider) GetEmail(number int, notNumbers []int) (email *Email, err error) {
	emails, err := provider.ListEmails(notNumbers)
	if err != nil {
		return nil, err
	}
	if email, exists := emails[number]; exists {
		return email, nil
	}
	return nil, fmt.Errorf("%v does not exist", number)
}

func (provider *s3Provider) GetEmailPayload(number int, notNumbers []int) (payload EmailPayload, err error) {
	email, err := provider.GetEmail(number, notNumbers)
	if err != nil {
		return nil, err
	}
	if email.Payload == nil {
		buf := aws.NewWriteAtBuffer([]byte{})
		_, err = provider.downloader.Download(buf, &s3.GetObjectInput{
			Bucket: aws.String(provider.bucket),
			Key:    aws.String(provider.prefix + email.ID),
		})
		if err != nil {
			return nil, err
		}
		var payload EmailPayload
		payload = buf.Bytes()
		email.Payload = &payload
	}
	return *email.Payload, nil
}

func (provider *s3Provider) DeleteEmail(number int) (err error) {
	email, err := provider.GetEmail(number, nil)
	if err != nil {
		return err
	}
	_, err = provider.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(provider.bucket),
		Key:    aws.String(provider.prefix + email.ID),
	})
	if err != nil {
		return err
	}
	err = provider.client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(provider.bucket),
		Key:    aws.String(provider.prefix + email.ID),
	})
	return err
}
