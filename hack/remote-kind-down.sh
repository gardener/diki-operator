#!/usr/bin/env bash

# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

SOURCE_CLUSTER="diki-source"
TARGET_CLUSTER="diki-target"

echo "Deleting source cluster ($SOURCE_CLUSTER)..."
kind delete cluster --name "$SOURCE_CLUSTER"

echo "Deleting target cluster ($TARGET_CLUSTER)..."
kind delete cluster --name "$TARGET_CLUSTER"

echo "Remote kind clusters deleted."
