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


# Check go fmt
echo "==> Checking that go fmt was run for code ..."
gofmt_files=$(go fmt ./...)
if [[ -n ${gofmt_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${gofmt_files}"
    echo "Use \`go fmt\` to reformat code."
    exit 1
fi

echo "> Go source code is well formatted"
exit 0