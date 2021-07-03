
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

package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	user     = "user"
	password = "password"
	host     = "localhost"
	port     = "3110"
	verbose  = "true"
)

func TestE2E(t *testing.T) {
	viper.SetConfigType("yaml")
	assert.NoError(t, viper.ReadConfig(strings.NewReader(fmt.Sprintf(`
user: %s
password: %s
host: %s
port: %s
verbose: %s
`, user, password, host, port, verbose))))

	providerCreator := initProviderCreator()
	handlerCreator := initHandlerCreator(providerCreator)
	serverCreator := initServerCreator(handlerCreator)
	server := serverCreator()
	go server.Listen()
	defer func() {
		assert.NoError(t, server.Close())
	}()

	connection, err := net.Dial("tcp", host+":"+port)
	assert.NoError(t, err)

	read(t, connection, "+OK")

	write(t, connection, "CAPA")
	read(t, connection, "+OK", "TOP", "UIDL", "USER", ".")

	write(t, connection, "USER "+user)
	read(t, connection, "+OK")

	write(t, connection, "PASS "+password)
	read(t, connection, "+OK")

	write(t, connection, "STAT")
	read(t, connection, "+OK 0 0")

	write(t, connection, "UIDL")
	read(t, connection, "+OK", ".")
}

func read(t *testing.T, connection net.Conn, wants ...string) {
	reader := bufio.NewReader(connection)
	for _, want := range wants {
		bytes, err := reader.ReadBytes('\n')
		assert.NoError(t, err)
		got := strings.TrimRight(string(bytes), "\r\n")
		assert.Equal(t, want, got)
	}
}

func write(t *testing.T, connection net.Conn, data string) {
	_, err := connection.Write([]byte(data + "\r\n"))
	assert.NoError(t, err)
}
