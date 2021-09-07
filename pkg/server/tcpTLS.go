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

package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/markushinz/aws-ses-pop3-server/pkg/handler"
)

type tcpTLSServer struct {
	handlerCreator handler.HandlerCreator
	listener       net.Listener
}

var _ Server = &tcpServer{}

func NewTCPTLSServerCreator(handlerCreator handler.HandlerCreator, host string, port int, certificate tls.Certificate) ServerCreator {
	return func() (server Server) {
		return newTCPTLSServer(handlerCreator, host, port, certificate)
	}
}

func newTCPTLSServer(handlerCreator handler.HandlerCreator, host string, port int, certificate tls.Certificate) (server *tcpTLSServer) {
	config := &tls.Config{Certificates: []tls.Certificate{certificate}}
	listener, err := tls.Listen("tcp", fmt.Sprintf("%v:%v", host, port), config)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fatal error: server.Listen(): %v", err))
	}
	return &tcpTLSServer{
		handlerCreator: handlerCreator,
		listener:       listener,
	}
}

func (server *tcpTLSServer) Listen() {
	acceptConnections(server.handlerCreator, server.listener)
}

func (server *tcpTLSServer) Close() error {
	return server.listener.Close()
}
