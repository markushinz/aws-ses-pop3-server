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

package handler

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/markushinz/aws-ses-pop3-server/pkg/provider"
)

type pop3Cache struct {
	user     *string
	provider provider.Provider
	dele     []int
}

type pop3Handler struct {
	providerCreator provider.ProviderCreator
	verbose         bool
	cache           pop3Cache
}

var _ Handler = &pop3Handler{}

func NewPOP3HandlerCreator(providerCreator provider.ProviderCreator, verbose bool) HandlerCreator {
	return func() (handler Handler, response string, err error) {
		return newPOP3Handler(providerCreator, verbose)
	}
}

func newPOP3Handler(providerCreator provider.ProviderCreator, verbose bool) (handler *pop3Handler, responses string, err error) {
	if err != nil {
		return nil, "", err
	}
	handler = &pop3Handler{
		providerCreator: providerCreator,
		verbose:         verbose,
	}
	response := "+OK"
	handler.log([]string{response}, false, verbose)
	return handler, response, nil
}

func (handler *pop3Handler) getState() string {
	if handler.cache.provider == nil {
		return "AUTHORIZATION"
	}
	return "TRANSACTION"
}

func (handler *pop3Handler) Handle(message string) (responses []string, quit bool) {
	handler.log([]string{message}, true, handler.verbose)
	switch {
	case message == "CAPA":
		responses = handler.handleCAPA()
	case strings.HasPrefix(message, "USER"):
		responses = handler.handleUSER(message)
	case strings.HasPrefix(message, "PASS"):
		responses = handler.handlePASS(message)
	case handler.getState() != "TRANSACTION":
		responses = []string{"-ERR"}
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
	handler.log(responses, false, handler.verbose)
	return responses, quit
}

func (handler *pop3Handler) log(data []string, incoming bool, verbose bool) {
	prefix := "-->"
	if incoming {
		prefix = "<--"
	}
	for index, datum := range data {
		if incoming && strings.HasPrefix(datum, "PASS") {
			datum = "PASS [ *** ]"
		}
		if index == 0 || index == len(data)-1 || verbose {
			log.Printf("%v %v %v", handler.getState(), prefix, datum)
		} else if index == len(data)-2 {
			log.Printf("%v %v [ ... ]", handler.getState(), prefix)
		}
	}
}
func (handler *pop3Handler) handleCAPA() (responses []string) {
	return []string{"+OK", "TOP", "UIDL", "USER", "."}
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
	provider, err := handler.providerCreator(*handler.cache.user, password)
	if err != nil {
		log.Printf("Error handlePASS(): %v", err)
		return []string{"-ERR"}
	}
	handler.cache.provider = provider
	return []string{"+OK"}
}

func (handler *pop3Handler) handleSTAT() (responses []string) {
	emails, err := handler.cache.provider.ListEmails(handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.cache.provider.ListEmails(): %v", err)
		return []string{"-ERR"}
	}
	var totalSize int64
	for _, email := range emails {
		totalSize += email.Size
	}
	return []string{fmt.Sprintf("+OK %v %v", len(emails), totalSize)}
}

func (handler *pop3Handler) handleUIDL() (responses []string) {
	emails, err := handler.cache.provider.ListEmails(handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.cache.provider.ListEmails(): %v", err)
		return []string{"-ERR"}
	}
	responses = append(responses, "+OK")
	for _, number := range provider.GetSortedMailNumbers(emails) {
		responses = append(responses, fmt.Sprintf("%v %v", number, emails[number].ID))
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
	email, err := handler.cache.provider.GetEmail(number, handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.cache.provider.GetEmail(): %v", err)
		return []string{"-ERR"}
	}
	return []string{fmt.Sprintf("+OK %v %v", number, email.ID)}
}

func (handler *pop3Handler) handleLIST() (responses []string) {
	emails, err := handler.cache.provider.ListEmails(handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.cache.provider.ListEmails(): %v", err)
		return []string{"-ERR"}
	}
	responses = append(responses, "+OK")
	for _, number := range provider.GetSortedMailNumbers(emails) {
		responses = append(responses, fmt.Sprintf("%v %v", number, emails[number].Size))
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
	email, err := handler.cache.provider.GetEmail(number, handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.cache.provider.GetEmail(): %v", err)
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
	payload, err := handler.cache.provider.GetEmailPayload(number, handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.cache.provider.GetEmailPayload(): %v", err)
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
	payload, err := handler.cache.provider.GetEmailPayload(number, handler.cache.dele)
	if err != nil {
		log.Printf("Error handler.cache.provider.GetEmailPayload(): %v", err)
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
	for _, number := range handler.cache.dele {
		err := handler.cache.provider.DeleteEmail(number)
		if err != nil {
			log.Printf("Error handleQUIT(): %v", err)
			return []string{"-ERR"}
		}
	}
	return []string{"+OK"}
}
