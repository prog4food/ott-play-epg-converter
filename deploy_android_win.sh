#!/bin/sh
BIN_NAME="ofoss-epg-conv"
BIN_DIR="build"

PATH=`cygpath $ANDROID_SDK_ROOT`/ndk/25.0.8775105/toolchains/llvm/prebuilt/windows-x86_64/bin:$PATH

go_compile() {
  echo "[.] Compiling for ${1}:${2}..."
  FN="${BIN_DIR}/${BIN_NAME}_${1}_${5}"
  GOOS=$1 GOARCH=$2 CC=$3 go build -ldflags "-s -w -X main.depl_ver=$DEPL_VER" $4 -o "${FN}_${DEPL_VER}${6}" && \
   gzip -c "${FN}_${DEPL_VER}${6}" > "${FN}.gz" && \
   rm "${FN}_${DEPL_VER}${6}"
}
export CGO_ENABLED=1

[ -d "$BIN_DIR" ] && mkdir -p "$BIN_DIR"
rm -f ./$BIN_DIR/${BIN_NAME}_android_*

# Get versions 
GIT_COMM="`git log -n1 --pretty='%h'`"
GIT_TAG="`git describe --exact-match --tags $GIT_COMM`"
[ -z $GIT_TAG ] && GIT_TAG="dev"
DEPL_VER="${GIT_TAG}-${GIT_COMM}"
echo "[i] Version: $DEPL_VER"

# Build releases: android
go_compile android arm64 "aarch64-linux-android21-clang" "" "arm64"
export GOARM=7
go_compile android arm "armv7a-linux-androideabi21-clang" "" "armv7a"