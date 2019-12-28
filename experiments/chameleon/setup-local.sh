#!/bin/bash
sudo apt update; \
sudo apt-get install -y software-properties-common
## setup go

wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz; \
sudo tar -xvf go1.13.3.linux-amd64.tar.gz; \
sudo mv go /usr/local

export GOROOT=/usr/local/go; \
export PATH=/usr/local/go/bin:$PATH; \
echo """export GOROOT=/usr/local/go
export PATH=/usr/local/go/bin:$PATH
""" >> .bashrc

## pull go libs
go get github.com/golang/protobuf/proto; \
go get github.com/gogo/protobuf/proto

## setup python, pip and its library.
sudo apt install -y python-matplotlib; \
sudo apt install -y python3-matplotlib; \
sudo apt-get -y install python-pandas; \
sudo apt-get -y install python3-pandas; \
sudo apt-get -y install python-numpy; \
sudo apt-get -y install python3-numpy

# sudo add-apt-repository ppa:jonathonf/python-3.6 -y
# sudo apt-get update
# sudo apt-get -y install python3.6
# sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.5 1
# sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.6 2
# sudo update-alternatives --config python30
sudo apt-get install -y python-pip python3-pip
# sudo pip3 install pandas
# sudo pip3 install --upgrade pip
# sudo python3 -m pip install -U matplotlib

## install tensorflow.
sudo pip install tensorflow; \
sudo pip3 install tensorflow

# sudo apt install golang-go
## download code
cd ~/ ; \
mkdir ~/go; mkdir ~/go/src; mkdir ~/go/src/github.com ; \
mkdir ~/go/src/github.com/pfnet-research; \
cd ~/go/src/github.com/pfnet-research; \
git clone https://github.com/lenhattan86/k8s-cluster-simulator 