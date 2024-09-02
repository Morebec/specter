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


./scripts/add_license.sh
./scripts/check_gofmt.sh
./scripts/golangci_lint.sh

# Run tests on main
if [ "$(git rev-parse --abbrev-ref HEAD)" == "main" ]; then
  ./scripts/go_test.sh
fi
