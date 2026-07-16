# Проверка производительности

## Целевой запрос

Необходимо получить все просроченные заявки конкретного исполнителя, находящиеся в статусе `in_progress`, и отсортировать их по сроку выполнения:

```sql
SELECT
    r.number,
    r.created_at,
    r.author_id,
    r.assignee_id,
    r.description,
    r.due_at,
    r.status_id,
    r.updated_at
FROM requests AS r
WHERE r.assignee_id = $1
  AND r.status_id = (
      SELECT id
      FROM request_statuses
      WHERE code = 'in_progress'
  )
  AND r.due_at < NOW()
ORDER BY r.due_at;
```

## Условия тестирования

- PostgreSQL 17.10;
- 1000 сотрудников;
- 1 000 000 заявок;
- исполнитель `21`;
- 962 подходящие заявки;
- один прогревочный запуск и три измеряемых запуска;
- одинаковая база и одинаковый исполнитель до и после оптимизации.

## Результат до оптимизации

Среднее время трёх запусков: **27.764 ms**.

Основная часть плана:

```text
Gather Merge
  -> Sort by r.due_at
       -> Parallel Seq Scan on requests
```

PostgreSQL параллельно просматривал таблицу `requests`, отбрасывая примерно 333 тысячи строк в каждом worker, а затем сортировал найденные заявки. Подробный план сохранён в [`../performance-results/baseline.txt`](../performance-results/baseline.txt).

## Выполненная оптимизация

Миграция `000002_optimize_overdue_assignee_query` создаёт индекс:

```sql
CREATE INDEX requests_assignee_status_due_at_idx
    ON requests (assignee_id, status_id, due_at);
```

Порядок полей выбран по условиям запроса:

1. `assignee_id` — равенство по конкретному исполнителю;
2. `status_id` — равенство по статусу;
3. `due_at` — диапазон `due_at < NOW()` и поле сортировки.

Поле `description` не включено в индекс, поскольку оно может быть большим и существенно увеличило бы размер индекса и стоимость операций записи.

## Результат после оптимизации

Среднее время трёх запусков: **2.354 ms**.

Основная часть плана:

```text
Sort by r.due_at
  -> Bitmap Heap Scan on requests
       -> Bitmap Index Scan on requests_assignee_status_due_at_idx
```

В зафиксированном плане PostgreSQL использовал индекс для получения ссылок только на подходящие строки, затем прочитал 962 heap-страницы и выполнил небольшую сортировку размером 130 kB. Составной индекс не устранил `Sort` в конкретном плане, потому что bitmap heap scan не сохраняет порядок B-tree. Основной выигрыш получен за счёт отказа от полного просмотра миллиона строк.

Подробный план сохранён в [`../performance-results/optimized.txt`](../performance-results/optimized.txt).

## Итог

| Показатель | Значение |
|---|---:|
| Среднее время до оптимизации | 27.764 ms |
| Среднее время после оптимизации | 2.354 ms |
| Ускорение | 11.79x |
| Снижение времени | 91.52% |
| Размер индекса | 39 MB |

Сводное сравнение находится в [`../performance-results/comparison.md`](../performance-results/comparison.md).

## Повторный запуск

Подготовить приложение и данные:

```bash
make up
make seed
make seed-verify
```

На текущей ветке индекс уже применён миграцией. Полный воспроизводимый цикл выполняется командой:

```bash
make benchmark-full
```

Сценарий:

1. удаляет индекс `requests_assignee_status_due_at_idx`;
2. выполняет baseline и сохраняет `baseline.txt`;
3. создаёт индекс заново и выполняет `ANALYZE requests`;
4. выполняет оптимизированный замер;
5. формирует `optimized.txt`, CSV-выборки и `comparison.md`.

Сценарий изменяет индекс в базе, поэтому предназначен только для локальной тестовой среды.
