#!/bin/bash
# node type: c6420 (clemson), c6320(clemson), or c8220(clemson)
# wicosin: c220g2
nodes="129.114.108.15"

username="cc"
SSH_CMD="ssh -i ~/chameleon.pem "

for server in $nodes; do
	$SSH_CMD $username@$server 'bash -s' < ./setup-local.sh
	# $SSH_CMD $username@$server 'bash -s' < ./download-data.sh
done