#!/bin/bash

DIR=$1

for file in $DIR/*.yml; do
  envsubst < "$file" | microk8s kubectl apply -f -
done