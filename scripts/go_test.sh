#!/usr/bin/env bash
# Copyright 2024 Morébec
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


function runUnitTests() {
  echo "===> Running unit tests ..."
  go test -v ./... -coverprofile cover.out

  echo "===> Checking coverage ..."
  go tool cover -html=cover.out #-o cover.html
}

function runMutationTests() {
  if ! command -v gremlins &> /dev/null
  then
      echo "/!\ gremlins is not installed, installing ..."
      # binary will be $(go env GOPATH)/bin/gremlins
      go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.5.0

      echo "===> gremlins was installed."
  fi

  echo "===> Running gremlins ..."
  gremlins unleash && echo "> Mutation tests executed successfully."
}


runUnitTests && runMutationTests