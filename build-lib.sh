#!/usr/bin/env bash
# Build the libbox AAR from the (modified) Go sources and drop it into the app.
# Run this only when the GO code changes. Slow on the very first run; fast
# afterwards as long as the Go build-cache volume (GOCACHE) is persisted.
set -euxo pipefail

make lib_android

mkdir -p clients/android/app/libs
cp ./*.aar clients/android/app/libs/
