#!/usr/bin/env bash

# @Brandon Blanker Lim-it

set -euf -o pipefail

TMP="./tmp"

BROWSER="${BROWSER:-vivaldi}"
ISWSL=false
ISMAC=false

if [[ $(uname) == "Darwin" ]]; then
	ISMAC=true
elif grep -qi Microsoft /proc/version; then
	ISWSL=true
fi

serve() {
	local -; set -x;

	if "${ISWSL}"; then
		cmd.exe /c "start ${BROWSER} http://localhost:8080/zion-english-admin"
	elif "${ISMAC}"; then
		open -a ${BROWSER} "http://localhost:8080/zion-english-admin"
	fi

	go run . web
}

prod() {
	echo "running in production server"

	if "${ISMAC}"; then
		PIDS=$(lsof -ti:1010 2>/dev/null || true)
		if [ -n "$PIDS" ]; then
			echo "$PIDS" | xargs kill -9 2>/dev/null || true
		fi
	elif "${ISWSL}"; then
		fuser -k 1010/tcp 2>/dev/null || true
	else
		fuser -k 1010/tcp 2>/dev/null || \
		(netstat -tlnp 2>/dev/null | grep :1010 | awk '{print $7}' | cut -d'/' -f1 | xargs -r kill -9) || true
	fi

	sleep 1

	echo "go run . web -p 1010 --https --address flamendless.xyz > outlog 2>&1"
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0}"
	echo "Commands:"
	echo "    serve"
	echo "    prod"
else
	echo "Running ${1}"
	time "$1" "$@"
fi
