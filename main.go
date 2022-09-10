/*
   Copyright 2022 Markus Hinz

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
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/markushinz/aws-ses-pop3-server/pkg/handler"
	"github.com/markushinz/aws-ses-pop3-server/pkg/provider"
	"github.com/markushinz/aws-ses-pop3-server/pkg/server"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	v.SetEnvPrefix("POP3")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/aws-ses-pop3-server/")
	v.AddConfigPath("$HOME/.aws-ses-pop3-server")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
			log.Fatal(fmt.Sprintf("Fatal error loadConfig(): %v", err))
		}
	}
	providerCreator := initProviderCreator(v)
	handlerCreator := initHandlerCreator(v, providerCreator)
	serverCreator := initServerCreator(v, handlerCreator)
	server := serverCreator()
	server.Listen()
}

func initProviderCreator(v *viper.Viper) provider.ProviderCreator {
	if v.IsSet("jwt-secret") {
		return provider.NewJWTProviderCreator(
			v.GetString("jwt-secret"),
		)
	}

	if v.IsSet("http-basic-auth-url") {
		rawURL := v.GetString("http-basic-auth-url")
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			log.Fatal("Fatal error initProviderCreator(): Cannot parse http-basic-auth-url")
		}

		v.SetDefault("http-basic-auth-url-insecure", false)
		if !(parsedURL.Scheme == "https" ||
			parsedURL.Hostname() == "localhost" ||
			parsedURL.Hostname() == "127.0.0.1" ||
			parsedURL.Hostname() == "[::1]" ||
			v.GetBool("http-basic-auth-url-insecure")) {
			log.Fatal("Fatal error initProviderCreator(): http-basic-auth-url uses the insecure http protocol")
		}

		return provider.NewHTTPBasicAuthProviderCreator(
			10*time.Second,
			rawURL,
		)
	}

	if !v.IsSet("user") {
		log.Print("Warning: No user specified. \"user\" will be used")
	}
	v.SetDefault("user", "user")
	if !v.IsSet("password") {
		log.Print("Warning: No password specified. \"changeit\" will be used. DO NOT USE IN PRODUCTION!")
	}
	v.SetDefault("password", "changeit")
	staticCreds := provider.StaticCredentials{
		User:     v.GetString("user"),
		Password: v.GetString("password"),
	}
	if v.IsSet("aws-access-key-id") && v.IsSet("aws-secret-access-key") {
		v.SetDefault("aws-s3-prefix", "")
		if !v.IsSet("aws-s3-region") {
			log.Fatal("Fatal error initProviderCreator(): No aws-s3-region specified")
		}
		if !v.IsSet("aws-s3-bucket") {
			log.Fatal("Fatal error initProviderCreator(): No aws-s3-bucket specified")
		}
		staticCreds.S3Bucket = &provider.S3Bucket{
			AWSAccessKeyID:     v.GetString("aws-access-key-id"),
			AWSSecretAccessKey: v.GetString("aws-secret-access-key"),
			Region:             v.GetString("aws-s3-region"),
			Bucket:             v.GetString("aws-s3-bucket"),
			Prefix:             v.GetString("aws-s3-prefix"),
		}
	}

	return provider.NewStaticCredentialsProviderCreator(staticCreds)
}

func initHandlerCreator(v *viper.Viper, providerCreator provider.ProviderCreator) handler.HandlerCreator {
	v.SetDefault("verbose", false)
	return handler.NewPOP3HandlerCreator(
		providerCreator,
		v.GetBool("verbose"),
	)
}

func initServerCreator(v *viper.Viper, handlerCreator handler.HandlerCreator) server.ServerCreator {
	var certificate tls.Certificate
	var err error
	if v.IsSet("tls-cert") && v.IsSet("tls-key") {
		certificate, err = tls.X509KeyPair(
			[]byte(v.GetString("tls-cert")),
			[]byte(v.GetString("tls-key")),
		)
	} else if v.IsSet("tls-cert-path") && v.IsSet("tls-key-path") {
		certificate, err = tls.LoadX509KeyPair(
			v.GetString("tls-cert-path"),
			v.GetString("tls-key-path"),
		)
	} else {
		log.Print("Warning: No tls-cert / tls-key or tls-cert-path / tls-cert-key specified. TLS will be disabled. DO NOT USE IN PRODUCTION!")
		v.SetDefault("port", 2110)
		return server.NewTCPServerCreator(handlerCreator,
			v.GetString("host"),
			v.GetInt("port"),
		)
	}
	if err != nil {
		log.Fatal(fmt.Sprintf("Fatal error loadServerCreator(): %v", err))
	}
	v.SetDefault("port", 2995)
	return server.NewTCPTLSServerCreator(handlerCreator,
		v.GetString("host"),
		v.GetInt("port"),
		certificate,
	)
}
