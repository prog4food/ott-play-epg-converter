#!/bin/sh
GO_URL="https://go.dev/dl/go1.18.1.linux-amd64.tar.gz"
BIN_NAME="ott-play-epg-converter"
BIN_DIR="build"

ENV_REQ="gcc x86_64-w64-mingw32-gcc i686-w64-mingw32-gcc aarch64-linux-gnu-gcc curl git"

which $ENV_REQ > /dev/null || {
  echo "[.] Installing DEV environment..."
  apt-get update
  apt-get install -y -qq git curl build-essential gcc-mingw-w64 gcc-aarch64-linux-gnu
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
  OEXT=".bin"
  [ "$1" = "windows" ] && OEXT=".exe"
  GOOS=$1 GOARCH=$2 CC=$3 go build -ldflags "-s -w" -o "${BIN_DIR}/${BIN_NAME}_${1}_${2}${OEXT}"
}
export CGO_ENABLED=1

[ -d "$BIN_DIR" ] && mkdir -p "$BIN_DIR"
rm -f ./$BIN_DIR/*
go mod download

# Build releases
go_compile windows 386 i686-w64-mingw32-gcc
go_compile windows amd64 x86_64-w64-mingw32-gcc
go_compile linux amd64 gcc
go_compile linux arm64 aarch64-linux-gnu-gcc

# Compress releases
gzip -f ./$BIN_DIR/*
