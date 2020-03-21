#!/bin/bash
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.
#

#
# Copyright 2020 Zachary Schneider
#

set -o pipefail

DOWNLOADS="https://zps.io/downloads"

DOWNLOAD_PATH="${TMPDIR:-/tmp}"
EXTRACT_PATH=""

OS=""
ARCH=""
DOWNLOAD_FILE=""

log_error() {
  local message=$1

  if [[ $TERM == *"color"* || $TERM == "screen" ]]; then
    echo -e "\e[31m$message\e[0m"
  else
    echo "$message"
  fi
}

log_info() {
  local message=$1

  if [[ $TERM == *"color"* || $TERM == "screen" ]]; then
    echo -e "\e[32m$message\e[0m"
  else
    echo "$message"
  fi
}

detect() {
  OS=$(uname -s | awk '{print tolower($0)}')
  ARCH=$(uname -m| awk '{print tolower($0)}')

  if [[ "$OS" != "darwin" && "$OS" != "linux" ]]; then
    log_error "Unsupported OS: ${OS}"
    exit 1
  fi

  if [[ "$ARCH" != "x86_64" ]]; then
    log_error "Unsupported ARCH: ${ARCH}"
    exit 1
  fi
}

download() {
  DOWNLOAD_FILE="zps-${OS}-${ARCH}.tar.gz"
  EXTRACT_PATH=$(mktemp -d)

  if ! curl "${DOWNLOADS}/${DOWNLOAD_FILE}" -o "${DOWNLOAD_PATH}/${DOWNLOAD_FILE}"; then
    log_error "Failed to download: ${DOWNLOADS}/${DOWNLOAD_FILE}"
  fi

  if ! tar -C "$EXTRACT_PATH" -zxf "${DOWNLOAD_PATH}/${DOWNLOAD_FILE}"; then
    log_error "Failed to extract: ${DOWNLOAD_PATH}/${DOWNLOAD_FILE}"
  fi
}

install() {
  "${EXTRACT_PATH}/usr/bin/zps" pki trust import --type ca "${EXTRACT_PATH}/usr/share/zps/certs/zps.io/ca.pem"
  "${EXTRACT_PATH}/usr/bin/zps" pki trust import --type intermediate "${EXTRACT_PATH}/usr/share/zps/certs/zps.io/intermediate.pem"
  "${EXTRACT_PATH}/usr/bin/zps" pki trust import --type user "${EXTRACT_PATH}/usr/share/zps/certs/zps.io/zps.pem"

  "${EXTRACT_PATH}/usr/bin/zps" refresh

  "${EXTRACT_PATH}/usr/bin/zps" image init --helper "${IMAGE_PATH}/default"
}

cleanup() {
  log_info "Cleaning up ..."
  rm -rf "${EXTRACT_PATH}"
  rm -f "${DOWNLOAD_PATH}/${DOWNLOAD_FILE}"

  log_info "ZPS Installed!"
  log_info "Please add .zps/init.sh to your shell profile"
}

IMAGE_PATH="$1"

if [[ -z "$IMAGE_PATH" ]]; then
  log_error "Please provide a path in which to store your ZPS images, eg: /Users/username/images"
  exit 1
fi

trap cleanup EXIT

log_info "Detecting OS and ARCH"
detect

log_info "Dowloading ZPS"
download

log_info "Installing ZPS into new default image in ${IMAGE_PATH}/default"
install