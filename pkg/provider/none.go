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
)

type noneProvider struct {
	emails map[int]*Email
}

var _ Provider = &noneProvider{}

func newNoneProvider(emails ...Email) (provider *noneProvider, err error) {
	emailsMap := make(map[int]*Email)
	for index, email := range emails {
		emailsMap[index+1] = &email
	}
	return &noneProvider{
		emails: emailsMap,
	}, nil
}

func (provider *noneProvider) ListEmails(notNumbers []int) (emails map[int]*Email, err error) {
	emails = make(map[int]*Email)
	for index, email := range provider.emails {
		emails[index] = email
	}
	for _, notNumber := range notNumbers {
		delete(emails, notNumber)
	}
	return emails, nil
}

func (provider *noneProvider) GetEmail(number int, notNumbers []int) (email *Email, err error) {
	emails, err := provider.ListEmails(notNumbers)
	if err != nil {
		return nil, err
	}
	if email, exists := emails[number]; exists {
		return email, nil
	}
	return nil, fmt.Errorf("%v does not exist", number)
}

func (provider *noneProvider) GetEmailPayload(number int, notNumbers []int) (payload EmailPayload, err error) {
	email, err := provider.GetEmail(number, notNumbers)
	if err != nil {
		return nil, err
	}
	return *email.Payload, nil
}

func (provider *noneProvider) DeleteEmail(number int) (err error) {
	if _, err := provider.GetEmail(number, nil); err != nil {
		return err
	}
	return nil
}
