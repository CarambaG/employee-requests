#!/bin/sh
set -eu

: "${DATABASE_URL:?DATABASE_URL is required}"

output_dir=${OUTPUT_DIR:-/results}
performance_dir=${PERFORMANCE_DIR:-/performance}
report_file=${REPORT_FILE:-optimized.txt}
rows_file=${ROWS_FILE:-optimized_rows.csv}
comparison_file=${COMPARISON_FILE:-comparison.md}
baseline_file=${BASELINE_FILE:-baseline.txt}

mkdir -p "$output_dir"
report_path="$output_dir/$report_file"
rows_path="$output_dir/$rows_file"
comparison_path="$output_dir/$comparison_file"
baseline_path="$output_dir/$baseline_file"

rm -f "$comparison_path"

index_exists=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "
    SELECT TO_REGCLASS('public.requests_assignee_status_due_at_idx') IS NOT NULL;
")

if [ "$index_exists" != "t" ]; then
    echo "The optimized index is missing. Run 'make migrate-up' first." >&2
    exit 1
fi

case "${ASSIGNEE_ID:-}" in
    "")
        if [ -f "$baseline_path" ]; then
            assignee_id=$(awk -F ': ' '/^Assignee ID:/ { print $2; exit }' "$baseline_path")
        else
            assignee_id=$(
                psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "
                    SELECT r.assignee_id
                    FROM requests AS r
                    JOIN request_statuses AS rs ON rs.id = r.status_id
                    WHERE rs.code = 'in_progress'
                      AND r.due_at < NOW()
                    GROUP BY r.assignee_id
                    ORDER BY COUNT(*) DESC, r.assignee_id
                    LIMIT 1;
                "
            )
        fi
        ;;
    *[!0-9]*|0)
        echo "ASSIGNEE_ID must be a positive integer" >&2
        exit 1
        ;;
    *)
        assignee_id=$ASSIGNEE_ID
        ;;
esac

case "$assignee_id" in
    ""|*[!0-9]*|0)
        echo "Could not determine a valid assignee ID. Set ASSIGNEE_ID explicitly." >&2
        exit 1
        ;;
esac

request_count=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "SELECT COUNT(*) FROM requests;")
employee_count=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "SELECT COUNT(*) FROM employees;")
index_size=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "SELECT PG_SIZE_PRETTY(PG_RELATION_SIZE('requests_assignee_status_due_at_idx'));")
matched_rows=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "
    SELECT COUNT(*)
    FROM requests AS r
    WHERE r.assignee_id = $assignee_id
      AND r.status_id = (
          SELECT id FROM request_statuses WHERE code = 'in_progress'
      )
      AND r.due_at < NOW();
")

if [ "$matched_rows" -eq 0 ]; then
    echo "Assignee $assignee_id has no overdue in-progress requests." >&2
    exit 1
fi

if [ -f "$baseline_path" ]; then
    baseline_assignee=$(awk -F ': ' '/^Assignee ID:/ { print $2; exit }' "$baseline_path")

    if [ -z "$baseline_assignee" ]; then
        echo "Could not read Assignee ID from $baseline_path." >&2
        exit 1
    fi

    if [ "$baseline_assignee" != "$assignee_id" ]; then
        echo "Baseline used assignee $baseline_assignee, but optimized benchmark would use $assignee_id." >&2
        echo "Run again with ASSIGNEE_ID=$baseline_assignee for a valid comparison." >&2
        exit 1
    fi
fi

{
    echo "Employee Requests optimized performance benchmark"
    echo "Generated at (UTC): $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
    echo "PostgreSQL: $(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c 'SHOW server_version;')"
    echo "Employees: $employee_count"
    echo "Requests: $request_count"
    echo "Assignee ID: $assignee_id"
    echo "Matching rows: $matched_rows"
    echo "Index: requests_assignee_status_due_at_idx (assignee_id, status_id, due_at)"
    echo "Index size: $index_size"
    echo
    echo "The first EXPLAIN ANALYZE is a warm-up; use the following three execution times for comparison."
    echo
} > "$report_path"

psql "$DATABASE_URL" \
    -X \
    -v ON_ERROR_STOP=1 \
    -v assignee_id="$assignee_id" \
    -f "$performance_dir/optimized.sql" >> "$report_path"

psql "$DATABASE_URL" \
    -X \
    --csv \
    -v ON_ERROR_STOP=1 \
    -v assignee_id="$assignee_id" \
    -f "$performance_dir/query.sql" > "$rows_path"

if [ -f "$baseline_path" ]; then
    baseline_average=$(grep 'Execution Time:' "$baseline_path" | tail -n 3 | awk '{ sum += $3; count++ } END { if (count == 3) printf "%.3f", sum / count }')
    optimized_average=$(grep 'Execution Time:' "$report_path" | tail -n 3 | awk '{ sum += $3; count++ } END { if (count == 3) printf "%.3f", sum / count }')

    if [ -n "$baseline_average" ] && [ -n "$optimized_average" ]; then
        speedup=$(awk -v before="$baseline_average" -v after="$optimized_average" 'BEGIN { if (after > 0) printf "%.2f", before / after }')
        reduction=$(awk -v before="$baseline_average" -v after="$optimized_average" 'BEGIN { if (before > 0) printf "%.2f", (before - after) * 100 / before }')
        postgres_version=$(awk -F ': ' '/^PostgreSQL:/ { print $2; exit }' "$report_path")
        baseline_plan=$(grep -m 1 -E 'Parallel Seq Scan on requests|Seq Scan on requests|Bitmap Heap Scan on requests|Index Scan.*requests' "$baseline_path" | sed 's/^[[:space:]]*//' || true)
        optimized_heap_plan=$(grep -m 1 -E 'Bitmap Heap Scan on requests|Index Scan.*requests_assignee_status_due_at_idx' "$report_path" | sed 's/^[[:space:]]*//' || true)
        optimized_index_plan=$(grep -m 1 -E 'Bitmap Index Scan on requests_assignee_status_due_at_idx|Index Only Scan.*requests_assignee_status_due_at_idx' "$report_path" | sed 's/^[[:space:]]*//' || true)
        baseline_sort=$(grep -m 1 -E '^[[:space:]]*(->  )?Sort ' "$baseline_path" | sed 's/^[[:space:]]*//' || true)
        optimized_sort=$(grep -m 1 -E '^[[:space:]]*(->  )?Sort ' "$report_path" | sed 's/^[[:space:]]*//' || true)

        {
            echo "# Сравнение производительности"
            echo
            echo "- PostgreSQL: \`${postgres_version:-unknown}\`"
            echo "- Сотрудников: \`$employee_count\`"
            echo "- Заявок: \`$request_count\`"
            echo "- Исполнитель: \`$assignee_id\`"
            echo "- Подходящих заявок: \`$matched_rows\`"
            echo "- Среднее время до оптимизации: \`$baseline_average ms\`"
            echo "- Среднее время после оптимизации: \`$optimized_average ms\`"
            echo "- Ускорение: \`${speedup}x\`"
            echo "- Снижение времени выполнения: \`${reduction}%\`"
            echo "- Размер индекса: \`$index_size\`"
            echo
            echo "## Планы выполнения"
            echo
            echo "До оптимизации:"
            echo
            echo '```text'
            echo "${baseline_plan:-plan not detected}"
            if [ -n "$baseline_sort" ]; then
                echo "$baseline_sort"
            fi
            echo '```'
            echo
            echo "После оптимизации:"
            echo
            echo '```text'
            echo "${optimized_heap_plan:-plan not detected}"
            if [ -n "$optimized_index_plan" ]; then
                echo "$optimized_index_plan"
            fi
            if [ -n "$optimized_sort" ]; then
                echo "$optimized_sort"
            fi
            echo '```'
            echo
            echo "Составной индекс начинается с полей равенства \`assignee_id\` и \`status_id\`, а затем содержит диапазонное поле \`due_at\`. Он устраняет полный просмотр таблицы. Наличие отдельной сортировки зависит от выбранного PostgreSQL способа доступа; bitmap heap scan не сохраняет порядок B-tree."
        } > "$comparison_path"
    fi
fi

printf '\nOptimized report: %s\n' "$report_path"
printf 'Sorted query result: %s\n' "$rows_path"
if [ -f "$comparison_path" ]; then
    printf 'Comparison summary: %s\n' "$comparison_path"
fi
printf '\n'
cat "$report_path"
