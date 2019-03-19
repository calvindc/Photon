#!/bin/sh
export CGO_ENABLED=0
export GIT_COMMIT=`git rev-list -1 HEAD`
export GO_VERSION=`go version|sed 's/ //g'`
export BUILD_DATE=`date "+%Y-%m-%d-%H:%M:%S"`
export VERSION=1.1.0--${GIT_COMMIT:0-40:4}
echo $GIT_COMMIT

go  build  -ldflags "   -X github.com/SmartMeshFoundation/Photon/cmd/photon/mainimpl.GitCommit=$GIT_COMMIT -X github.com/SmartMeshFoundation/Photon/cmd/photon/mainimpl.GoVersion=$GO_VERSION -X github.com/SmartMeshFoundation/Photon/cmd/photon/mainimpl.BuildDate=$BUILD_DATE -X github.com/SmartMeshFoundation/Photon/cmd/photon/mainimpl.Version=$VERSION "

cp photon $GOPATH/bin
