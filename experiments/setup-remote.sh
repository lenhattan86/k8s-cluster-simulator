#!/bin/bash
# node type: c6420 (clemson), c6320(clemson), or c8220(clemson)

nodes="clnode221.clemson.cloudlab.us
clnode246.clemson.cloudlab.us
"
username="tanle"
SSH_CMD="ssh "

for server in $nodes; do
	$SSH_CMD $username@$server 'bash -s' < ./setup-local.sh
done