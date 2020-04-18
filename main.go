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
	loadConfig()

	var providerCreator provider.ProviderCreator
	if viper.IsSet("aws-access-key-id") && viper.IsSet("aws-secret-access-key") {
		providerCreator = provider.NewAWSS3ProviderCreator(
			viper.GetString("aws-access-key-id"),
			viper.GetString("aws-secret-access-key"),
			viper.GetString("aws-s3-region"),
			viper.GetString("aws-s3-bucket"),
			viper.GetString("aws-s3-prefix"),
		)
	} else {
		providerCreator = provider.NewNoneProviderCreator()
	}

	var handlerCreator handler.HandlerCreator
	handlerCreator = handler.NewPOP3HandlerCreator(
		providerCreator,
		viper.GetString("user"),
		viper.GetString("password"),
		viper.GetBool("verbose"),
	)

	var serverCreator server.ServerCreator
	if viper.GetBool("tls") {
		serverCreator = server.NewTCPTLSServerCreator(handlerCreator,
			viper.GetString("host"),
			viper.GetInt("port"),
			loadCertificate(),
		)
	} else {
		serverCreator = server.NewTCPServerCreator(handlerCreator,
			viper.GetString("host"),
			viper.GetInt("port"),
		)
	}

	server := serverCreator()
	server.Listen()
}

func loadConfig() {
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
	viper.SetDefault("host", "localhost")
	viper.SetDefault("tls", false)
	if !viper.GetBool("tls") {
		log.Print("Warning: TLS is disabled")
	}
	viper.SetDefault("port", func() int {
		if viper.GetBool("tls") {
			return 995
		}
		return 110
	}())
	viper.SetDefault("verbose", false)
	viper.SetDefault("user", "user")
	viper.SetDefault("password", "changeit")
	if viper.GetString("password") == "changeit" {
		log.Print("Warning: The password is set to the default \"changeit\"")
	}
	if viper.IsSet("aws-access-key-id") && viper.IsSet("aws-secret-access-key") {
		viper.SetDefault("aws-s3-prefix", "")
		if !viper.IsSet("aws-s3-region") {
			log.Fatal("Fatal error loadConfig(): No aws-s3-region specified")
		}
		if !viper.IsSet("aws-s3-bucket") {
			log.Fatal("Fatal error loadConfig(): No aws-s3-bucket specified")
		}
	} else {
		log.Print("Warning: No AWS credentials provided")
	}
}

func loadCertificate() (certificate tls.Certificate) {
	err := fmt.Errorf("no tls-cert / tls-key or tls-cert-path / tls-cert-key specified")
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
	}
	if err != nil {
		log.Fatal(fmt.Sprintf("Fatal error loadCertificate(): %v", err))
	}
	return certificate
}
