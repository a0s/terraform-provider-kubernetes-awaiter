#!/usr/bin/env sh
set -e

PROVIDER_NAME="kubernetes-awaiter"
PROVIDER_FULL_NAME="terraform-provider-$PROVIDER_NAME"
PROVIDER_EXECUTABLE="$(pwd)/$PROVIDER_FULL_NAME"
PROVIDER_VERSION="$($PROVIDER_EXECUTABLE --version)"
OS_NAME="$(uname -s)"
MACHINE_HARDWARE="$(uname -m)"

if [ "$OS_NAME" = "Darwin" ] && [ "$MACHINE_HARDWARE" = "x86_64" ]; then
  TARGET="darwin_amd64"
  PROVIDER_FOLDER="$HOME/.terraform.d/plugins/localhost/a0s/$PROVIDER_NAME/$PROVIDER_VERSION/$TARGET"
  mkdir -p "$PROVIDER_FOLDER"
  ln -sf "$PROVIDER_EXECUTABLE" "$PROVIDER_FOLDER/$PROVIDER_FULL_NAME"
  chmod +x "$PROVIDER_FOLDER/$PROVIDER_FULL_NAME"

  echo "Driver: \n  $PROVIDER_FOLDER/$PROVIDER_FULL_NAME\n"
  cat << EOF
Config:
  terraform {
    required_providers {
      $PROVIDER_NAME = {
        version = "~> $PROVIDER_VERSION"
        source  = "localhost/a0s/$PROVIDER_NAME"
      }
    }
  }
EOF
fi
