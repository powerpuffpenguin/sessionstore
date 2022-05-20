#!/bin/bash
set -ex
cd "$(dirname $BASH_SOURCE)"

protoc -I pb --go_out='.' --go_opt=paths=source_relative sessionstore/session/session.proto
