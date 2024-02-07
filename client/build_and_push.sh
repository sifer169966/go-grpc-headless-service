#!/bin/bash

set -e
env="dev"
tag_name=$(git describe --tags 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "unknown version" );
if [ "$tag_name" == "unknown version" ]; then
    echo "Error: Unable to determine the version."
    exit 1
fi

latest_tag_name="latest"
image_base_url="${ARTIFACT_REGION}-docker.pkg.dev/${ARTIFACT_PROJECT_ID}/demo-grpc-client/$env"

# build in context images
docker build --no-cache --progress=plain \
 -t "$image_base_url:$tag_name" \
 -t "$image_base_url:$latest_tag_name" \
 -f ./client/Dockerfile .

# push all tags for in context images
docker push $image_base_url --all-tags

# clean local images
docker rmi -f "$image_base_url:$tag_name"
docker rmi -f "$image_base_url:$latest_tag_name"