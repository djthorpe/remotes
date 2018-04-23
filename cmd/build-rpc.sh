#!/bin/bash
##############################################################
# Build RPC binaries
##############################################################

CURRENT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GO=`which go`
PROTOC=`which protoc`
LDFLAGS="-w -s"
cd "${CURRENT_PATH}/.."

##############################################################
# Sanity checks

if [ ! -d ${CURRENT_PATH} ] ; then
  echo "Not found: ${CURRENT_PATH}" >&2
  exit -1
fi
if [ "${GO}" == "" ] || [ ! -x ${GO} ] ; then
  echo "go not installed or executable" >&2
  exit -1
fi

##############################################################
# Install RPC binaries

COMMANDS=(
  rpc/remotes-service.go
  rpc/remotes-client.go
)

echo "go get -u github.com/golang/protobuf/protoc-gen-go" >&2
go get -u github.com/golang/protobuf/protoc-gen-go || exit 1
echo "go generate github.com/djthorpe/remotes/protobuf"
go generate -x github.com/djthorpe/remotes/protobuf || exit 1
for COMMAND in ${COMMANDS[@]}; do
  echo "go install cmd/${COMMAND}"
  go install -ldflags "${LDFLAGS}" "cmd/${COMMAND}" || exit -1
done
