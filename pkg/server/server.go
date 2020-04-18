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
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/markushinz/aws-ses-pop3-server/pkg/handler"
)

type ServerCreator func() (server Server)

type Server interface {
	Listen()
}

func acceptConnections(handlerCreator handler.HandlerCreator, listener net.Listener) {
	log.Printf("Listening on %v", listener.Addr().String())
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error: acceptConnections(): %v", err)
		} else {
			go handleConnection(handlerCreator, connection)
		}
	}
}

func handleConnection(handlerCreator handler.HandlerCreator, connection net.Conn) {
	log.Printf("%v connected", connection.RemoteAddr().String())
	handler, response, err := handlerCreator()
	if err != nil {
		log.Fatal(fmt.Sprintf("Fatal error: handleConnection(): %v", err))
	}
	connection.Write([]byte(response + "\r\n"))
	for {
		bytes, err := bufio.NewReader(connection).ReadBytes('\n')
		if err != nil {
			closeConnection(handler, connection)
			return
		}
		message := strings.TrimRight(string(bytes), "\r\n")
		responses, quit := handler.Handle(message)
		for _, response := range responses {
			connection.Write([]byte(response + "\r\n"))
		}
		if quit {
			closeConnection(handler, connection)
			return
		}
	}
}

func closeConnection(handler handler.Handler, connection net.Conn) {
	log.Printf("%v disconnected", connection.RemoteAddr().String())
	connection.Close()
	handler.CloseConnection()
}
