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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()
	type args struct {
		payload EmailPayload
		all     bool
		x       int
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "all",
			args: args{
				payload: []byte(`Date: Mon, 20 Apr 2020 11:134:13 +0200
From: Jane Doe <jane.doe@example.com>
To: john.doe@example.com
Subject: Hello
Content-Type: text/plain; charset=us-ascii; format=flowed
Content-Transfer-Encoding: 7bit

Hi John,

How are you?

Best
Jane`),
				all: true,
			},
			want: []string{"Date: Mon, 20 Apr 2020 11:134:13 +0200",
				"From: Jane Doe <jane.doe@example.com>",
				"To: john.doe@example.com",
				"Subject: Hello",
				"Content-Type: text/plain; charset=us-ascii; format=flowed",
				"Content-Transfer-Encoding: 7bit",
				"",
				"Hi John,",
				"",
				"How are you?",
				"",
				"Best",
				"Jane"},
		},
		{
			name: "all trailing new line",
			args: args{
				payload: []byte(`Date: Mon, 20 Apr 2020 11:134:13 +0200
From: Jane Doe <jane.doe@example.com>
To: john.doe@example.com
Subject: Hello
Content-Type: text/plain; charset=us-ascii; format=flowed
Content-Transfer-Encoding: 7bit

Hi John,

How are you?

Best
Jane
`),
				all: true,
			},
			want: []string{"Date: Mon, 20 Apr 2020 11:134:13 +0200",
				"From: Jane Doe <jane.doe@example.com>",
				"To: john.doe@example.com",
				"Subject: Hello",
				"Content-Type: text/plain; charset=us-ascii; format=flowed",
				"Content-Transfer-Encoding: 7bit",
				"",
				"Hi John,",
				"",
				"How are you?",
				"",
				"Best",
				"Jane"},
		},
		{
			name: "headers",
			args: args{
				payload: []byte(`Date: Mon, 20 Apr 2020 11:134:13 +0200
From: Jane Doe <jane.doe@example.com>
To: john.doe@example.com
Subject: Hello
Content-Type: text/plain; charset=us-ascii; format=flowed
Content-Transfer-Encoding: 7bit

Hi John,

How are you?

Best
Jane`),
				all: false,
			},
			want: []string{"Date: Mon, 20 Apr 2020 11:134:13 +0200",
				"From: Jane Doe <jane.doe@example.com>",
				"To: john.doe@example.com",
				"Subject: Hello",
				"Content-Type: text/plain; charset=us-ascii; format=flowed",
				"Content-Transfer-Encoding: 7bit",
				"",
			},
		},
		{
			name: "headers x",
			args: args{
				payload: []byte(`Date: Mon, 20 Apr 2020 11:134:13 +0200
From: Jane Doe <jane.doe@example.com>
To: john.doe@example.com
Subject: Hello
Content-Type: text/plain; charset=us-ascii; format=flowed
Content-Transfer-Encoding: 7bit

Hi John,

How are you?

Best
Jane`),
				x: 3,
			},
			want: []string{"Date: Mon, 20 Apr 2020 11:134:13 +0200",
				"From: Jane Doe <jane.doe@example.com>",
				"To: john.doe@example.com",
				"Subject: Hello",
				"Content-Type: text/plain; charset=us-ascii; format=flowed",
				"Content-Transfer-Encoding: 7bit",
				"",
				"Hi John,",
				"",
				"How are you?",
			},
		},
		{
			name: "headers x out of range",
			args: args{
				payload: []byte(`Date: Mon, 20 Apr 2020 11:134:13 +0200
From: Jane Doe <jane.doe@example.com>
To: john.doe@example.com
Subject: Hello
Content-Type: text/plain; charset=us-ascii; format=flowed
Content-Transfer-Encoding: 7bit

Hi John,

How are you?

Best
Jane`),
				x: 7000,
			},
			want: []string{"Date: Mon, 20 Apr 2020 11:134:13 +0200",
				"From: Jane Doe <jane.doe@example.com>",
				"To: john.doe@example.com",
				"Subject: Hello",
				"Content-Type: text/plain; charset=us-ascii; format=flowed",
				"Content-Transfer-Encoding: 7bit",
				"",
				"Hi John,",
				"",
				"How are you?",
				"",
				"Best",
				"Jane"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parse(tt.args.payload, tt.args.all, tt.args.x)
			assert.EqualValues(t, tt.wantErr, err != nil)
			assert.EqualValues(t, tt.want, got)
		})
	}
}
