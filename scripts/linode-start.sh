#!/usr/bin/env bash
# Linode / Ubuntu: start an already-built Zion English admin web server (no git pull, no compile).
#
# Prerequisites: Go, .env, binary under ./tmp/ from deploy (zion-english-admin-prod or -dev).
# Default clone path: /root/zion-english-admin (prod) or /root/zion-english-dev (dev).
#
# Optional .env for prod TLS:
#   ZION_WEB_HTTPS=1
#   ZION_TLS_ADDRESS=your.domain.example   (Let's Encrypt cert path uses this hostname)
# If ZION_TLS_ADDRESS is unset in prod, defaults to flamendless.xyz (matches sshdeploy prod).
#
# Usage:
#   ./scripts/linode-start.sh [--prod|--dev] start
#
# systemd:
#   cp scripts/systemd/zion-english.service.example /etc/systemd/system/zion-english.service
#   systemctl daemon-reload && systemctl enable --now zion-english

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
export PATH="${HOME}/go/bin:${PATH}"

TIER=prod
CMD=start

usage() {
	cat <<'USAGE'
Usage: ./scripts/linode-start.sh [--prod|--dev] {start|stop|status|preflight|foreground}

  start       migrate, stop old process, nohup web server (logs: ./out)
  stop        stop matching web process
  status      show whether server is running
  preflight   migrations only; verify binary exists
  foreground  exec web server in foreground (for systemd ExecStart)

Prod: ./tmp/zion-english-admin-prod web -p PORT [--https --address DOMAIN]
Dev:  ./tmp/zion-english-admin-dev web -p PORT -b zion-english-admin
USAGE
}

load_env() {
	if [[ -f .env ]]; then
		set -a
		# shellcheck disable=SC1091
		source .env
		set +a
	fi
	PORT="${PORT:-8080}"
	GOOSE_DRIVER="${GOOSE_DRIVER:-sqlite3}"
	GOOSE_DBSTRING="${GOOSE_DBSTRING:-./data/zion.db}"
	GOOSE_MIGRATION_DIR="${GOOSE_MIGRATION_DIR:-./migrations/sqlite3}"
}

resolve_tier() {
	case "$TIER" in
	prod)
		BIN="${ROOT}/tmp/zion-english-admin-prod"
		PROCESS_PATTERN="${BIN} web"
		;;
	dev)
		BIN="${ROOT}/tmp/zion-english-admin-dev"
		PROCESS_PATTERN="${BIN} web"
		;;
	*)
		echo "Error: invalid tier ${TIER}" >&2
		exit 1
		;;
	esac
}

web_args() {
	local -a args
	args=(web -p "$PORT")
	case "$TIER" in
	prod)
		local https="${ZION_WEB_HTTPS:-1}"
		local tls_addr="${ZION_TLS_ADDRESS:-flamendless.xyz}"
		if [[ "$https" == "1" || "$https" == "true" || "$https" == "yes" ]]; then
			if [[ -z "${tls_addr}" ]]; then
				echo "Error: ZION_TLS_ADDRESS is required when ZION_WEB_HTTPS is enabled." >&2
				exit 1
			fi
			args+=(--https --address "$tls_addr")
		fi
		;;
	dev)
		args+=(-b zion-english-admin)
		;;
	esac
	WEB_ARGS=("${args[@]}")
}

run_migrations() {
	export GOOSE_DRIVER GOOSE_DBSTRING GOOSE_MIGRATION_DIR
	echo "Running migrations (goose up)..."
	go tool goose up
}

require_binary() {
	if [[ ! -x "$BIN" ]]; then
		echo "Error: missing executable ${BIN}. Build on server via sshdeploy or go build -o ${BIN} ." >&2
		exit 1
	fi
}

is_running() {
	pgrep -af "$PROCESS_PATTERN" >/dev/null 2>&1
}

do_stop() {
	if is_running; then
		pkill -f "$PROCESS_PATTERN" || true
		sleep 1
	fi
	if is_running; then
		echo "Warning: process may still be running." >&2
		return 1
	fi
	echo "Stopped (or no process was running)."
}

do_status() {
	load_env
	resolve_tier
	if is_running; then
		echo "Running (${TIER}, port ${PORT}):"
		pgrep -af "$PROCESS_PATTERN" || true
	else
		echo "Not running (${TIER})."
	fi
}

do_preflight() {
	load_env
	resolve_tier
	require_binary
	run_migrations
	echo "Preflight OK (${TIER})."
}

do_foreground() {
	load_env
	resolve_tier
	require_binary
	web_args
	exec "$BIN" "${WEB_ARGS[@]}"
}

do_start() {
	load_env
	resolve_tier
	require_binary
	run_migrations
	do_stop || true
	web_args
	echo "Starting: ${BIN} ${WEB_ARGS[*]}"
	nohup "$BIN" "${WEB_ARGS[@]}" >"${ROOT}/out" 2>&1 &
	disown
	sleep 1
	if is_running; then
		echo "Web server is running (${TIER}). Log: ${ROOT}/out"
	else
		echo "Warning: server may not have started. Check ${ROOT}/out" >&2
		exit 1
	fi
}

while [[ $# -gt 0 ]]; do
	case "$1" in
	--prod | -p)
		TIER=prod
		;;
	--dev | -d)
		TIER=dev
		;;
	-h | --help)
		usage
		exit 0
		;;
	start | stop | status | preflight | foreground)
		CMD="$1"
		;;
	*)
		echo "Error: unknown argument '$1'." >&2
		usage >&2
		exit 1
		;;
	esac
	shift
done

case "$CMD" in
start) do_start ;;
stop) load_env; resolve_tier; do_stop ;;
status) do_status ;;
preflight) do_preflight ;;
foreground) do_foreground ;;
*)
	usage >&2
	exit 1
	;;
esac
