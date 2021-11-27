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

	"github.com/golang-jwt/jwt"
	"github.com/markushinz/aws-ses-pop3-server/pkg/provider"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host    = "localhost"
	port    = "3110"
	verbose = "true"
)

func TestE2E(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]string
		run    func(t *testing.T, connection net.Conn)
	}{
		{
			name: "static credentials",
			config: map[string]string{
				"user":     "user",
				"password": "password",
			},
			run: func(t *testing.T, connection net.Conn) {
				read(t, connection, "+OK")

				write(t, connection, "CAPA")
				read(t, connection, "+OK", "TOP", "UIDL", "USER", ".")

				write(t, connection, "USER user")
				read(t, connection, "+OK")

				write(t, connection, "PASS password")
				read(t, connection, "+OK")

				write(t, connection, "STAT")
				read(t, connection, "+OK 0 0")

				write(t, connection, "UIDL")
				read(t, connection, "+OK", ".")

				write(t, connection, "LIST")
				read(t, connection, "+OK", ".")

				write(t, connection, "NOOP")
				read(t, connection, "+OK")

				write(t, connection, "RSET")
				read(t, connection, "+OK")

				write(t, connection, "QUIT")
				read(t, connection, "+OK")
			},
		},
		{
			name: "JWT none",
			config: map[string]string{
				"jwt-secret": "secret",
			},
			run: func(t *testing.T, connection net.Conn) {
				read(t, connection, "+OK")

				write(t, connection, "USER jwt")
				read(t, connection, "+OK")

				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, provider.JWTClaims{
					Provider: "none",
				}).SignedString([]byte("secret"))
				assert.NoError(t, err)
				write(t, connection, "PASS "+token)
				read(t, connection, "+OK")

				write(t, connection, "QUIT")
				read(t, connection, "+OK")
			},
		},
		{
			name: "JWT demo",
			config: map[string]string{
				"jwt-secret": "secret",
			},
			run: func(t *testing.T, connection net.Conn) {
				read(t, connection, "+OK")

				write(t, connection, "USER jwt")
				read(t, connection, "+OK")

				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, provider.JWTClaims{
					Provider: "demo",
				}).SignedString([]byte("secret"))
				assert.NoError(t, err)
				write(t, connection, "PASS "+token)
				read(t, connection, "+OK")

				write(t, connection, "STAT")
				read(t, connection, fmt.Sprintf("+OK 1 %v", provider.DemoEmail.Size))

				write(t, connection, "UIDL")
				read(t, connection, "+OK", "1 "+provider.DemoEmail.ID, ".")

				write(t, connection, "LIST")
				read(t, connection, "+OK", fmt.Sprintf("1 %v", provider.DemoEmail.Size), ".")

				write(t, connection, "RETR 1")
				wants := []string{"+OK"}
				wants = append(wants, strings.Split(string(*provider.DemoEmail.Payload), "\n")...)
				wants = append(wants, ".")
				read(t, connection, wants...)

				write(t, connection, "DELE 1")
				read(t, connection, "+OK")

				write(t, connection, "QUIT")
				read(t, connection, "+OK")
			},
		},
		{
			name: "JWT s3",
			config: map[string]string{
				"jwt-secret": "secret",
			},
			run: func(t *testing.T, connection net.Conn) {
				read(t, connection, "+OK")

				write(t, connection, "USER jwt")
				read(t, connection, "+OK")

				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, provider.JWTClaims{
					Provider: "s3",
				}).SignedString([]byte("secret"))
				assert.NoError(t, err)
				write(t, connection, "PASS "+token)
				read(t, connection, "+OK")
			},
		},
		{
			name: "JWT invalid",
			config: map[string]string{
				"jwt-secret": "secret",
			},
			run: func(t *testing.T, connection net.Conn) {
				read(t, connection, "+OK")

				write(t, connection, "USER jwt")
				read(t, connection, "+OK")

				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, provider.JWTClaims{
					Provider: "demo",
				}).SignedString([]byte("invalid"))
				assert.NoError(t, err)
				write(t, connection, "PASS "+token)
				read(t, connection, "-ERR")
			},
		},
		{
			name: "JWT unknown provider",
			config: map[string]string{
				"jwt-secret": "secret",
			},
			run: func(t *testing.T, connection net.Conn) {
				read(t, connection, "+OK")

				write(t, connection, "USER jwt")
				read(t, connection, "+OK")

				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, provider.JWTClaims{
					Provider: "unknown",
				}).SignedString([]byte("secret"))
				assert.NoError(t, err)
				write(t, connection, "PASS "+token)
				read(t, connection, "-ERR")

				write(t, connection, "STAT")
				read(t, connection, "-ERR")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetConfigType("yaml")
			tt.config["host"] = host
			tt.config["port"] = port
			tt.config["verbose"] = verbose
			yaml := ""
			for key, value := range tt.config {
				yaml += fmt.Sprintf("%s: %s\n", key, value)
			}
			require.NoError(t, viper.ReadConfig(strings.NewReader(yaml)))
			defer viper.Reset()
			providerCreator := initProviderCreator()
			handlerCreator := initHandlerCreator(providerCreator)
			serverCreator := initServerCreator(handlerCreator)
			server := serverCreator()
			go server.Listen()
			defer func() {
				require.NoError(t, server.Close())
			}()
			connection, err := net.Dial("tcp", host+":"+port)
			require.NoError(t, err)
			tt.run(t, connection)
		})
	}
}

func read(t *testing.T, connection net.Conn, wants ...string) {
	reader := bufio.NewReader(connection)
	for _, want := range wants {
		bytes, err := reader.ReadBytes('\n')
		require.NoError(t, err)
		got := strings.TrimRight(string(bytes), "\r\n")
		require.Equal(t, want, got)
	}
}

func write(t *testing.T, connection net.Conn, data string) {
	_, err := connection.Write([]byte(data + "\r\n"))
	require.NoError(t, err)
}
