#!/bin/bash

# inspired by https://github.com/envoyproxy/go-control-plane/blob/master/build/generate_protos.sh

set -e
set -x

basedir="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../api && pwd )"
outputdir=$basedir/external/
protodir=$basedir/external/
envoyprotodir=$protodir/envoy

# TODO once we move to go mod this can be $(go env GOMOD)/api/external
gobasepkg=github.com/solo-io/solo-kit/api/external

protoc=$(which protoc)

cd $basedir
mappings=(
  "google/api/annotations.proto=github.com/gogo/googleapis/google/api"
  "google/api/http.proto=github.com/gogo/googleapis/google/api"
  "google/rpc/code.proto=github.com/gogo/googleapis/google/rpc"
  "google/rpc/error_details.proto=github.com/gogo/googleapis/google/rpc"
  "google/rpc/status.proto=github.com/gogo/googleapis/google/rpc"
  "google/protobuf/any.proto=github.com/gogo/protobuf/types"
  "google/protobuf/duration.proto=github.com/gogo/protobuf/types"
  "google/protobuf/empty.proto=github.com/gogo/protobuf/types"
  "google/protobuf/struct.proto=github.com/gogo/protobuf/types"
  "google/protobuf/timestamp.proto=github.com/gogo/protobuf/types"
  "google/protobuf/wrappers.proto=github.com/gogo/protobuf/types"
  "gogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto"
#  "trace.proto=istio.io/gogo-genproto/opencensus/proto/trace"
#  "metrics.proto=istio.io/gogo-genproto/prometheus"
)

# assign importmap for canonical protos
for mapping in "${mappings[@]}"
do
  gogoarg+=",M$mapping"
done


for path in $(find ${envoyprotodir} -type d)
do
  if compgen -G ${path}/*.proto
  then
    path_protos=(${path}/*.proto)
    for path_proto in "${path_protos[@]}"
    do
      mapping=${path_proto##${protodir}/}=${gobasepkg}/${path##${protodir}/}
      gogoarg+=",M$mapping"
    done
  fi
done

for path in $(find ${envoyprotodir} -type d)
do
  if compgen -G ${path}/*.proto
  then
    path_protos=(${path}/*.proto)
    echo "Generating protos ${path} ..."
    $protoc --proto_path=${protodir} ${path}/*.proto \
      --gogo_out=${gogoarg}:$outputdir
  fi
done