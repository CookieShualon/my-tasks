#!/usr/bin/env sh
set -eu

APP_NAME="my-tasks"
INSTALL_DIR="${MY_TASKS_INSTALL_DIR:-${GOBIN:-$HOME/.local/bin}}"

cd "$(dirname "$0")"

if ! command -v go >/dev/null 2>&1; then
	printf '%s\n' "error: Go is required to install $APP_NAME." >&2
	exit 1
fi

mkdir -p "$INSTALL_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

printf '%s\n' "Building $APP_NAME..."
go build -o "$TMP_DIR/$APP_NAME" .

printf '%s\n' "Installing $APP_NAME to $INSTALL_DIR/$APP_NAME..."
install -m 0755 "$TMP_DIR/$APP_NAME" "$INSTALL_DIR/$APP_NAME"

case ":$PATH:" in
	*":$INSTALL_DIR:"*)
		printf '%s\n' "Installed. Run it with: $APP_NAME"
		;;
	*)
		printf '%s\n' "Installed, but $INSTALL_DIR is not on your PATH."
		printf '%s\n' "Add this to your shell config, then restart your terminal:"
		printf '%s\n' "  export PATH=\"$INSTALL_DIR:\$PATH\""
		printf '%s\n' "Then run it with: $APP_NAME"
		;;
esac
