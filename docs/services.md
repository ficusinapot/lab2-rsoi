# Hotels Booking System

Gateway (`8080`) предоставляет внешний `/api/v1` API варианта v2.
Reservation (`8070`) хранит отели и бронирования; Payment (`8060`) — оплаты;
Loyalty (`8050`) — счётчик, уровень и скидку. Все межсервисные вызовы выполняет
Gateway по HTTP. Gateway не использует БД.

Каждый сервис — отдельный Go-модуль. Компоненты получают зависимости через
конструкторы; общая инфраструктура находится в `modules/platform`.

## Запуск и готовность

```sh
docker compose up --build -d
WAIT_PORTS=8080,8070,8060,8050 scripts/wait-script.sh
```

PostgreSQL 13 создаёт БД `reservations`, `payments`, `loyalties` на новом volume.
Учебные credentials: `program` / `test`. Если volume уже существует, проверьте
список БД; init-скрипты PostgreSQL повторно не выполняются. Недостающие БД
создаются отдельно, без удаления volume:

```sh
docker compose exec reservation-db psql -U program -d postgres -c 'CREATE DATABASE payments OWNER program'
docker compose exec reservation-db psql -U program -d postgres -c 'CREATE DATABASE loyalties OWNER program'
```

Все приложения читают выбранный YAML через `--config`. Адреса и таймауты
HTTP-клиентов Gateway находятся в секции `clients`. Настройки серверов,
документации, логирования и метрик также задаются YAML. Нет настройки через
переменные окружения. `config.docker.yaml` применяется в Docker.

Для Payment/Loyalty Docker entrypoint запускает внешний Atlas runner, который
читает URL БД из YAML через yq. Схема создаётся native SQL-миграциями с `atlas.sum`;
приложения не выполняют миграции. Generated Ent-клиенты хранятся в репозитории.

```sh
make -C services/payment generate
make -C services/loyalty generate
make -C services/payment migrate
make -C services/loyalty migrate
make -C services/gateway run CONFIG=config.yaml
```

Каждый сервис предоставляет `/manage/health`, `/metrics`, `/api/v1/docs` и
`/api/v1/openapi.json`. Проверка готовности сервисов с БД проверяет подключение;
Gateway проверяет собственную доступность. RED использует шаблоны маршрутов;
метки не содержат пользователей и UUID. Platform сериализует короткий участок
подготовки и записи autometrics v1.1.0, использующий общий небезопасный PRNG
и карту вызовов. Сами операции выполняются параллельно.

## Внутренние HTTP-контракты

Reservation сохраняет контракт, описанный в `reservation.md`.

Payment:

- `POST /api/v1/payments`, `{ "price": 27000 }` → `201`,
  `{ "paymentUid": "UUID", "status": "PAID", "price": 27000 }`.
- `GET /api/v1/payments/{paymentUid}` → `200`, тот же объект.
- `DELETE /api/v1/payments/{paymentUid}` → `204`, статус `CANCELED`.
  Повторная отмена сохраняет статус; неизвестная оплата возвращает `404`.

Loyalty использует `X-User-Name`:

- `GET /api/v1/loyalty` → `200`,
  `{ "status": "BRONZE", "discount": 5, "reservationCount": 0 }`.
- `POST /api/v1/loyalty/reservations` увеличивает счётчик.
- `DELETE /api/v1/loyalty/reservations` уменьшает его, минимум — ноль.
- Оба изменения возвращают `200` с актуальным объектом Loyalty.

Первое обращение создаёт пользователя `BRONZE/0/5`. Изменения счётчика и уровня
выполняются транзакцией с блокировкой строки. При 10 бронированиях уровень
`SILVER` и скидка 7%, при 20 — `GOLD` и 10%; отмена пересчитывает уровень.

## Внешнее поведение

Контракт Gateway соответствует материалам `v2`, внутренний `paymentUid`
не выдаётся. Чтение бронирований включает объект отеля и состояние оплаты;
`/me` дополнительно включает Loyalty. Внешний POST возвращает `200`,
внутренний POST Reservation — `201`.

Дата передаётся в формате `YYYY-MM-DD`; конечная дата должна быть позже
начальной. Стоимость считается целыми числами: ночи × цена × (100 − скидка) / 100,
дробная часть отбрасывается. Переполнение и результат вне диапазона PostgreSQL
INT отклоняются до создания оплаты. Применяется скидка до увеличения счётчика.
Пагинация Reservation начинается с 1; Gateway допускает page=0 из OpenAPI как
первую страницу и сохраняет 0 в ответе. Размер внешней страницы — от 1 до 100.

Создание выполняет Payment → Reservation → увеличение Loyalty после чтения
отеля и скидки. Отмена выполняет Reservation → Payment → уменьшение Loyalty
после проверки владельца. Чужое бронирование возвращает `404`.
Уже отменённое бронирование возвращает `204` без повторных изменений.

При первой ошибке сценарий останавливается: повторов и компенсаций нет.
Сетевые ошибки возвращают `502`, таймауты — `504`; ошибки пользователя — `400`,
отсутствующие сущности — `404`. Выполненные шаги сохраняются. Повторный DELETE
после частичного отказа не восстанавливает оставшиеся шаги. Конкурентная отмена
не имеет гарантий распределённой транзакции; на каждом отдельном сервисе
сохраняются его локальные гарантии.

## Проверки

```sh
make -C modules/platform check
make -C services/reservation check
make -C services/payment check
make -C services/loyalty check
make -C services/gateway check
make -C services/gateway e2e-test
```

Все Go-тесты запускаются через gotestsum. E2E самостоятельно строит четыре
образа и запускает изолированную сеть с PostgreSQL и случайными портами хоста.
Внешние сценарии выполняются через Gateway; отдельные внутренние вызовы
проверяют конкурентные изменения Loyalty.

Для проверки неизменённой Postman-коллекции нужны Newman и работающий Compose:

```sh
scripts/prepare-classroom.sh
newman run -e v2/postman/environment.json v2/postman/collection.json
```

Подготовка явно загружает отель и пользователя `Test Max` с `GOLD/25/10`,
не перезаписывая существующие записи. Fixtures не входят в миграции.
Postman-коллекцию и environment можно импортировать в Postman вручную.
Workflow Classroom выполняет сборки, проверки, unit/e2e-тесты, запускает Compose
и autograding v2. Удалённый результат появляется после push на GitHub.

## Refactoring and concurrency limits

Loyalty, Payment and Gateway use domain status types; Ent enums remain inside
storage. Gateway injects separate Reservation, Payment and Loyalty interfaces.
External JSON fields and enum values are unchanged.

Gateway requires these additional YAML parameters in every selected config:

```yaml
enrichment_concurrency: 4
clients:
  max_in_flight: 32
  max_idle_connections: 64
  max_idle_connections_per_host: 16
  idle_connection_timeout: 60s
```

These keys supplement the existing URLs and client timeout. All values must be
positive. Up to four workers enrich one reservation list, preserving its order.
Each worker reads Hotel and Payment sequentially. Any error cancels sibling
work, waits for it to finish and returns an error without a partial list.
The limit of 32 applies to all outbound requests sharing the Gateway client;
waiting callers respect request cancellation. HTTP connections are reused and
idle connections are closed after the HTTP server shuts down.
Creation and cancellation remain sequential; `/me` also remains sequential.
Limits apply per Gateway instance, not across replicas.

Nonunique indexes accelerate lookup by Payment UUID and Reservation username.
New Atlas migrations add them without changing existing migration files.
Loyalty counter updates retain PostgreSQL row locks and a transaction, so
independent connection pools cannot lose increments or split tier fields.

## Verification and diagnostics

```sh
make -C services/payment integration-test
make -C services/loyalty integration-test
make -C services/reservation integration-test
make -C services/gateway e2e-test
```

Payment and Loyalty integration targets create isolated PostgreSQL 13 containers
and apply SQL through the external `atlas` executable on PATH. Docker and Atlas
are required. Reservation integration tests retain their documented test database
and YAML configuration. All integration targets use gotestsum and `-race`.
Unit, integration and e2e targets write JUnit reports into each module's `bin/`.
CI uploads them; failed container tests print logs, and acceptance CI also uploads
Compose logs. `scripts/check-generated.sh` checks that committed Ent clients and
Reservation mocks match their generators; run it from a clean checkout.
Atlas is pinned to the same immutable image in CI and Docker entrypoints.

Gateway e2e covers stopped upstream services and the resulting partial states:
a Reservation cancellation failure leaves Payment and Loyalty unchanged;
a Payment cancellation failure leaves Reservation canceled, Payment paid and
Loyalty unchanged. A failed payment creation does not create a Reservation or
increase Loyalty. There are no compensations or automatic retries. A repeated
cancel of an already canceled Reservation does not repair a partial failure.
Concurrent cancellation across Gateway instances still has no distributed
exactly-once guarantee; this refactoring does not introduce one.

The list benchmark uses a fixed 1 ms hotel lookup delay with immediate payment
responses. A local run on linux/amd64 with Intel Core Ultra 5 125H measured:

| Reservations | Workers | Time per list | Bytes per list | Allocations | Peak hotel calls |
| --- | --- | --- | --- | --- | --- |
| 1 | 1 | 1.11 ms | 1030 | 17 | 1 |
| 1 | 4 | 1.10 ms | 1027 | 17 | 1 |
| 10 | 1 | 10.78 ms | 5659 | 71 | 1 |
| 10 | 4 | 3.30 ms | 6379 | 78 | 4 |
| 50 | 1 | 54.25 ms | 26496 | 312 | 1 |
| 50 | 4 | 14.33 ms | 27042 | 319 | 4 |

This synthetic comparison isolates enrichment scheduling; it is not a production
latency guarantee. There is no timing threshold in CI. Run the benchmark through
`gotestsum -- -run '^$' -bench BenchmarkList -benchtime=100ms ./internal/usecase/bookings`
from the Gateway module. Gotestsum v1.12.3 may list benchmark events as incomplete
cases with Go 1.27; the benchmark output and underlying Go process exit status
remain the source of the measurements.
