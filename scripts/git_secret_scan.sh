#!/usr/bin/env bash
set -euo pipefail

# Simple git blob scanner: searches blob contents for secret-like keywords and
# computes Shannon entropy for short blobs. Skips blobs over 100KB to avoid
# scanning large binaries.

THRESHOLD=4.5
MAX_SIZE=$((100*1024))

echo "Scanning git objects for secret-like patterns (this may take a few seconds)..."

git rev-list --all | while read -r commit; do
  git ls-tree -r -l "$commit" | awk '{print $3" "$4" "$5}'
done | sort -u | while read -r mode size path; do
  # Ignore empty
  if [[ -z "$path" || "$path" == "0" ]]; then
    continue
  fi
  # Only consider regular files
  if [[ "$size" -gt "$MAX_SIZE" ]]; then
    continue
  fi
  blob=$(git rev-parse "HEAD:$path" 2>/dev/null || true)
  if [[ -z "$blob" ]]; then
    continue
  fi
done >/dev/null

# We'll instead iterate over all blobs
git rev-list --objects --all | cut -d' ' -f1 | while read -r oid; do
  # skip very large objects
  size=$(git cat-file -s "$oid")
  if [[ "$size" -gt "$MAX_SIZE" ]]; then
    continue
  fi
  content=$(git cat-file -p "$oid" 2>/dev/null || true)
  if [[ -z "$content" ]]; then
    continue
  fi

  # keyword checks
  if echo "$content" | egrep -i "PRIVATE KEY|BEGIN RSA PRIVATE KEY|BEGIN OPENSSH PRIVATE KEY|BEGIN EC PRIVATE KEY|aws_secret|aws_secret_access_key|api_key|password|secret|credentials" >/dev/null; then
    echo "POSSIBLE SECRET in blob $oid (size: $size): keyword match"
    continue
  fi

  # entropy check on base64-looking lines
  lines=$(echo "$content" | tr -d '\r' | sed -n '1,200p')
  # compute Shannon entropy for each whitespace-free token longer than 20
  for token in $(echo "$lines" | tr -s '[:space:]' '\n' | egrep -v '^[[:space:]]*$' | awk 'length($0)>20'); do
    # compute entropy
    ent=$(echo -n "$token" | awk '
      {s=split($0,a,""); for(i=1;i<=s;i++){c[a[i]]++} for(k in c){p=c[k]/s; H+= -p*log(p)/log(2)} print H}' 2>/dev/null || true)
    # fallback: if token looks like base64 and long
    if [[ -n "$ent" ]]; then
      ent_n=$(printf "%.2f" "$ent")
      if (( $(echo "$ent_n > $THRESHOLD" | bc -l) )); then
        echo "POSSIBLE SECRET in blob $oid (size: $size): high entropy token (entropy=$ent_n)"
        break
      fi
    fi
  done
done

echo "Scan complete."
