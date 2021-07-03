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
