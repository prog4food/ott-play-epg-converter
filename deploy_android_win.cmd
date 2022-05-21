@echo off
set PATH=%ANDROID_SDK_ROOT%\ndk\21.4.7075529\toolchains\llvm\prebuilt\windows-x86_64\bin;%PATH%
set go_compile=go build -ldflags "-s -w" --tags android

sh -c "rm -f ./build/ott-play-epg-converter_android_*"

rem public env
set CGO_ENABLED=1
set GOOS=android

rem aarch64: platform env
set GOARCH=arm64
set CC=aarch64-linux-android21-clang
%go_compile% -o .\build\ott-play-epg-converter_android_arm64

rem armv7a: platform env
set GOARM=7
set GOARCH=arm
set CC=armv7a-linux-androideabi21-clang
%go_compile% -o .\build\ott-play-epg-converter_android_armv7a

rem finish
sh -c "gzip ./build/ott-play-epg-converter_android_*"
exit /b
