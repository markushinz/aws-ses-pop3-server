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
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"github.com/markushinz/aws-ses-pop3-server/pkg/handler"
	"github.com/markushinz/aws-ses-pop3-server/pkg/provider"
	"github.com/markushinz/aws-ses-pop3-server/pkg/server"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("POP3")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/aws-ses-pop3-server/")
	viper.AddConfigPath("$HOME/.aws-ses-pop3-server")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
			log.Fatal(fmt.Sprintf("Fatal error loadConfig(): %v", err))
		}
	}
	providerCreator := initProviderCreator()
	handlerCreator := initHandlerCreator(providerCreator)
	serverCreator := initServerCreator(handlerCreator)
	server := serverCreator()
	server.Listen()
}

func initProviderCreator() provider.ProviderCreator {
	if viper.IsSet("jwt-secret") {
		return provider.NewJWTProviderCreator(
			viper.GetString("jwt-secret"),
		)
	}

	if viper.IsSet("authorization-lambda") {
		if !viper.IsSet("aws-access-key-id") || !viper.IsSet("aws-secret-access-key") || !viper.IsSet("aws-s3-region") {
			log.Fatal("Fatal error initProviderCreator: authorization-lambda requires aws-access-key-id, aws-secret-access-key and aws-s3-region")
		}

		s3Bucket := provider.S3Bucket{
			AWSAccessKeyID:     viper.GetString("aws-access-key-id"),
			AWSSecretAccessKey: viper.GetString("aws-secret-access-key"),
			Region:             viper.GetString("aws-s3-region"),
			Bucket:             viper.GetString("aws-s3-bucket"),
			Prefix:             viper.GetString("aws-s3-prefix"),
		}
		return provider.NewLambdaAuthorizationProviderCreator(viper.GetString("authorization-lambda"), s3Bucket)
	}

	if !viper.IsSet("user") {
		log.Print("Warning: No user specified. \"user\" will be used")
	}
	viper.SetDefault("user", "user")
	if !viper.IsSet("password") {
		log.Print("Warning: No password specified. \"changeit\" will be used. DO NOT USE IN PRODUCTION!")
	}
	viper.SetDefault("password", "changeit")
	staticCreds := provider.StaticCredentials{
		User:     viper.GetString("user"),
		Password: viper.GetString("password"),
	}
	if viper.IsSet("aws-access-key-id") && viper.IsSet("aws-secret-access-key") {
		viper.SetDefault("aws-s3-prefix", "")
		if !viper.IsSet("aws-s3-region") {
			log.Fatal("Fatal error initProviderCreator(): No aws-s3-region specified")
		}
		if !viper.IsSet("aws-s3-bucket") {
			log.Fatal("Fatal error initProviderCreator(): No aws-s3-bucket specified")
		}
		staticCreds.S3Bucket = &provider.S3Bucket{
			AWSAccessKeyID:     viper.GetString("aws-access-key-id"),
			AWSSecretAccessKey: viper.GetString("aws-secret-access-key"),
			Region:             viper.GetString("aws-s3-region"),
			Bucket:             viper.GetString("aws-s3-bucket"),
			Prefix:             viper.GetString("aws-s3-prefix"),
		}
	}

	return provider.NewStaticCredentialsProviderCreator(staticCreds)
}

func initHandlerCreator(providerCreator provider.ProviderCreator) handler.HandlerCreator {
	viper.SetDefault("verbose", false)
	return handler.NewPOP3HandlerCreator(
		providerCreator,
		viper.GetBool("verbose"),
	)
}

func initServerCreator(handlerCreator handler.HandlerCreator) server.ServerCreator {
	var certificate tls.Certificate
	var err error
	if viper.IsSet("tls-cert") && viper.IsSet("tls-key") {
		certificate, err = tls.X509KeyPair(
			[]byte(viper.GetString("tls-cert")),
			[]byte(viper.GetString("tls-key")),
		)
	} else if viper.IsSet("tls-cert-path") && viper.IsSet("tls-key-path") {
		certificate, err = tls.LoadX509KeyPair(
			viper.GetString("tls-cert-path"),
			viper.GetString("tls-key-path"),
		)
	} else {
		log.Print("Warning: No tls-cert / tls-key or tls-cert-path / tls-cert-key specified. TLS will be disabled. DO NOT USE IN PRODUCTION!")
		viper.SetDefault("port", 2110)
		return server.NewTCPServerCreator(handlerCreator,
			viper.GetString("host"),
			viper.GetInt("port"),
		)
	}
	if err != nil {
		log.Fatal(fmt.Sprintf("Fatal error loadServerCreator(): %v", err))
	}
	viper.SetDefault("port", 2995)
	return server.NewTCPTLSServerCreator(handlerCreator,
		viper.GetString("host"),
		viper.GetInt("port"),
		certificate,
	)
}
