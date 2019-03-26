#!/bin/bash

# Copyright 2019 Preferred Networks, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Checks whether all files have an appropriate license header.

LICENSE_STR='Licensed under the Apache License, Version 2.0 (the "License");'

cd $(git rev-parse --show-toplevel)

files=$(git ls-files | grep -v vendor | grep -e ".go" -e ".sh" -e ".py")
status=0
for f in ${files[@]}; do
    if ! grep "$LICENSE_STR" $f --quiet; then
        if [ $status -eq 0 ]; then
            echo "The following files are missing license headers."
        fi
        echo "- $f"
        status=1
    fi
done

exit $status
