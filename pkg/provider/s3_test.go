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
	"io"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type mockClientListObjectsV2 struct {
	s3iface.S3API
	res *s3.ListObjectsV2Output
	err error
}

func (mock mockClientListObjectsV2) ListObjectsV2(input *s3.ListObjectsV2Input) (output *s3.ListObjectsV2Output, err error) {
	return mock.res, mock.err
}

type mockDownloaderDownload struct {
	s3manageriface.DownloaderAPI
	bytes []byte
	size  int64
	err   error
}

func (mock mockDownloaderDownload) Download(writer io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (size int64, err error) {
	writer.WriteAt(mock.bytes, 0)
	return mock.size, mock.err
}
