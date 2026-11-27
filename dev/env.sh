# shellcheck shell=bash
# Sourced by bin/dev. Gives every checkout (main repo or git worktree) its own
# "slot" so several dev stacks can run side by side without colliding:
#
#   slot 0 = main checkout (default ports), slot N = a worktree (ports + N*100)
#
# Slots are recorded in $DEV_STATE/slots ("<slot> <path>" per line) and reused
# by the same path; entries for deleted worktrees are pruned automatically.

DEV_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export DEV_STATE="${XDG_STATE_HOME:-$HOME/.local/state}/simple-track-dev"
mkdir -p "$DEV_STATE" "${XDG_CONFIG_HOME:-$HOME/.config}/process-compose"  # latter silences a process-compose debug warning

_dev_slot() {
  local root="$1" git_dir common_dir
  git_dir="$(git -C "$root" rev-parse --absolute-git-dir)"
  common_dir="$(cd "$root" && cd "$(git rev-parse --git-common-dir)" && pwd)"
  if [[ "$git_dir" == "$common_dir" ]]; then
    echo 0
    return
  fi
  (
    flock 9
    local file="$DEV_STATE/slots" tmp slot path n
    touch "$file"
    tmp="$(mktemp)"
    while read -r slot path; do
      [[ -d "$path" ]] && echo "$slot $path"
    done < "$file" > "$tmp"
    mv "$tmp" "$file"
    slot="$(awk -v p="$root" '$2 == p { print $1 }' "$file")"
    if [[ -z "$slot" ]]; then
      n=1
      while awk -v s="$n" '$1 == s { found = 1 } END { exit !found }' "$file"; do n=$((n + 1)); done
      slot=$n
      echo "$slot $root" >> "$file"
      # A freed slot number may still have the previous worktree's database;
      # drop it so this worktree starts empty and applies migrations cleanly.
      "$DEV_ROOT/dev/purge-slot" "$slot" >&2
    fi
    echo "$slot"
  ) 9> "$DEV_STATE/slots.lock"
}

export DEV_SLOT="$(_dev_slot "$ROOT")"
_off=$((DEV_SLOT * 100))

export DEV_FRONTEND_PORT=$((3000 + _off))
export DEV_BACKEND_PORT=$((8080 + _off))
export DEV_PGWEB_PORT=$((8081 + _off))
export DEV_PG_PORT=$((5432 + _off))
export DEV_FRONTEND_URL="http://localhost:$DEV_FRONTEND_PORT"
export DEV_BACKEND_URL="http://127.0.0.1:$DEV_BACKEND_PORT"
# Shared by every slot: `dev phone` serves whichever checkout last ran it on this fixed port.
export DEV_PHONE_PORT="${DEV_PHONE_PORT:-2999}"

# Separate docker compose project => separate postgres container + volume.
if [[ "$DEV_SLOT" == 0 ]]; then
  export COMPOSE_PROJECT_NAME="simple-track"
else
  export COMPOSE_PROJECT_NAME="simple-track-wt$DEV_SLOT"
fi

# process-compose control API: per-slot unix socket (its TCP default is 8080).
export PC_SOCKET_PATH="${XDG_RUNTIME_DIR:-/tmp}/simple-track-dev-$DEV_SLOT.sock"
export PC_LOG_FILE="$DEV_STATE/process-compose-$DEV_SLOT.log"
# Don't let process-compose inject the root .env (docker compose vars) into processes.
export PC_DISABLE_DOTENV=1

# Backend (env vars take precedence over backend/.env).
export LISTEN_ADDR="0.0.0.0:$DEV_BACKEND_PORT"
export DATABASE_URL="postgresql://postgres:postgres@localhost:$DEV_PG_PORT/simpletracker_dev?sslmode=disable"
# Also allow this machine's LAN addresses (phone testing via http://<lan-ip>:<port>).
# Override with DEV_LAN_HOST="ip1 ip2" (space separated); default is every private IPv4.
# The phone proxy's origins (http://<host>.local:$DEV_PHONE_PORT and the LAN IPs) are allowed too.
CORS_ORIGINS="http://localhost:$DEV_FRONTEND_PORT,http://127.0.0.1:$DEV_FRONTEND_PORT,http://$(hostname).local:$DEV_PHONE_PORT"
for _ip in ${DEV_LAN_HOST:-$(hostname -I 2>/dev/null)}; do
  case "$_ip" in
    10.*|192.168.*|172.1[6-9].*|172.2?.*|172.3[01].*)
      CORS_ORIGINS="$CORS_ORIGINS,http://$_ip:$DEV_FRONTEND_PORT,http://$_ip:$DEV_PHONE_PORT" ;;
  esac
done
unset _ip
export CORS_ORIGINS
export ALLOW_BYPASS=true
# One JWT secret for every slot, so a login survives `dev phone` switching checkouts behind the
# same origin (each backend/.env has its own random one, which this overrides).
_jwt="$DEV_STATE/jwt-secret"
[[ -s "$_jwt" ]] || (umask 077 && openssl rand -hex 32 > "$_jwt")
export JWT_SECRET="$(cat "$_jwt")"
unset _jwt
export DEV_AUTH_USER=dev

# Frontend: vite proxies /api to this slot's backend.
export VITE_PROXY_TARGET="$DEV_BACKEND_URL"

unset _off
