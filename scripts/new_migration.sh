#!/usr/bin/env sh
set -eu

NAME="${1:-}"

if [ -z "$NAME" ]; then
  echo "usage: $0 <migration_name>" >&2
  exit 1
fi

TS="$(date +%Y%m%d%H%M%S)"
FILE="./db/migrations/${TS}_${NAME}.sql"

cat > "$FILE" <<'EOF'
-- 请在这里编写 Atlas versioned migration SQL。
EOF

echo "created migration: $FILE"
