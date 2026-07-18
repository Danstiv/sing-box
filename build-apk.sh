#!/usr/bin/env bash
# Assemble the debug APK. Assumes the libbox AAR is already in
# clients/android/app/libs — run build-lib.sh first whenever the Go code changed.
# For Android-only changes (Kotlin / resources) just run this; Gradle stays
# incremental because the AAR is not regenerated.
set -euxo pipefail

if ! ls clients/android/app/libs/*.aar >/dev/null 2>&1; then
    echo "no AAR in clients/android/app/libs — run build-lib.sh first" >&2
    exit 1
fi

cd clients/android
./gradlew --no-daemon :app:assembleOtherDebug

echo "=== built APK(s) ==="
find app/build/outputs/apk -name '*.apk'
