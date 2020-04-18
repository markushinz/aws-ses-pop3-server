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

package handler

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/markushinz/aws-ses-pop3-server/pkg/provider"
)

var mutex sync.Mutex

type pop3Cache struct {
	user *string
	dele []int
}

type pop3Handler struct {
	provider provider.Provider
	user     string
	password string
	verbose  bool
	state    string
	cache    pop3Cache
}

func NewPOP3HandlerCreator(providerCreator provider.ProviderCreator, user, password string, verbose bool) func() (handler Handler, response string, err error) {
	return func() (handler Handler, response string, err error) {
		return newPOP3Handler(providerCreator, user, password, verbose)
	}
}

func newPOP3Handler(providerCreator provider.ProviderCreator, user, password string, verbose bool) (handler *pop3Handler, responses string, err error) {
	provider, err := providerCreator()
	if err != nil {
		return nil, "", err
	}
	response := "+OK"
	pop3log([]string{response}, false, verbose)
	return &pop3Handler{
		provider: provider,
		user:     user,
		password: password,
		verbose:  verbose,
		state:    "AUTHORIZATION",
	}, response, nil
}

func (handler *pop3Handler) Handle(message string) (responses []string, quit bool) {
	pop3log([]string{message}, true, handler.verbose)
	switch handler.state {
	case "AUTHORIZATION":
		switch {
		case message == "CAPA":
			responses = []string{"+OK", "TOP", "UIDL", "USER", "."}
		case strings.HasPrefix(message, "USER"):
			responses = handler.handleUSER(message)
		case strings.HasPrefix(message, "PASS"):
			responses = handler.handlePASS(message)
		default:
			responses = []string{"-ERR"}
		}
	case "TRANSACTION":
		switch {
		case message == "STAT":
			responses = handler.handleSTAT()
		case message == "UIDL":
			responses = handler.handleUIDL()
		case strings.HasPrefix(message, "UIDL"):
			responses = handler.handleUIDLn(message)
		case message == "LIST":
			responses = handler.handleLIST()
		case strings.HasPrefix(message, "LIST"):
			responses = handler.handleLISTn(message)
		case strings.HasPrefix(message, "TOP"):
			responses = handler.handleTOP(message)
		case strings.HasPrefix(message, "RETR"):
			responses = handler.handleRETR(message)
		case strings.HasPrefix(message, "DELE"):
			responses = handler.handleDELE(message)
		case message == "NOOP":
			responses = []string{"+OK"}
		case message == "RSET":
			handler.cache.dele = nil
			responses = []string{"+OK"}
		case message == "QUIT":
			responses = handler.handleQUIT()
			quit = true
		default:
			responses = []string{"-ERR"}
		}
	default:
		responses = []string{"-ERR"}
	}
	pop3log(responses, false, handler.verbose)
	return responses, quit
}

func (handler *pop3Handler) CloseConnection() {
	if handler.state != "AUTHORIZATION" {
		handler.state = "AUTHORIZATION"
		mutex.Unlock()
	}
}

func pop3log(data []string, incoming bool, verbose bool) {
	prefix := "-->"
	if incoming {
		prefix = "<--"
	}
	for index, datum := range data {
		if incoming && strings.HasPrefix(datum, "PASS") {
			datum = "PASS [ *** ]"
		}
		if index == 0 || index == len(data)-1 || verbose {
			log.Printf("%v %v", prefix, datum)
		} else if index == len(data)-2 {
			log.Printf("%v [ ... ]", prefix)
		}
	}
}
func (handler *pop3Handler) handleUSER(message string) (responses []string) {
	if len(strings.Split(message, " ")) < 2 {
		err := fmt.Errorf("invalid message")
		log.Printf("Error handleUSER(): %v", err)
		return []string{"-ERR"}
	}
	user := strings.TrimPrefix(message, "USER ")
	handler.cache.user = &user
	return []string{"+OK"}
}

func (handler *pop3Handler) handlePASS(message string) (responses []string) {
	if len(strings.Split(message, " ")) < 2 {
		err := fmt.Errorf("invalid message")
		log.Printf("Error handlePASS(): %v", err)
		return []string{"-ERR"}
	}
	password := strings.TrimPrefix(message, "PASS ")
	if handler.user == *handler.cache.user && handler.password == password {
		mutex.Lock()
		handler.state = "TRANSACTION"
		return []string{"+OK"}
	}
	return []string{"-ERR"}
}

func (handler *pop3Handler) handleSTAT() (responses []string) {
	emails, err := handler.provider.ListEmails(handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.provider.ListEmails(): %v", err)
		return []string{"-ERR"}
	}
	var totalSize int64
	for _, email := range emails {
		totalSize += email.Size
	}
	return []string{fmt.Sprintf("+OK %v %v", len(emails), totalSize)}
}

func (handler *pop3Handler) handleUIDL() (responses []string) {
	emails, err := handler.provider.ListEmails(handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.provider.ListEmails(): %v", err)
		return []string{"-ERR"}
	}
	responses = append(responses, "+OK")
	for number, email := range emails {
		responses = append(responses, fmt.Sprintf("%v %v", number, email.ID))
	}
	responses = append(responses, ".")
	return responses
}

func (handler *pop3Handler) handleUIDLn(message string) (responses []string) {
	parts := strings.Split(message, " ")
	if len(parts) != 2 {
		err := fmt.Errorf("invalid message")
		log.Printf("Error handleUIDLn(): %v", err)
		return []string{"-ERR"}
	}
	number, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Error handleUIDLn(): %v", err)
		return []string{"-ERR"}
	}
	email, err := handler.provider.GetEmail(number, handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.provider.GetEmail(): %v", err)
		return []string{"-ERR"}
	}
	return []string{fmt.Sprintf("+OK %v %v", number, email.ID)}
}

func (handler *pop3Handler) handleLIST() (responses []string) {
	emails, err := handler.provider.ListEmails(handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.provider.ListEmails(): %v", err)
		return []string{"-ERR"}
	}
	responses = append(responses, "+OK")
	for number, email := range emails {
		responses = append(responses, fmt.Sprintf("%v %v", number, email.Size))
	}
	responses = append(responses, ".")
	return responses
}

func (handler *pop3Handler) handleLISTn(message string) (responses []string) {
	parts := strings.Split(message, " ")
	if len(parts) != 2 {
		err := fmt.Errorf("invalid message")
		log.Printf("Error handleLISTn(): %v", err)
		return []string{"-ERR"}
	}
	number, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Error handleLISTn(): %v", err)
		return []string{"-ERR"}
	}
	email, err := handler.provider.GetEmail(number, handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.provider.GetEmail(): %v", err)
		return []string{"-ERR"}
	}
	return []string{fmt.Sprintf("+OK %v %v", number, email.Size)}
}

func (handler *pop3Handler) handleTOP(message string) (responses []string) {
	parts := strings.Split(message, " ")
	if len(parts) != 3 {
		err := fmt.Errorf("invalid message")
		log.Printf("Error handleTOP(): %v", err)
		return []string{"-ERR"}
	}
	number, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Error handleTOP(): %v", err)
		return []string{"-ERR"}
	}
	x, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Printf("Error handleTOP(): %v", err)
		return []string{"-ERR"}
	}
	payload, err := handler.provider.GetEmailPayload(number, handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.provider.GetEmailPayload(): %v", err)
		return []string{"-ERR"}
	}
	lines, err := payload.ParseHeaders(x)
	if err != nil {
		log.Printf("Error payload.ParseHeaders(): %v", err)
		return []string{"-ERR"}
	}
	responses = append(responses, "+OK")
	for _, line := range lines {
		responses = append(responses, line)
	}
	responses = append(responses, ".")
	return responses
}

func (handler *pop3Handler) handleRETR(message string) (responses []string) {
	parts := strings.Split(message, " ")
	if len(parts) != 2 {
		err := fmt.Errorf("invalid message")
		log.Printf("Error handleRETR(): %v", err)
		return []string{"-ERR"}
	}
	number, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Error handleRETR(): %v", err)
		return []string{"-ERR"}
	}
	payload, err := handler.provider.GetEmailPayload(number, handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.provider.GetEmailPayload(): %v", err)
		return []string{"-ERR"}
	}
	lines, err := payload.ParseAll()
	if err != nil {
		log.Printf("Error payload.ParseAll(): %v", err)
		return []string{"-ERR"}
	}
	responses = append(responses, "+OK")
	for _, line := range lines {
		responses = append(responses, line)
	}
	responses = append(responses, ".")
	return responses
}

func (handler *pop3Handler) handleDELE(message string) (responses []string) {
	parts := strings.Split(message, " ")
	if len(parts) != 2 {
		err := fmt.Errorf("invalid message")
		log.Printf("Error handleDELE(): %v", err)
		return []string{"-ERR"}
	}
	number, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Error handleDELE(): %v", err)
		return []string{"-ERR"}
	}
	handler.cache.dele = append(handler.cache.dele, number)
	return []string{"+OK"}
}

func (handler *pop3Handler) handleQUIT() (responses []string) {
	handler.state = "UPDATE"
	for _, number := range handler.cache.dele {
		err := handler.provider.DeleteEmail(number)
		if err != nil {
			log.Printf("Error handleQUIT(): %v", err)
			return []string{"-ERR"}
		}
	}
	return []string{"+OK"}
}
