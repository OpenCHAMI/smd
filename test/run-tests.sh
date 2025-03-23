#!/usr/bin/env bash

#
# MIT License
#
# (C) Copyright [2022,2024-2025] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
set -x

function cleanup() {
  echo "Cleaning up containers..."
  docker compose down
  if ! [[ $? -eq 0 ]]; then
    echo "Failed to decompose environment!"
    exit 1
  fi
  exit $1
}

script_path="$(dirname "$(readlink -f "$0")")"

pushd ${script_path}/compose

# Configure docker compose
export COMPOSE_PROJECT_NAME=$RANDOM
export COMPOSE_FILES="-f base.yml -f postgres.yml -f jwt-security.yml -f haproxy-api-gateway.yml -f openchami-svcs.yml -f autocert.yml -f coredhcp.yml -f configurator.yml -f computes.yml"

echo "COMPOSE_PROJECT_NAME: ${COMPOSE_PROJECT_NAME}"
echo "COMPOSE_FILES: $COMPOSE_FILES"

# Get the base containers running
echo "Starting containers..."
# docker compose build --no-cache
# docker compose up --exit-code-from wait-for-smd wait-for-smd
docker compose ${COMPOSE_FILES} up

# # execute the CT smoke tests
# if ! docker compose up --exit-code-from smoke-tests smoke-tests; then
#   echo "CT smoke tests FAILED!"
#   cleanup 1
# fi
# 
# # execute the CT Tavern tests
# if ! docker compose up --exit-code-from tavern-tests tavern-tests; then
#   echo "CT tavern tests FAILED!"
#   cleanup 1
# fi
# 
# # Cleanup
# echo "CT tests PASSED!"
# cleanup 0

popd
