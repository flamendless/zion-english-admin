#!/usr/bin/env bash

set -euf -o pipefail

usage() {
	cat <<'USAGE'
Usage: SSH_ADDR=user@host ./scripts/sshdeploy.sh --dev|--prod

Environment variables:
  SSH_ADDR    Remote SSH address (required), e.g. user@example.com

Flags:
  --dev, -d   Deploy to ~/zion-english-admin-dev
  --prod, -p  Deploy to ~/zion-english-admin
USAGE
}

SSH_ADDR="${SSH_ADDR:-}"
if [ -z "$SSH_ADDR" ]; then
	echo "Error: SSH_ADDR environment variable is required." >&2
	usage >&2
	exit 1
fi

ENV=""
while [ $# -gt 0 ]; do
	case "$1" in
	--dev | -d)
		ENV="dev"
		;;
	--prod | -p)
		ENV="prod"
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "Error: unknown argument '$1'." >&2
		usage >&2
		exit 1
		;;
	esac
	shift
done

if [ -z "$ENV" ]; then
	echo "Error: --dev or --prod flag is required." >&2
	usage >&2
	exit 1
fi

read -rsp "SSH password: " SSH_PASSWORD
echo

ssh_cmd() {
	if command -v sshpass >/dev/null 2>&1; then
		SSHPASS="$SSH_PASSWORD" sshpass -e ssh "$@"
	else
		echo "sshpass not found; ssh will prompt for password." >&2
		ssh "$@"
	fi
}

echo "Deploying $ENV to $SSH_ADDR..."

ssh_cmd "$SSH_ADDR" env ENV="$ENV" bash --login -s <<'EOF'
set -euf -o pipefail

ENV="${ENV:?missing ENV}"
BIN_PREFIX="zion-english-admin"

if [ "$ENV" = "dev" ]; then
	PROJECT_DIR=~/zion-english-dev
	BIN="./tmp/${BIN_PREFIX}-dev"
	WEB_ARGS=(-b zion-english-admin)
else
	PROJECT_DIR=~/zion-english-admin
	BIN="./tmp/${BIN_PREFIX}-prod"
	WEB_ARGS=(--https --address flamendless.xyz)
fi

PROCESS_PATTERN="${BIN} web"

echo "Navigating to project directory..."
cd "$PROJECT_DIR"

echo "Syncing repository..."
git fetch --prune --tags origin

if ! git diff --quiet || ! git diff --cached --quiet; then
	echo "Stashing local changes..."
	git stash push -u -m "deploy stash $(date -u +%Y-%m-%dT%H:%M:%SZ)"
fi

if git show-ref --verify --quiet refs/heads/master; then
	git checkout master
elif git show-ref --verify --quiet refs/remotes/origin/master; then
	git checkout -B master origin/master
else
	echo "Error: origin/master not found." >&2
	exit 1
fi

git reset --hard origin/master

echo "Syncing Go modules..."
go mod download

echo "Installing/updating Go tools..."
go install tool

if [ -f .env ]; then
	set -a
	# shellcheck disable=SC1091
	source .env
	set +a
fi
PORT="${PORT:-8080}"
GOOSE_DRIVER="${GOOSE_DRIVER:-sqlite3}"
GOOSE_DBSTRING="${GOOSE_DBSTRING:-./data/zion.db}"
GOOSE_MIGRATION_DIR="${GOOSE_MIGRATION_DIR:-./migrations/sqlite3}"

echo "Stopping existing process..."
if pgrep -af "$PROCESS_PATTERN" >/dev/null 2>&1; then
	pkill -f "$PROCESS_PATTERN" || true
	sleep 1
else
	echo "No existing process found."
fi

echo "Running migrations..."
export GOOSE_DRIVER GOOSE_DBSTRING GOOSE_MIGRATION_DIR
go tool goose up

echo "Generating sqlc..."
./run.sh gensql

echo "Generating templ..."
./run.sh gentempl

mkdir -p ./tmp
echo "Building $BIN..."
go build -o "$BIN" .

RUN_CMD=("$BIN" web -p "$PORT" "${WEB_ARGS[@]}")
echo "Starting web server with: ${RUN_CMD[*]}"
nohup "${RUN_CMD[@]}" > out 2>&1 &
disown

sleep 1
if pgrep -af "$PROCESS_PATTERN" >/dev/null 2>&1; then
	echo "Web server process is running."
else
	echo "Warning: web server process may not have started. Check the out file." >&2
fi

echo "Done!"
EOF
