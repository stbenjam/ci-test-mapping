#!/bin/bash

set -euo pipefail
set -x

make mapping
if ! git --no-pager diff --exit-code data/openshift-gce-devel/ci_analysis_us/component_mapping.json
then
  echo "ERROR: Please run 'make mapping' and commit the result."
  exit 1
fi
