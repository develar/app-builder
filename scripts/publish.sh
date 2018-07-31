#!/usr/bin/env bash
set -e

BASEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "$GITHUB_TOKEN" ] ; then
  SEC=`security find-generic-password -l GH_TOKEN -g 2>&1`
  export GITHUB_TOKEN=`echo "$SEC" | grep "password" | cut -d \" -f 2`
fi

NAME=app-builder
VERSION=2.1.1

OUT_DIR="$BASEDIR/../dist"

publish()
{
  outDir=$1
  archiveName="$NAME-v$VERSION-$2"
  archiveFile="$OUT_DIR/$archiveName.7z"

  cd "$OUT_DIR/$outDir"
  7za a -mx=9 -mfb=64 "$archiveFile" .
}

publish "darwin_amd64" mac

publish "linux_386" linux-ia32
publish "linux_amd64" linux-x64
publish "linux_arm_7" linux-armv7
publish "linux_arm64" linux-armv8

publish "windows_386" win-ia32
publish "windows_amd64" win-x64

tool-releaser develar/app-builder "v$VERSION" master "" "$OUT_DIR/*.7z"