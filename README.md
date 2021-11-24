# aws-ses-pop3-server 💌

[![CI](https://github.com/markushinz/aws-ses-pop3-server/actions/workflows/ci.yaml/badge.svg)](https://github.com/markushinz/aws-ses-pop3-server/actions/workflows/ci.yaml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=markushinz_aws-ses-pop3-server&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=markushinz_aws-ses-pop3-server)

The missing POP3 server for [Amazon Simple Email Service](https://aws.amazon.com/de/ses/) - written in golang.
Tested with Apple Mail 14.0 on macOS 11.1, Apple Mail on iOS 14.1 and Microsoft Outlook for Mac 16.45.

AWS SES is powerful when it comes to sending emails but has only limited functionality to receive them.
Officially, only storing them in [Amazon S3](https://aws.amazon.com/de/s3/) and triggering [Amazon Lambda](https://aws.amazon.com/de/lambda/) functions is supported (in certain regions such as *eu-west-1*).

This implementation serves a fully compliant [RFC1939](https://tools.ietf.org/html/rfc1939) POP3 server backed with an S3 bucket for SES.

## Usage

First, follow the official tutorial [Receiving Email with Amazon SES](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/receiving-email.html) to store emails in a S3 bucket.

Next, create an IAM user that has read and write permissions to the desired S3 bucket. [Create a config file](#config) and start the server using one of the followig options.

> Restrict access to your local machine or use TLS!

> Change the default values for user and password!

Finally, configure your favorite email client using the POP3 credentials from your config file 🥳.
Follow the official tutorial [Using the Amazon SES SMTP Interface to Send Email](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/send-email-smtp.html) to obtain SMTP credentials for sending emails.

### Docker 🐳 / docker-compose / Kubernetes

[`markushinz/aws-ses-pop3-server`](https://hub.docker.com/r/markushinz/aws-ses-pop3-server/tags)

*Hint*: If you want to deploy aws-ses-pop3-server to Kubernetes check [this tutorial](https://minikube.sigs.k8s.io/docs/tutorials/nginx_tcp_udp_ingress/) on how to configure your NGINX Ingress Controller.

### Linux / macOS

```shell
sudo curl -L "https://github.com/markushinz/aws-ses-pop3-server/releases/latest/download/aws-ses-pop3-server-$(uname -m)-$(uname -s)" -o /usr/local/bin/aws-ses-pop3-server
sudo chmod +x /usr/local/bin/aws-ses-pop3-server
aws-ses-pop3-server
```

## Authorization

There are two distinct mechanisms for authentication.

One way is to provide "user" and "password" entries in the Config file (see below).
If specified, only USER and PASS exchanges that match these values are considered `authorized`.

In order to be useful, "aws-access-key-id", "aws-secret-access-key" and "aws-region" should also be specified.

The other way is to provide the name of an Authorization Lambda function which is invoked after USER and PASS have been entered.

The Lambda should expect to be called with { user, password } parameters and will return "OK" (or XXXXX) if the user/password is authorized for access.

In order to call the Authorization Lambda, you will need to specify "aws-access-key-id", "aws-secret-access-key" and "aws-region".

These two methods are meant to be mutually exclusive, and a startup error will be thrown if you specify configuration for both methods.

## Config

aws-ses-pop3-server can be configured using environment variables and / or a config file.
aws-ses-pop3-server looks for config files at the following locations and in the depicted order:

* `/etc/aws-ses-pop3-server/config.yaml`
* `$HOME/.aws-ses-pop3-server/config.yaml` (`~/.aws-ses-pop3-server/config.yaml`)
* `$(pwd)/config.yaml` (present working directory)

Environment variables use the prefix `POP3_` followed by the config key where `-` have to be replaced with `_`. Environment variables take precedence.

Check the following example `config.yaml` for possible keys:

```yaml
# The following aws-* keys are optional but required if you want to load emails
# These values have to be set here and are not inferred from other envrionment variables or ~/.aws/credentials
# You need read and write permissions to the desired S3 bucket
aws-access-key-id: "AKIAIOSFODNN7EXAMPLE"
aws-secret-access-key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

# The following aws-s3-* keys are required iff you set aws-access-key-id and aws-secret-access-key
aws-s3-region: "eu-central-1"
aws-s3-bucket: "aws-ses-pop3-server"
aws-s3-prefix: "" # optional, defaults to "" (set this if the emails are not stored in the root directory of the S3 bucket)

authorization-lambda: "" # optional (name of the lambda to invoke for authenticating "user" and "password")

verbose: false # optional, defaults to false
user: "jane.doe@example.com" # optional, defaults to "user"
password: "6xRkiWA4mZBSaNmv" # optional, defaults to "changeit". DO CHANGE IT!
tls-cert: |- # optional, only valid in combination with tls-key, takes precedence over tls-cert-path / tls-key-path
  -----BEGIN CERTIFICATE-----
  [ ... ]
  -----END CERTIFICATE-----
tls-key: |- # optional, only valid in combination with tls-cert, takes precedence over tls-cert-path / tls-key-path
  -----BEGIN PRIVATE KEY-----
  [ ... ]
  -----END PRIVATE KEY-----
tls-cert-path: "etc/aws-ses-pop3-server/tls.crt"  # optional, only valid in combination with tls-key-path
tls-key-path: "etc/aws-ses-pop3-server/tls"  # optional, only valid in combination with tls-cert-path
host: "localhost" # optional, defaults to "" (or 0.0.0.0; [::];  listening on all NICs)
port: 2110 # optional, defaults to 2110 (or 2995 if you specified tls-cert / tls-key or tls-cert-path / tls-key-path)
```
