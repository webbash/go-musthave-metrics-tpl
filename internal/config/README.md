# internal/config

В этом пакете хранятся конфигурации приложения.

Загрузку конфигурации можно реализовать из различных источников, например: файлов, переменных окружения, баз данных и других.

Агент и сервер выбирают один транспорт: `http` (по умолчанию) или `grpc`.

| Параметр | Флаг | Окружение | JSON | По умолчанию |
| --- | --- | --- | --- | --- |
| Транспорт | `-transport` | `TRANSPORT` | `transport` | `http` |
| Адрес gRPC | `-grpc-address` | `GRPC_ADDRESS` | `grpc_address` | `localhost:8081` |
| Адрес HTTP | `-a` | `ADDRESS` | `address` | `localhost:8080` |

Приоритет: окружение → явно заданный флаг → JSON (`-c`, `-config`, `CONFIG`) → default.
Как и у существующих параметров, сервер учитывает пустые значения окружения,
а агент использует только непустые значения строк из окружения.

Для gRPC на сервере используется существующий `-t`, `TRUSTED_SUBNET`,
`trusted_subnet`: пустое значение отключает проверку, некорректный CIDR
останавливает запуск. Заявленный IP передаётся в metadata `x-real-ip`.
В первой версии gRPC не использует HTTP-подпись, шифрование или аудит.

Пример запуска через PostgreSQL:

```bash
DATABASE_DSN='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
  go run ./cmd/server -transport=grpc -grpc-address=localhost:8081 -t=127.0.0.0/8
go run ./cmd/agent -transport=grpc -grpc-address=127.0.0.1:8081
```

Адрес агента в примере задаёт IPv4 для соответствия подсети. Агент и сервер
настраиваются независимо: transport должен совпадать, переключения после ошибки нет.

Примеры JSON: [сервер](../../configs/server.example.json), [агент](../../configs/agent.example.json).

Проверка gRPC с PostgreSQL:

```bash
TEST_DATABASE_DSN='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
  go test -v ./cmd/server -run TestPostgresRoundTrip
```

Тест создаёт временную таблицу в отдельной сессии и не меняет существующую таблицу metrics.
