#!/bin/bash
# If CI is true, run golangci-lint directly. If we're on a developer's
# local machine, run golangci-lint from a container so we're ensuring
# a consistent environment.

set -ex

if [ "$CI" = "true" ];
then
  curl -L https://github.com/golangci/golangci-lint/releases/download/v1.52.2/golangci-lint-1.52.2-linux-386.tar.gz | tar xzv
  mv golangci-lint-1.52.2-linux-386/golangci-lint /usr/local/bin
  golangci-lint "${@}"
else
  DOCKER=${DOCKER:-podman}

  if ! which "$DOCKER" > /dev/null 2>&1;
  then
    echo "$DOCKER not found, please install."
    exit 1
  fi

  $DOCKER run --rm \
    --volume "${PWD}:/go/src/github.com/openshift-eng/ci-test-mapping:z" \
    --workdir /go/src/github.com/openshift-eng/ci-test-mapping \
    docker.io/golangci/golangci-lint:v1.52 \
    golangci-lint "${@}"
fi
