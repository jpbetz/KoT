#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

VERSION=v0.2.2-1-g747dd959
CONTROLLER_GEN_BASENAME=controller-gen-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64
CONTROLLER_GEN=${CONTROLLER_GEN_BASENAME}-${VERSION}

test -x "hack/${CONTROLLER_GEN}" || curl -f -L -o "hack/${CONTROLLER_GEN}" "https://github.com/openshift/kubernetes-sigs-controller-tools/releases/download/${VERSION}/${CONTROLLER_GEN_BASENAME}"
chmod +x "hack/${CONTROLLER_GEN}"

hack/${CONTROLLER_GEN} schemapatch:manifests=./manifests/kubernetes-1.15-crds paths="./apis/..." output:dir=./manifests/kubernetes-1.15-crds
hack/${CONTROLLER_GEN} schemapatch:manifests=./manifests/kubernetes-1.16-crds paths="./apis/..." output:dir=./manifests/kubernetes-1.16-crds
