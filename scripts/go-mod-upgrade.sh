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


  if ! command -v go-mod-upgrade &> /dev/null
  then
      echo "/!\ go-mod-upgrade is not installed, installing ..."
      # binary will be $(go env GOPATH)/bin/go-mod-upgrade
      go install github.com/oligot/go-mod-upgrade@latest

      echo "===> go-mod-upgrade was installed."
  fi

  echo "===> Running go-mod-upgrade ..."
  echo "Colors in module names help identify the update type:
  - GREEN for a minor update
  - YELLOW for a patch update
  - RED for a prerelease update
"
  go-mod-upgrade
  echo "Running go mod tidy ..."
  go mod tidy