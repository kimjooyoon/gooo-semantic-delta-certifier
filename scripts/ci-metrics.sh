#!/usr/bin/env bash
set -euo pipefail

build_time=""
test_time=""
conformance_time=""
cache_dir=""
output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --build-time) build_time="$2"; shift 2 ;;
    --test-time) test_time="$2"; shift 2 ;;
    --conformance-time) conformance_time="$2"; shift 2 ;;
    --cache-dir) cache_dir="$2"; shift 2 ;;
    --output) output="$2"; shift 2 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done

read_wall_ms() {
  local file="$1"
  awk -F': ' '/Elapsed \(wall clock\) time/ {print $2}' "$file" | awk -F: '{
    if (NF == 2) { split($2, s, "\\."); printf "%.0f\n", (($1 * 60) + s[1]) * 1000 + (s[2] + 0) * 10 }
    else { split($3, s, "\\."); printf "%.0f\n", (($1 * 3600) + ($2 * 60) + s[1]) * 1000 + (s[2] + 0) * 10 }
  }'
}

read_peak_rss_kib() {
  local file="$1"
  awk -F': ' '/Maximum resident set size/ {print $2; exit}' "$file"
}

build_wall="$(read_wall_ms "$build_time")"
test_wall="$(read_wall_ms "$test_time")"
conformance_wall="$(read_wall_ms "$conformance_time")"
build_rss="$(read_peak_rss_kib "$build_time")"
test_rss="$(read_peak_rss_kib "$test_time")"
conformance_rss="$(read_peak_rss_kib "$conformance_time")"
total_wall="$((build_wall + test_wall + conformance_wall))"
peak_rss="$build_rss"
if [[ "$test_rss" -gt "$peak_rss" ]]; then peak_rss="$test_rss"; fi
if [[ "$conformance_rss" -gt "$peak_rss" ]]; then peak_rss="$conformance_rss"; fi
cache_bytes="$(du -sk "$cache_dir" | awk '{print $1 * 1024}')"

jq -n \
  --argjson build "$build_wall" \
  --argjson test "$test_wall" \
  --argjson conformance "$conformance_wall" \
  --argjson wall "$total_wall" \
  --argjson peak "$peak_rss" \
  --argjson cache "$cache_bytes" \
  '{
    build: {value:$build, unit:"ms", status:"CLOSED", unknown:null},
    test: {value:$test, unit:"ms", status:"CLOSED", unknown:null},
    conformance: {value:$conformance, unit:"ms", status:"CLOSED", unknown:null},
    wall: {value:$wall, unit:"ms", status:"CLOSED", unknown:null},
    peak_rss: {value:$peak, unit:"kib", status:"CLOSED", unknown:null},
    cache: {value:$cache, unit:"bytes", status:"CLOSED", unknown:null}
  }' > "$output"
