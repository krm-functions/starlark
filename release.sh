#!/bin/bash

GITHUB_SHA=$1
IMAGE=$(ko build --base-import-paths --platform="linux/arm64,linux/amd64" -t latest -t "sha-$GITHUB_SHA" ./)
echo "built $IMAGE"
cosign sign "$IMAGE" --yes -a sha="sha-$GITHUB_SHA"

./update-catalog.sh $(echo $IMAGE | cut -d@ -f1) $(echo $IMAGE | cut -d@ -f2)
