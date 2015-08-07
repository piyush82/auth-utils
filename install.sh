#!/bin/bash
sudo apt-get update
sudo apt-get upgrade
sudo apt-get install -y gcc
sudo apt-get install -y git
sudo apt-get install -y uuid-runtime

cd $HOME

echo "downloading and installing go runtime from google servers, please wait ..."
wget https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz

sudo -k

echo "configuring your environment for go projects ..."

cat >> $HOME/.profile << EOF

export GOPATH=$HOME/go
EOF

cd $HOME
source .profile

cat >> $HOME/.profile << EOF
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
EOF

cd $HOME

source .profile

echo "done."

echo "testing new environment variables ..."
echo "GOPATH: $GOPATH"
echo "Path: $PATH"

echo "Downloading auth-utils code now, please wait ..."
mkdir $HOME/go
mkdir -p $HOME/go/src/github.com/piyush82
cd $HOME/go/src/github.com/piyush82
git clone https://github.com/piyush82/auth-utils.git
echo "done."

cd auth-utils
echo "getting all code dependencies for auth-utils now, be patient ~ 1-2 minutes"
go get
echo "done."

echo "compiling and installing the package"
go install
echo "done."

cd

echo "starting the auth-service next, you can start using it at port :8000"
echo "use Ctrl+c to stop it. The executable is located at: $GOPATH/bin/auth-utils"

auth-utils