#!/usr/bin/env bash
set -euo pipefail

# seq 1 1000 | sed 's|.*|https://example.com/&|' > test_create_urls.txt
# seq 1001 2000 | sed 's|.*|https://example.com/&|' > test_create_urls_api.txt
# seq 2001 3000 | sed 's|.*|https://example.com/&|' > test_create_urls_api_batch.txt

HOST="${1:-localhost:8080}"
COOKIE_FILE="/tmp/cookie.txt"
RESULTS_DIR="created_urls"

rm -rf "$RESULTS_DIR"
mkdir -p "$RESULTS_DIR"

# Предварительный запрос для получения JWT-cookie
curl -sS -c "$COOKIE_FILE" -o /dev/null "http://${HOST}/"

# ──────────────────────────────────────────────
# Шаг 1: POST / (plain text)
# ──────────────────────────────────────────────
echo "=== [1/6] POST / (plain text) ==="
parallel --results "${RESULTS_DIR}/plain" \
 'curl -sS -X POST -H "Content-Type: text/plain" -b '"${COOKIE_FILE}"' -d {} '"http://${HOST}/"'' \
 :::: test_create_urls.txt

# ──────────────────────────────────────────────
# Шаг 2: POST /api/shorten (JSON)
# ──────────────────────────────────────────────
echo "=== [2/6] POST /api/shorten ==="
parallel --results "${RESULTS_DIR}/api" \
  'curl -sS -X POST -H "Content-Type: application/json" -b '"${COOKIE_FILE}"' -d '\''{"url":"{}"}'\'' '"http://${HOST}/api/shorten"'' \
  :::: test_create_urls_api.txt

# ──────────────────────────────────────────────
# Шаг 3: POST /api/shorten/batch (chunk по 100)
# ──────────────────────────────────────────────
echo "=== [3/6] POST /api/shorten/batch ==="
parallel --results "${RESULTS_DIR}/batch" -N 5 \
  'echo {} | jq -R -s "split(\" \") | to_entries | map({correlation_id: (.key|tostring), original_url: .value})" | curl -sS -X POST -H "Content-Type: application/json" -b '"${COOKIE_FILE}"' -d @- '"http://${HOST}/api/shorten/batch"'' \
  :::: test_create_urls_api_batch.txt

# ──────────────────────────────────────────────
# Шаг 4: Собираем ВСЕ короткие ID из трёх источников
# ──────────────────────────────────────────────
echo "=== [4/6] Собираем все короткие ID ==="
ALL_IDS_FILE="${RESULTS_DIR}/all_short_ids.txt"
> "$ALL_IDS_FILE"

# Из plain-ответов: строка вида http://host/A1b2C3d4E5
grep -rhoP 'http://[^/]+/\K[[:alnum:]]{10}' "${RESULTS_DIR}/plain/" 2>/dev/null >> "$ALL_IDS_FILE" || true

# Из API-ответов: JSON {"result":"http://host/A1b2C3d4E5"}
grep -rl '"result"' "${RESULTS_DIR}/api/" 2>/dev/null | \
  xargs cat 2>/dev/null | \
  jq -r '.result | split("/")[-1]' 2>/dev/null >> "$ALL_IDS_FILE" || true

# Из batch-ответов: JSON [{"short_url":"http://host/A1b2C3d4E5"}, ...]
grep -rl '"short_url"' "${RESULTS_DIR}/batch/" 2>/dev/null | \
  xargs cat 2>/dev/null | \
  jq -r '.[]?.short_url | split("/")[-1]' 2>/dev/null >> "$ALL_IDS_FILE" || true

# Убираем пустые строки и дубликаты
sed -i '/^$/d' "$ALL_IDS_FILE"
sort -u -o "$ALL_IDS_FILE" "$ALL_IDS_FILE"

TOTAL=$(wc -l < "$ALL_IDS_FILE")
echo "  Всего уникальных ID: $TOTAL"
HEY_IDS_FILE="${RESULTS_DIR}/hey_ids.txt"
head -300 "$ALL_IDS_FILE" > "$HEY_IDS_FILE"
HEY_COUNT=$(wc -l < "$HEY_IDS_FILE")

# ──────────────────────────────────────────────
# Шаг 5: hey для GET /{id}
# hey умеет бить только один URL, поэтому запускаем по одному hey на каждый ID параллельно
# ──────────────────────────────────────────────
echo "=== [5/6] hey: GET /{id} ($HEY_COUNT ID) ==="
HEY_REQUESTS="${2:-500}"
HEY_CONCURRENCY="${3:-200}"
HEY_JOBS="${4:-50}"

if [[ -s "$HEY_IDS_FILE" ]]; then
  parallel --results "${RESULTS_DIR}/hey_get" -j "$HEY_JOBS" \
    "./hey -n ${HEY_REQUESTS} -c ${HEY_CONCURRENCY} http://${HOST}/{}" \
    :::: "$HEY_IDS_FILE"
fi

# ──────────────────────────────────────────────
# Шаг 6: DELETE /api/user/urls (ВСЕ созданные ID)
# ──────────────────────────────────────────────
echo "=== [6/6] DELETE /api/user/urls (все $TOTAL ID) ==="

# Группируем ID в JSON-массивы по 50 и удаляем через parallel
# printf разбивает склеенные пробелами ID на отдельные строки,
# jq строит JSON-массив, curl читает тело из stdin (-d @-)
parallel --results "${RESULTS_DIR}/delete" -N 50 \
  'printf "%s\n" {} | jq -R -s '\''split("\n") | map(select(length > 0))'\'' | curl -sS -X DELETE -H "Content-Type: application/json" -b '"${COOKIE_FILE}"' -d @- http://'"${HOST}"'/api/user/urls' \
  :::: "$ALL_IDS_FILE"

echo "=== Готово ==="
