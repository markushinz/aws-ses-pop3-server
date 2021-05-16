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
	"sort"
)

type ProviderCreator func() (Provider, error)

type Provider interface {
	ListEmails(notNumbers []int) (emails map[int]*Email, err error)
	GetEmail(number int, notNumbers []int) (email *Email, err error)
	GetEmailPayload(number int, notNumbers []int) (payload EmailPayload, err error)
	DeleteEmail(number int) (err error)
}

type EmailPayload []byte

type Email struct {
	ID              string
	Size            int64
	payloadOptional *EmailPayload
}

func (payload EmailPayload) ParseAll() (lines []string, err error) {
	return parse(payload, true, 0)
}

func (payload EmailPayload) ParseHeaders(x int) (lines []string, err error) {
	return parse(payload, false, x)
}

func parse(payload EmailPayload, all bool, x int) (lines []string, err error) {
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

func GetSortedMailNumbers(emails map[int]*Email) []int {
	var keys []int
	for key := range emails {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}
