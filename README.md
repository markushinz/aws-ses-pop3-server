# aws-ses-pop3-server

### Docker 🐳 / docker-compose / Kubernetes

[`docker.pkg.github.com/markushinz/aws-ses-pop3-server/aws-ses-pop3-server:1.0.0`](https://github.com/markushinz/aws-ses-pop3-server/packages/190841?version=1.0.0)

### Linux / macOS

```bash
sudo curl -L "https://github.com/markushinz/aws-ses-pop3-server/releases/download/aws-ses-pop3-server-$(uname -s)" -o /usr/local/bin/aws-ses-pop3-server
sudo chmod +x /usr/local/bin/aws-ses-pop3-server
```

## Config

aws-ses-pop3-server can be configured using environment variables and / or a config file.
aws-ses-pop3-server looks for config files at the following locations and in the depicted order:

* `/etc/aws-ses-pop3-server/config.yaml`
* `$HOME/.aws-ses-pop3-server/config.yaml` (`~/config.yaml`)
* `$(pwd)/config.yaml` (present working directory)

Environment variables use the prefix `POP3_` followed by the config key where `-` have to be replaced with `_`. Environment variables take precedence.

Check the following example `config.yaml` for possible keys:

```yaml
# The following aws-* keys are optional, but required if you want to load emails
aws-access-key-id: "AKIAIOSFODNN7EXAMPLE"
aws-secret-access-key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

# The following aws-s3-* keys are required iff you set aws-access-key-id and aws-secret-access-key
aws-s3-region: "eu-central-1"
aws-s3-bucket: "aws-ses-pop3-server"
aws-s3-prefix: "" # optional, defaults to ""

verbose: false # optional, defaults to false
user: "jane.doe@example.com" # optional, defaults to "user"
password: "6xRkiWA4mZBSaNmv" # optional, defaults to "changeme"
tls-cert: |- # optional, only valid in combination with tls-key, takes precedence over tls-cert-path / tls-key -path
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
port: 110 # optional, defaults to 110 (or 995 if you specified tls-cert / tls-key or tls-cert-path / tls-key-path)
```




https://minikube.sigs.k8s.io/docs/tutorials/nginx_tcp_udp_ingress/
