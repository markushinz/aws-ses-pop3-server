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
	"bufio"
	"bytes"
)

type ProviderCreator func() (Provider, error)

type Provider interface {
	ListEmails(notNumbers []int) (emails map[int]*email, err error)
	GetEmail(number int, notNumbers []int) (email *email, err error)
	GetEmailPayload(number int, notNumbers []int) (payload emailPayload, err error)
	DeleteEmail(number int, notNumbers []int) (err error)
}

type emailPayload []byte

type email struct {
	ID              string
	Size            int64
	payloadOptional *emailPayload
}

func (payload emailPayload) ParseAll() (lines []string, err error) {
	return parse(payload, true, 0)
}

func (payload emailPayload) ParseHeaders(x int) (lines []string, err error) {
	return parse(payload, false, x)
}

func parse(payload emailPayload, all bool, x int) (lines []string, err error) {
	reader := bytes.NewReader(payload)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if !all && line == "" {
			for x > 0 {
				if scanner.Scan() {
					lines = append(lines, scanner.Text())
					x--
				} else {
					break
				}
			}
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
