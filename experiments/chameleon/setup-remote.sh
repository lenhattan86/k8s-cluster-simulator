#!/bin/bash
# node type: c6420 (clemson), c6320(clemson), or c8220(clemson)
# wincosin: c220g2
nodes="cc@129.114.108.15
cc@129.114.108.73
tanle@c220g2-011010.wisc.cloudlab.us	
tanle@c220g2-011011.wisc.cloudlab.us
"

username="cc"
SSH_CMD="ssh -i ~/chameleon.pem "

for server in $nodes; do
	$SSH_CMD $username@$server 'bash -s' < ./setup-local.sh
	# $SSH_CMD $username@$server 'bash -s' < ./download-data.sh
done