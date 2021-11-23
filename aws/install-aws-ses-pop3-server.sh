GO_TAR=go1.17.3.linux-amd64.tar.gz

rm -f ${GO_TAR}
wget https://golang.org/dl/${GO_TAR}

sudo bash -c "rm -rf /usr/local/go && tar -C /usr/local -xzf ${GO_TAR}"
sudo rm -f /usr/local/bin/go
sudo ln -s /usr/local/go/bin/go /usr/local/bin/go

rm -rf aws-ses-pop3-server

git clone https://github.com/powerstandards/aws-ses-pop3-server
cd aws-ses-pop3-server
git checkout development

go build main.go
