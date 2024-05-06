#!/usr/bin/env bash

# Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/xwlearn/miniokr.


# Common utilities, variables and checks for all build scripts.
set -o errexit
set -o nounset
set -o pipefail

# The root of the build/dist directory
PROJ_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source ${PROJ_ROOT}/scripts/lib/logging.sh

INSECURE_SERVER="127.0.0.1:8999"

Header="-HContent-Type: application/json"
CCURL="curl -f -s -XPOST" # Create
UCURL="curl -f -s -XPUT" # Update
RCURL="curl -f -s -XGET" # Retrieve
DCURL="curl -f -s -XDELETE" # Delete

# 后续补充
