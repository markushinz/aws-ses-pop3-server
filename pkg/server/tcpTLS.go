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

package server

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/markushinz/aws-ses-pop3-server/pkg/handler"
)

type tcpTLSServer struct {
	handlerCreator handler.HandlerCreator
	host           string
	port           int
	certificate    tls.Certificate
}

func NewTCPTLSServerCreator(handlerCreator handler.HandlerCreator, host string, port int, certificate tls.Certificate) func() (server Server) {
	return func() (server Server) {
		return newTCPTLSServer(handlerCreator, host, port, certificate)
	}
}

func newTCPTLSServer(handlerCreator handler.HandlerCreator, host string, port int, certificate tls.Certificate) (server *tcpTLSServer) {
	return &tcpTLSServer{
		handlerCreator: handlerCreator,
		host:           host,
		port:           port,
		certificate:    certificate,
	}
}

func (server *tcpTLSServer) Listen() {
	config := &tls.Config{Certificates: []tls.Certificate{server.certificate}}
	listener, err := tls.Listen("tcp", fmt.Sprintf("%v:%v", server.host, server.port), config)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fatal error: server.Listen(): %v", err))
	}
	acceptConnections(server.handlerCreator, listener)
}
