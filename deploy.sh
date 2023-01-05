#!/bin/sh
GO_URL="https://go.dev/dl/go1.19.4.linux-amd64.tar.gz"
BIN_NAME="ofoss-epg-conv"
BIN_DIR="build"

ENV_REQ="gcc x86_64-w64-mingw32-gcc i686-w64-mingw32-gcc aarch64-linux-gnu-gcc arm-linux-gnueabi-gcc curl git"

which $ENV_REQ > /dev/null || {
  echo "[.] Installing DEV environment..."
  apt-get update
  apt-get install -y -qq git curl build-essential gcc-mingw-w64 gcc-aarch64-linux-gnu gcc-arm-linux-gnueabi gcc-arm-linux-gnueabihf
}

export GOROOT=/go
export GOPATH=/home/go
export GOBIN=${GOPATH}/bin
export GOCACHE=${GOPATH}/.cache
export PATH=${GOROOT}/bin:$GOBIN:$PATH

# Download golang
[ -f "$GOROOT/bin/go" ] || {
  echo "[.] Installing GO environment..."
  mkdir $GOROOT $GOPATH
  curl -sL "$GO_URL" | tar zxf - -C /
}

go_compile() {
  echo "[.] Compiling for ${1}:${5}..."
  FN="${BIN_DIR}/${BIN_NAME}_${1}_${5}"
  GOOS=$1 GOARCH=$2 CC=$3 go build -ldflags "-s -w -X main.depl_ver=$DEPL_VER" $4 -o "${FN}_${DEPL_VER}${6}" && \
   gzip -c "${FN}_${DEPL_VER}${6}" > "${FN}.gz" && \
   rm "${FN}_${DEPL_VER}${6}"
}
export CGO_ENABLED=1

[ -d "$BIN_DIR" ] && mkdir -p "$BIN_DIR"
rm -f ./$BIN_DIR/${BIN_NAME}_windows_* ./$BIN_DIR/${BIN_NAME}_linux_*
go mod download

# Get versions 
GIT_COMM="`git log -n1 --pretty='%h'`"
GIT_TAG="`git describe --exact-match --tags $GIT_COMM`"
[ -z $GIT_TAG ] && GIT_TAG="dev"
DEPL_VER="${GIT_TAG}-${GIT_COMM}"
echo "[i] Version: $DEPL_VER"

# Build releases: Windows
go_compile windows 386 "i686-w64-mingw32-gcc" "" "x86" ".exe"
go_compile windows amd64 "x86_64-w64-mingw32-gcc" "" "x64" ".exe"
# Build releases: Linux
go_compile linux amd64 "gcc" "" "x64"
go_compile linux arm64 "aarch64-linux-gnu-gcc" "" "arm64"
export GOARM=7
go_compile linux arm "arm-linux-gnueabi-gcc -march=armv7-a  -mfpu=neon-vfpv3"  "" "armv7a"
go_compile linux arm "arm-linux-gnueabihf-gcc -march=armv7-a -mfpu=neon-vfpv3"  "" "armv7a-hf"
