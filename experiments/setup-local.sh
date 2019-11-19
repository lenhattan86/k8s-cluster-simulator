#!/bin/bash

sudo apt update
sudo apt-get install -y software-properties-common
## setup go

wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz; \
sudo tar -xvf go1.13.3.linux-amd64.tar.gz; \
sudo mv go /usr/local

export GOROOT=/usr/local/go
export PATH=/usr/local/go/bin:$PATH
echo """export GOROOT=/usr/local/go
export PATH=/usr/local/go/bin:$PATH
""" >> .bashrc

## setup python, pip and its library.
sudo apt install -y python-matplotlib
sudo apt-get -y install python-pandas
sudo apt-get -y install python-numpy

# sudo add-apt-repository ppa:jonathonf/python-3.6
# sudo apt-get update
# sudo apt-get install python3.6
# sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.5 1
# sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.6 2
# sudo update-alternatives --config python30
#sudo apt-get install -y python-pip python3-pip
# sudo pip3 install pandas
# sudo pip3 install --upgrade pip
#sudo python3 -m pip install -U matplotlib



## download data
cd /proj/yarnrm-PG0 
mkdir google-trace-data
cd /proj/yarnrm-PG0/google-trace-data
wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1mh3eWQUr0_Y8fkBqiZydN186NgwSJQQh' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1mh3eWQUr0_Y8fkBqiZydN186NgwSJQQh" \
  -O tasks.tar && rm -rf /tmp/cookies.txt
sudo tar -xvf tasks.tar
mv tasks-new tasks

wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1ymFRBvW1wKIHrdi-v5wyzJLKu81ZxFOx' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1ymFRBvW1wKIHrdi-v5wyzJLKu81ZxFOx" \
  -O machines.tar && rm -rf /tmp/cookies.txt
sudo tar -xvf machines.tar

## download code
cd ~/
mkdir ~/go; mkdir ~/go/src; mkdir ~/go/src/github.com
mkdir ~/go/src/github.com/pfnet-research
cd ~/go/src/github.com/pfnet-research
git clone https://github.com/lenhattan86/k8s-cluster-simulator