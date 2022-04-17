#!/bin/sh
GO_URL="https://go.dev/dl/go1.18.1.linux-amd64.tar.gz"
BIN_NAME="ott-play-epg-converter"
BIN_DIR="build"

ENV_REQ="gcc x86_64-w64-mingw32-gcc i686-w64-mingw32-gcc aarch64-linux-gnu-gcc arm-linux-gnueabi-gcc curl git"

which $ENV_REQ > /dev/null || {
  echo "[.] Installing DEV environment..."
  apt-get update
  apt-get install -y -qq git curl build-essential gcc-mingw-w64 gcc-aarch64-linux-gnu gcc-arm-linux-gnueabi gcc-arm-linux-gnueabihf
}

which go > /dev/null || {
  export GOROOT=/go
  export GOPATH=/home/go
  export GOBIN=${GOPATH}/bin
  export GOCACHE=${GOPATH}/.cache
  export PATH=${GOROOT}/bin:$GOBIN:$PATH
  which go > /dev/null || {
    echo "[.] Installing GO environment..."
    mkdir $GOROOT $GOPATH
    curl -sL "$GO_URL" | tar zxf - -C /
  }
}

go_compile() {
  echo "[.] Compiling for ${1}:${2}..."
  GOOS=$1 GOARCH=$2 CC=$3 go build -ldflags "-s -w" $4 -o "${BIN_DIR}/${BIN_NAME}_${1}_${5}"
}
export CGO_ENABLED=1

[ -d "$BIN_DIR" ] && mkdir -p "$BIN_DIR"
rm -f ./$BIN_DIR/${BIN_NAME}_windows_* ./$BIN_DIR/${BIN_NAME}_linux_*
go mod download

# Build releases: Windows
go_compile windows 386 "i686-w64-mingw32-gcc" "--tags windows" "x64.exe"
go_compile windows amd64 "x86_64-w64-mingw32-gcc" "--tags windows" "x86.exe"
# Build releases: Linux
go_compile linux amd64 "gcc" "--tags linux" "x64"
go_compile linux arm64 "aarch64-linux-gnu-gcc" "--tags linux" "arm64"
export GOARM=7
go_compile linux arm "arm-linux-gnueabi-gcc -march=armv7-a  -mfpu=neon-vfpv3"  "--tags linux" "armv7a"
go_compile linux arm "arm-linux-gnueabihf-gcc -march=armv7-a -mfpu=neon-vfpv3"  "--tags linux" "armv7a-hf"

# Compress releases
gzip -f ./$BIN_DIR/${BIN_NAME}_windows_* ./$BIN_DIR/${BIN_NAME}_linux_*
