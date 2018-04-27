#!/bin/bash
##############################################################
# Build command-line binaries
##############################################################

CURRENT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GO=`which go`
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
# Install

COMMANDS=(
  tool/ir_rcv.go
  tool/ir_learn.go
  tool/ir_send.go
)

for COMMAND in ${COMMANDS[@]}; do
  echo "go install cmd/${COMMAND}"
  go install -ldflags "${LDFLAGS}" "cmd/${COMMAND}" || exit -1
done
