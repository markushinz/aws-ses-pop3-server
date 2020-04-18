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
)

type noneProvider struct {
}

func NewNoneProviderCreator() func() (Provider, error) {
	return func() (Provider, error) {
		return newNoneProvider()
	}
}

func newNoneProvider() (provider *noneProvider, err error) {
	return &noneProvider{}, nil
}

func (provider *noneProvider) ListEmails(notNumbers []int) (emails map[int]*Email, err error) {
	return make(map[int]*Email), nil
}

func (provider *noneProvider) GetEmail(number int, notNumbers []int) (email *Email, err error) {
	return nil, fmt.Errorf("%v does not exist", number)
}

func (provider *noneProvider) GetEmailPayload(number int, notNumbers []int) (payload EmailPayload, err error) {
	return nil, fmt.Errorf("%v does not exist", number)
}

func (provider *noneProvider) DeleteEmail(number int) (err error) {
	return fmt.Errorf("%v does not exist", number)
}
