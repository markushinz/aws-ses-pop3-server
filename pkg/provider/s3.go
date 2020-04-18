/*
   Copyright 2020 Markus Hinz

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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type awsS3Cache struct {
	emails map[int]*email
}

type awsS3Provider struct {
	bucket     string
	prefix     string
	client     s3iface.S3API
	downloader s3manageriface.DownloaderAPI
	cache      *awsS3Cache
}

func NewAWSS3ProviderCreator(region, bucket, prefix string) func() (Provider, error) {
	return func() (Provider, error) {
		return newAWSS3Provider(region, bucket, prefix)
	}
}

func newAWSS3Provider(region, bucket, prefix string) (provider *awsS3Provider, err error) {
	client, downloader, err := initClientAndDownloader(region)
	if err != nil {
		return nil, err
	}
	return &awsS3Provider{
		bucket:     bucket,
		prefix:     prefix,
		client:     client,
		downloader: downloader,
	}, nil
}

func initClientAndDownloader(region string) (client *s3.S3, downloader *s3manager.Downloader, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil, nil, err
	}
	return s3.New(sess), s3manager.NewDownloader(sess), nil
}

func (provider *awsS3Provider) ListEmails(notNumbers []int) (emails map[int]*email, err error) {
	if provider.cache == nil {
		err := provider.initCache()
		if err != nil {
			return nil, err
		}
	}
	emails = make(map[int]*email)
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

func (provider *awsS3Provider) initCache() (err error) {
	res, err := provider.client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(provider.bucket)})
	if err != nil {
		return err
	}
	provider.cache = &awsS3Cache{
		emails: make(map[int]*email),
	}
	for index, item := range res.Contents {
		provider.cache.emails[index+1] = &email{
			ID:   strings.TrimPrefix(*item.Key, provider.prefix),
			Size: *item.Size,
		}
	}
	return nil
}

func (provider *awsS3Provider) GetEmail(number int, notNumbers []int) (email *email, err error) {
	if provider.cache == nil {
		err := provider.initCache()
		if err != nil {
			return nil, err
		}
	}
	if email, exists := provider.cache.emails[number]; exists {
		return email, nil
	}
	return nil, fmt.Errorf("%v does not exist", number)
}

func (provider *awsS3Provider) GetEmailPayload(number int, notNumbers []int) (payload emailPayload, err error) {
	email, err := provider.GetEmail(number, notNumbers)
	if err != nil {
		return nil, err
	}
	if email.payloadOptional == nil {
		buf := aws.NewWriteAtBuffer([]byte{})
		_, err = provider.downloader.Download(buf, &s3.GetObjectInput{
			Bucket: aws.String(provider.bucket),
			Key:    aws.String(provider.prefix + email.ID),
		})
		if err != nil {
			return nil, err
		}
		var payload emailPayload
		payload = buf.Bytes()
		email.payloadOptional = &payload
	}
	return *email.payloadOptional, nil
}

func (provider *awsS3Provider) DeleteEmail(number int, notNumbers []int) (err error) {
	email, err := provider.GetEmail(number, notNumbers)
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
