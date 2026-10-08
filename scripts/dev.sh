#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

command -v python3 >/dev/null || { echo 'python3 is required for startup timing.' >&2; exit 1; }
command -v curl >/dev/null || { echo 'curl is required for readiness checks.' >&2; exit 1; }

clock_now_ns() {
  python3 -c 'import time; print(time.time_ns())'
}
format_timestamp() {
  python3 -c 'import datetime, sys; seconds, nanos = divmod(int(sys.argv[1]), 1_000_000_000); moment = datetime.datetime.fromtimestamp(seconds).astimezone(); print(moment.strftime("%Y-%m-%d %H:%M:%S.") + f"{nanos // 1_000_000:03d} " + moment.strftime("%Z"))' "$1"
}
elapsed_seconds() {
  python3 -c 'import sys; print(f"{(int(sys.argv[2]) - int(sys.argv[1])) / 1_000_000_000:.1f}")' "$1" "$2"
}

started_at=$(clock_now_ns)
startup_second=$SECONDS
echo "Startup began at: $(format_timestamp "$started_at")"

startup_timeout=${STARTUP_TIMEOUT_SECONDS:-90}
if [[ ! $startup_timeout =~ ^[0-9]+$ ]] || (( startup_timeout < 1 )); then
  echo 'STARTUP_TIMEOUT_SECONDS must be a positive integer.' >&2
  exit 1
fi

services=(adapters providers login products invoices shipping)
default_ports=(3006 3007 3001 3002 3005 3004)
ports=()
ready=()
port_for() {
  local service=$1 default_port=$2 key value env_file
  key="$(printf '%s' "$service" | tr '[:lower:]' '[:upper:]')_PORT"
  value="${!key:-}"
  env_file=${ENV_FILE:-.env}
  if [[ -z $value && -f $env_file ]]; then
    value=$(awk -F= -v key="$key" '$1 == key { gsub(/[[:space:]\047\042]/, "", $2); print $2; exit }' "$env_file")
  fi
  value=${value:-$default_port}
  if [[ ! $value =~ ^[0-9]+$ ]] || (( value < 1 || value > 65535 )); then
    echo "Invalid $key port: $value" >&2
    return 1
  fi
  printf '%s' "$value"
}
for index in "${!services[@]}"; do
  ports+=("$(port_for "${services[$index]}" "${default_ports[$index]}")")
  ready+=(0)
done

# Stop listeners before opening logs: a previous `make start` may still be alive.
command -v lsof >/dev/null || { echo 'lsof is required to free service ports before startup.' >&2; exit 1; }
old_pids=''
for index in "${!services[@]}"; do
  while IFS= read -r pid; do
    [[ $pid =~ ^[0-9]+$ ]] || continue
    if [[ " $old_pids " != *" $pid "* ]]; then
      old_pids="$old_pids $pid"
      echo "Stopping process $pid on port ${ports[$index]} (${services[$index]})"
    fi
  done < <(lsof -nP -t -iTCP:"${ports[$index]}" -sTCP:LISTEN 2>/dev/null || true)
done
for pid in $old_pids; do
  if kill -0 "$pid" 2>/dev/null && ! kill -TERM "$pid" 2>/dev/null; then
    echo "Cannot stop process $pid; check permissions." >&2
    exit 1
  fi
done

ports_busy() {
  local port
  for port in "${ports[@]}"; do
    if lsof -nP -t -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
      return 0
    fi
  done
  return 1
}
if [[ -n $old_pids ]]; then
  for (( attempt = 0; attempt < 50; attempt++ )); do
    ports_busy || break
    sleep 0.2
  done
  if ports_busy; then
    echo 'Existing listeners did not stop after 10 seconds; forcing them to exit.' >&2
    for pid in $old_pids; do
      kill -KILL "$pid" 2>/dev/null || true
    done
    for (( attempt = 0; attempt < 25; attempt++ )); do
      ports_busy || break
      sleep 0.2
    done
  fi
  if ports_busy; then
    echo 'Cannot free all service ports; startup cancelled.' >&2
    exit 1
  fi
fi

mkdir -p logs
pids=()
cleanup() {
  trap - EXIT INT TERM
  for pid in "${pids[@]}"; do kill "$pid" 2>/dev/null || true; done
  for pid in "${pids[@]}"; do wait "$pid" 2>/dev/null || true; done
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
for service in "${services[@]}"; do
  "./bin/$service" >"logs/$service.log" 2>&1 &
  pids+=("$!")
  launched_at=$(clock_now_ns)
  echo "Started $service (PID $!, launch +$(elapsed_seconds "$started_at" "$launched_at") s, log: logs/$service.log)"
done

check_processes() {
  local index
  for index in "${!pids[@]}"; do
    if ! kill -0 "${pids[$index]}" 2>/dev/null; then
      echo "${services[$index]} stopped. Check logs/${services[$index]}.log; shutting down the other services." >&2
      exit 1
    fi
  done
}

remaining=${#services[@]}
while (( remaining > 0 )); do
  check_processes
  for index in "${!services[@]}"; do
    [[ ${ready[$index]} == 1 ]] && continue
    endpoint="http://127.0.0.1:${ports[$index]}/readyz"
    if response=$(curl --fail --silent --max-time 2 --output /dev/null --write-out '%{http_code} %{time_total}' "$endpoint"); then
      status=${response%% *}
      response_seconds=${response#* }
      if [[ $status == 200 && $response_seconds =~ ^[0-9]+([.][0-9]+)?$ ]]; then
        latency_ms=$(LC_ALL=C awk -v seconds="$response_seconds" 'BEGIN {printf "%.1f", seconds * 1000}')
        ready[$index]=1
        remaining=$((remaining - 1))
        echo "Ready ${services[$index]} (HTTP $status, $endpoint, response ${latency_ms} ms)"
      fi
    fi
  done
  if (( remaining > 0 && SECONDS - startup_second >= startup_timeout )); then
    echo "Startup timed out after ${startup_timeout}s. Check logs/ for services that are not ready." >&2
    exit 1
  fi
  (( remaining == 0 )) || sleep 0.2
done

ended_at=$(clock_now_ns)
echo "Startup completed at: $(format_timestamp "$ended_at")"
elapsed=$(elapsed_seconds "$started_at" "$ended_at")
echo "Startup duration: ${elapsed} s"
echo 'Press Ctrl+C to stop all services.'
while true; do
  check_processes
  sleep 1
done
