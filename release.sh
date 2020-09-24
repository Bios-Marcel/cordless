#!/bin/bash

#
# This tiny script helps me not to mess up the procedure of releasing a new
# version of cordless.
#
# Dependencies:
#   * sha256sum
#   * envsubst
#   * git
#   * date
#   * go
#   * xclip
#

#
# This will cause a script failure in case any of the commands below fail.
#

set -e

#
# RELEASE_DATE is the current date in format Year-Month-Day, since cordless
# uses dates for versioning.
#

RELEASE_DATE="$(date +%Y-%m-%d)"
export RELEASE_DATE

#
# Setting the cordless version-number to the current release date.
# This number can be requested on startup and is only used for that purpose.
#
# This has to happen before building, since version numbers are incorrect otherwise.
#

envsubst < version.go_template > version/version.go
git commit version/version.go -m "Bump cordless versionnumber to version $RELEASE_DATE"

#
# Define executable names for later usage.
#

BIN_LINUX="cordless_linux_64"
BIN_DARWIN="cordless_darwin"
BIN_WINDOWS_64="cordless_64.exe"
BIN_WINDOWS_32="cordless_32.exe"

#
# Building cordless for darwin, linux and windows.
#
# The binaries are built without debug symbols in order to save filesize.
#

GOOS=linux go build -o $BIN_LINUX -ldflags="-s -w"
GOOS=darwin go build -o $BIN_DARWIN -ldflags="-s -w"
GOOS=windows go build -o $BIN_WINDOWS_64 -ldflags="-s -w"
GOOS=windows GOARCH=386 go build -o $BIN_WINDOWS_32 -ldflags="-s -w"

#
# EXE_HASH is sha256 of the previously built cordless.exe and is required
# for scoop to properly work.
#
# Since envsubst can not see the unexported variables, we export them here
# and unexport them at a later point and time.
#

EXE_64_HASH="$(sha256sum ./$BIN_WINDOWS_64 | cut -f 1 -d " ")"
export EXE_64_HASH
EXE_32_HASH="$(sha256sum ./$BIN_WINDOWS_32 | cut -f 1 -d " ")"
export EXE_32_HASH

#
# Substituting the variables in the scoop manifest template into the actual
# manifest.
#

envsubst < cordless.json_template > cordless.json

#
# Commit and push the new scoop manifest.
#

git commit cordless.json -m "Bump scoop package to version $RELEASE_DATE"
git push

#
# Create a new tag and push it.
#

git tag -s "$RELEASE_DATE" -m "Update scoop package to version ${RELEASE_DATE}"
git push --tags

#
# Copies the changelog for pasting into the github release. The changes will
# include all commits between the latest and the previous tag.
#

RELEASE_BODY="$(git log --pretty=oneline --abbrev-commit "$(git describe --abbrev=0 "$(git describe --abbrev=0)"^)".."$(git describe --abbrev=0)")"

#
# Temporarily disable that the script exists on subcommand failure.
#

set +e

#
# Look up the release to create and save whether the lookup was successful.
# This has to be manually saved due to the fact that the next `set -e` would
# reset the `$?` variable.
#

hub release show "$RELEASE_DATE"
RELEASE_EXISTS=$?

#
# Let script exit again on subcommand failure.
#

set -e

#
# If the release already exists, we edit the existing one instead of creating a
# new one.
#

if [ $RELEASE_EXISTS -eq 0 ]
then
    hub release edit -a "$BIN_LINUX" -a "$BIN_DARWIN" -a "$BIN_WINDOWS_64" -a "$BIN_WINDOWS_32" -m "" -m "${RELEASE_BODY}" "$RELEASE_DATE"
else
    hub release create -a "$BIN_LINUX" -a "$BIN_DARWIN" -a "$BIN_WINDOWS_64" -a "$BIN_WINDOWS_32" -m "${RELEASE_DATE}" -m "${RELEASE_BODY}" "$RELEASE_DATE"
fi

#
# Substitutes the manifest template for the homebrew package. We need to
# download the latest tarball in order to get its sha256 sum.
#

wget https://github.com/Bios-Marcel/cordless/archive/$RELEASE_DATE.tar.gz
TAR_HASH="$(sha256sum ./$RELEASE_DATE.tar.gz | cut -f 1 -d " ")"
export TAR_HASH
rm ./$RELEASE_DATE.tar.gz
envsubst < cordless.rb_template > cordless.rb

#
# Unsetting (and unexporting) previously exported environment variables.
#

unset RELEASE_DATE
unset EXE_64_HASH
unset EXE_32_HASH
unset TAR_HASH

