#!/usr/bin/env bash

# @Brandon Blanker Lim-it

set -eufx -o pipefail

TMP="./tmp"

BROWSER="${BROWSER:-vivaldi}"
ISWSL=false
ISMAC=false

if [[ $(uname) == "Darwin" ]]; then
	ISMAC=true
elif grep -qi Microsoft /proc/version; then
	ISWSL=true
fi

load_env() {
	if [ -f .env ]; then
		set -a
		# shellcheck disable=SC1091
		source .env
		set +a
	fi
	PORT="${PORT:-8080}"
}

serve() {
	load_env
	gensql
	gentempl
	local -; set -x;

	if "${ISWSL}"; then
		cmd.exe /c "start ${BROWSER} http://localhost:${PORT}/zion-english-admin"
	elif "${ISMAC}"; then
		open -a ${BROWSER} "http://localhost:${PORT}/zion-english-admin"
	fi

	go run . web
}

prod() {
	load_env
	echo "running in production server"
	gensql
	gentempl

	if "${ISMAC}"; then
		PIDS=$(lsof -ti:"${PORT}" 2>/dev/null || true)
		if [ -n "$PIDS" ]; then
			echo "$PIDS" | xargs kill -9 2>/dev/null || true
		fi
	elif "${ISWSL}"; then
		fuser -k "${PORT}"/tcp 2>/dev/null || true
	else
		fuser -k "${PORT}"/tcp 2>/dev/null || \
		(netstat -tlnp 2>/dev/null | grep ":${PORT}" | awk '{print $7}' | cut -d'/' -f1 | xargs -r kill -9) || true
	fi

	sleep 1

	echo "Please run:"
	echo "    go run . web -p ${PORT} --https --address flamendless.xyz > outlog 2>&1 &"

}

gentempl() {
	go tool templ generate templ -v
}

gensql() {
	go tool sqlc generate
}

test() {
	go test ./...
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0}"
	echo "Commands:"
	echo "    serve"
	echo "    prod"
	echo "    gentempl"
	echo "    gensql"
	echo "    test"
else
	echo "Running ${1}"
	time "$1" "$@"
fi
