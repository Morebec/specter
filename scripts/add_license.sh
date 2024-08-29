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


if ! command -v addlicense &> /dev/null
then
    echo "/!\ addlicense is not installed, installing ..."
    go install github.com/morebec/addlicense@latest
    echo "===> addlicense was installed."
fi

echo "===> Running addlicense ..."
addlicense . && echo "> License added to Go source files successfully."