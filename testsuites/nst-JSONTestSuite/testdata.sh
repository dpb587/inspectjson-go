#!/bin/bash

set -euo pipefail

git clone --depth=1 https://github.com/nst/JSONTestSuite.git testdata

cd testdata/

git rev-parse HEAD > .git/HEAD

GZIP=-9 tar -czf ../testdata.tar.gz .git/HEAD ./LICENSE* ./test_parsing/*.json

cd ../

rm -fr testdata/
