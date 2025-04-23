#!/bin/bash

set -euo pipefail

cd examples

go mod tidy

go run ./parse-value <<<'{"n":true}' \
  > parse-value/readme-output.txt

go run ./tokenize-offsets <<<'{"n":true}' \
  > tokenize-offsets/readme-output.txt

go run ./tokenize-log-lax <<<'[01,TRUE,"hello	world",]//test' \
  > tokenize-log-lax/readme-output.txt

go run ./tokenize-sanitize <<<'[01,TRUE,"hello	world",]//test' \
  > tokenize-sanitize/readme-output.txt
