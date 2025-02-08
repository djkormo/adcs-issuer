#!/bin/bash
#
# Copyright (c) 2020, 2020 Red Hat, IBM Corporation and others.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# modified by Adam Krawczyk - jamallorock

###############################  utilities  #################################

set -euo pipefail # Exit immediately on error

check_running() {
    local check_pod="$1"
    local prometheus_ns='monitoring'
    local kubectl_cmd="kubectl -n ${prometheus_ns}"

    echo "Info: Waiting for ${check_pod} to become ready..."

    # Use `kubectl wait` for more efficient waiting instead of polling
    if ! "${kubectl_cmd}" get pod -l "app=${check_pod}" --no-headers | grep -q '.'; then
        >&2 echo "Error: No pods found for app=${check_pod}. Exiting."
        exit 1  
    fi
    if ! ${kubectl_cmd} wait --for=condition=Ready pod -l "app=${check_pod}" --timeout=60s; then
        >&2 echo "Error: ${check_pod} failed to become ready within timeout."
        exit 1
    fi
    echo "Info: ${check_pod} is now running."
    "${kubectl_cmd}" get pods -l "app=${check_pod}" 
    echo
}

# Check error code from the last command, exit on failure
check_err() {
    local err=$?
    if [ "${err}" -ne 0 ]; then
        >&2 echo "Error: $*"
        exit 1  # Use standard Unix exit code
    fi
}