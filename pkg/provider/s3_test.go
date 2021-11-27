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
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/stretchr/testify/assert"
)

type mockItem struct {
	key   string
	size  int64
	bytes []byte
}

type mockClient struct {
	s3iface.S3API
	items     []mockItem
	listErr   error
	deleteErr error
}

var _ s3iface.S3API = &mockClient{}

func (mock *mockClient) ListObjectsV2(input *s3.ListObjectsV2Input) (output *s3.ListObjectsV2Output, err error) {
	var contents []*s3.Object
	for _, item := range mock.items {
		key := item.key
		size := item.size
		contents = append(contents, &s3.Object{
			Key:  &key,
			Size: &size,
		})
	}
	return &s3.ListObjectsV2Output{Contents: contents}, mock.listErr
}

func (mock *mockClient) DeleteObject(*s3.DeleteObjectInput) (input *s3.DeleteObjectOutput, err error) {
	return &s3.DeleteObjectOutput{}, mock.deleteErr
}

func (mock *mockClient) WaitUntilObjectNotExists(input *s3.HeadObjectInput) error {
	return mock.deleteErr
}

type mockDownloader struct {
	s3manageriface.DownloaderAPI
	mockItem mockItem
	err      error
}

var _ s3manageriface.DownloaderAPI = &mockDownloader{}

func (mock *mockDownloader) Download(writer io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (size int64, err error) {
	writer.WriteAt(mock.mockItem.bytes, 0)
	return int64(len(mock.mockItem.bytes)), mock.err
}

func TestInitCache(t *testing.T) {
	t.Parallel()
	type args struct {
		provider s3Provider
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]*Email
		wantErr bool
	}{
		{
			name: "no emails",
			args: args{
				provider: s3Provider{
					client: &mockClient{},
				},
			},
		},
		{
			name: "emails",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
							{
								key:  "def456",
								size: 2000,
							},
							{
								key:  "ghi789",
								size: 3000,
							},
						},
					},
				},
			},
			want: map[int]*Email{
				1: {
					ID:   "abc123",
					Size: 1000,
				},
				2: {
					ID:   "def456",
					Size: 2000,
				},
				3: {
					ID:   "ghi789",
					Size: 3000,
				},
			},
		},
		{
			name: "overwrite cache",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "should be loaded as the only item as cache is overwritten",
								size: 0000,
							},
						},
					},
					cache: &s3Cache{
						emails: map[int]*Email{
							1: {
								ID:   "abc123",
								Size: 1000,
							},
							2: {
								ID:   "def456",
								Size: 2000,
							},
						},
					},
				},
			},
			want: map[int]*Email{
				1: {
					ID:   "should be loaded as the only item as cache is overwritten",
					Size: 0000,
				},
			},
		},
		{
			name: "error",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						listErr: fmt.Errorf("this should fail"),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.args.provider.initCache()
			assert.EqualValues(t, tt.wantErr, err != nil)
			if !tt.wantErr {
				got := tt.args.provider.cache.emails
				assert.EqualValues(t, len(tt.want), len(got))
				for id, email := range got {
					assert.Contains(t, tt.want, id)
					assert.EqualValues(t, tt.want[id], email)
				}
			}
		})
	}
}

func TestListEmails(t *testing.T) {
	t.Parallel()
	type args struct {
		provider   s3Provider
		notNumbers []int
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]*Email
		wantErr bool
	}{
		{
			name: "no emails",
			args: args{
				provider: s3Provider{
					client: &mockClient{},
				},
			},
		},
		{
			name: "no emails excluded",
			args: args{
				provider: s3Provider{
					client: &mockClient{},
				},
				notNumbers: []int{2},
			},
		},
		{
			name: "emails",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
							{
								key:  "def456",
								size: 2000,
							},
							{
								key:  "ghi789",
								size: 3000,
							},
						},
					},
				},
			},
			want: map[int]*Email{
				1: {
					ID:   "abc123",
					Size: 1000,
				},
				2: {
					ID:   "def456",
					Size: 2000,
				},
				3: {
					ID:   "ghi789",
					Size: 3000,
				},
			},
		},
		{
			name: "emails excluded",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
							{
								key:  "def456",
								size: 2000,
							},
							{
								key:  "ghi789",
								size: 3000,
							},
						},
					},
				},
				notNumbers: []int{-10, 2, 7},
			},
			want: map[int]*Email{
				1: {
					ID:   "abc123",
					Size: 1000,
				},
				3: {
					ID:   "ghi789",
					Size: 3000,
				},
			},
		},
		{
			name: "cache",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "should not be loaded as cache is present",
								size: 0000,
							},
						},
					},
					cache: &s3Cache{
						emails: map[int]*Email{
							1: {
								ID:   "abc123",
								Size: 1000,
							},
							2: {
								ID:   "def456",
								Size: 2000,
							},
							3: {
								ID:   "ghi789",
								Size: 3000,
							},
						},
					},
				},
				notNumbers: []int{2},
			},
			want: map[int]*Email{
				1: {
					ID:   "abc123",
					Size: 1000,
				},
				3: {
					ID:   "ghi789",
					Size: 3000,
				},
			},
		},
		{
			name: "error",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						listErr: fmt.Errorf("this should fail"),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.args.provider.ListEmails(tt.args.notNumbers)
			assert.EqualValues(t, tt.wantErr, err != nil)
			if !tt.wantErr {
				assert.EqualValues(t, len(tt.want), len(got))
				for id, email := range got {
					assert.Contains(t, tt.want, id)
					assert.EqualValues(t, tt.want[id], email)
				}
			}
		})
	}
}

func TestGetEmail(t *testing.T) {
	t.Parallel()
	type args struct {
		provider   s3Provider
		number     int
		notNumbers []int
	}
	tests := []struct {
		name    string
		args    args
		want    *Email
		wantErr bool
	}{
		{
			name: "no emails",
			args: args{
				provider: s3Provider{
					client: &mockClient{},
				},
				number: 1,
			},
			wantErr: true,
		},
		{
			name: "emails",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
							{
								key:  "def456",
								size: 2000,
							},
							{
								key:  "ghi789",
								size: 3000,
							},
						},
					},
				},
				number: 1,
			},
			want: &Email{
				ID:   "abc123",
				Size: 1000,
			},
		},
		{
			name: "emails out of range",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
							{
								key:  "def456",
								size: 2000,
							},
							{
								key:  "ghi789",
								size: 3000,
							},
						},
					},
				},
				number: 7,
			},
			wantErr: true,
		},
		{
			name: "emails excluded",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
							{
								key:  "def456",
								size: 2000,
							},
							{
								key:  "ghi789",
								size: 3000,
							},
						},
					},
				},
				number:     2,
				notNumbers: []int{-8, 2},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.args.provider.GetEmail(tt.args.number, tt.args.notNumbers)
			assert.EqualValues(t, tt.wantErr, err != nil)
			if !tt.wantErr {
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestGetEmaiPayload(t *testing.T) {
	t.Parallel()
	type args struct {
		provider   s3Provider
		number     int
		notNumbers []int
	}
	tests := []struct {
		name    string
		args    args
		want    EmailPayload
		wantErr bool
	}{
		{
			name: "emails out of range",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
						},
					},
					downloader: &mockDownloader{
						mockItem: mockItem{
							bytes: []byte("Hello World!"),
						},
					},
				},
				number: 7,
			},
			wantErr: true,
		},
		{
			name: "emails excluded",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
						},
					},
					downloader: &mockDownloader{
						mockItem: mockItem{
							bytes: []byte("Hello World!"),
						},
					},
				},
				number:     2,
				notNumbers: []int{-8, 2},
			},
			wantErr: true,
		},
		{
			name: "download",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
						},
					},
					downloader: &mockDownloader{
						mockItem: mockItem{
							bytes: []byte("Hello World!"),
						},
					},
				},
				number:     1,
				notNumbers: []int{-8, 2},
			},
			want: []byte("Hello World!"),
		},
		{
			name: "cache payload not loaded",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
							{
								key:  "def456",
								size: 2000,
							},
						},
					},
					downloader: &mockDownloader{
						mockItem: mockItem{
							bytes: []byte("Hello World!"),
						},
					},
					cache: &s3Cache{
						emails: map[int]*Email{
							1: {
								ID:   "abc123",
								Size: 1000,
							},
							2: {
								ID:   "def456",
								Size: 2000,
							},
						},
					},
				},
				number:     2,
				notNumbers: []int{-8, 9},
			},
			want: []byte("Hello World!"),
		},
		{
			name: "cache payload loaded",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
							{
								key:  "def456",
								size: 2000,
							},
						},
					},
					downloader: &mockDownloader{
						mockItem: mockItem{
							bytes: []byte("This message should not be loaded as the message is already cached"),
						},
					},
					cache: &s3Cache{
						emails: map[int]*Email{
							1: {
								ID:   "abc123",
								Size: 1000,
							},
							2: {
								ID:   "def456",
								Size: 2000,
								Payload: func() *EmailPayload {
									var greeting EmailPayload
									greeting = []byte("This is the message that should be loaded")
									return &greeting
								}(),
							},
						},
					},
				},
				number:     2,
				notNumbers: []int{-8, 9},
			},
			want: []byte("This is the message that should be loaded"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.args.provider.GetEmailPayload(tt.args.number, tt.args.notNumbers)
			assert.EqualValues(t, tt.wantErr, err != nil)
			if !tt.wantErr {
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestDeleteEmail(t *testing.T) {
	t.Parallel()
	type args struct {
		provider s3Provider
		number   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "emails out of range",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
						},
					},
				},
				number: 7,
			},
			wantErr: true,
		},
		{
			name: "delete",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
						},
					},
				},
				number: 1,
			},
		},
		{
			name: "delete cache",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
						},
					},
					cache: &s3Cache{
						emails: map[int]*Email{
							1: {
								ID:   "abc123",
								Size: 1000,
							},
							2: {
								ID:   "def456",
								Size: 2000,
							},
						},
					},
				},
				number: 2,
			},
		},
		{
			name: "delete error",
			args: args{
				provider: s3Provider{
					client: &mockClient{
						items: []mockItem{
							{
								key:  "abc123",
								size: 1000,
							},
						},
						deleteErr: fmt.Errorf("this should fail"),
					},
				},
				number: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.args.provider.DeleteEmail(tt.args.number)
			assert.EqualValues(t, tt.wantErr, err != nil)
		})
	}
}
