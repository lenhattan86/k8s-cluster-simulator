#!/usr/bin/env bash

if [ -z "$1" ]
then
  message="regular update"
else
  message=$1
fi

cd ../
git pull && git add --all && git commit -m "$message" && git push
cd experiments
