#!/bin/bash

# Set the URL you want to request
URL=$1

# Number of times to loop
LOOP_COUNT=$2

# Loop through the curl request
for ((i=1; i<=$LOOP_COUNT; i++)); do
    echo "Sending request $i"
    curl "$URL"
    echo "Sleeping for 1 second..."
    sleep 1
done