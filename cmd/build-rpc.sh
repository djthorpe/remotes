#!/bin/bash
##############################################################
# Build RPC binaries
##############################################################

CURRENT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GO=`which go`
PROTOC=`which protoc`
PROTOC_GEN_GO=`which protoc-gen-go`
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
if [ ! -x "${PROTOC}" ] ; then
  echo "protoc not installed or executable" >&2
  exit -1
fi

##############################################################
# gRPC Plugin compile

if [ ! -x "${PROTOC_GEN_GO}" ]; then
  echo "go get -u github.com/golang/protobuf/protoc-gen-go" >&2
  go get -u google.golang.org/grpc || exit 1
  go get -u github.com/golang/protobuf/protoc-gen-go || exit 1
fi

##############################################################
# Install RPC binaries

COMMANDS=(
  rpc/remotes-service.go
  rpc/remotes-client.go
)

echo "go generate github.com/djthorpe/remotes/rpc/protobuf"
go generate -x github.com/djthorpe/remotes/rpc/protobuf || exit 1
for COMMAND in ${COMMANDS[@]}; do
  echo "go install cmd/${COMMAND}"
  go install -ldflags "${LDFLAGS}" "cmd/${COMMAND}" || exit -1
done
