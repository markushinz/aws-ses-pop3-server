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
	"fmt"
	"log"
	"net"

	"github.com/markushinz/aws-ses-pop3-server/pkg/handler"
)

type tcpServer struct {
	handlerCreator handler.HandlerCreator
	listener       net.Listener
}

var _ Server = &tcpServer{}

func NewTCPServerCreator(handlerCreator handler.HandlerCreator, host string, port int) ServerCreator {
	return func() (server Server) {
		return newTCPServer(handlerCreator, host, port)
	}
}

func newTCPServer(handlerCreator handler.HandlerCreator, host string, port int) (server *tcpServer) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", host, port))
	if err != nil {
		log.Fatal(fmt.Sprintf("Fatal error: server.Listen(): %v", err))
	}
	return &tcpServer{
		handlerCreator: handlerCreator,
		listener:       listener,
	}
}

func (server *tcpServer) Listen() {
	acceptConnections(server.handlerCreator, server.listener)
}

func (server *tcpServer) Close() error {
	return server.listener.Close()
}
