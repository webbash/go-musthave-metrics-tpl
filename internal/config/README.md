# internal/config

В этом пакете хранятся конфигурации приложения.

Загрузку конфигурации можно реализовать из различных источников, например: файлов, переменных окружения, баз данных и других.

Агент и сервер выбирают один транспорт: `http` (по умолчанию) или `grpc`.

| Параметр | Флаг | Окружение | JSON | По умолчанию |
| --- | --- | --- | --- | --- |
| Транспорт | `-transport` | `TRANSPORT` | `transport` | `http` |
| Адрес gRPC | `-grpc-address` | `GRPC_ADDRESS` | `grpc_address` | `localhost:8081` |
| Сертификат gRPC (агент: доверенный CA/самоподписанный сертификат сервера; сервер: свой сертификат) | `-grpc-tls-cert` | `GRPC_TLS_CERT` | `grpc_tls_cert` | нет, обязателен для gRPC |
| Приватный TLS-ключ gRPC (только сервер) | `-grpc-tls-key` | `GRPC_TLS_KEY` | `grpc_tls_key` | нет, обязателен для gRPC |
| Адрес HTTP | `-a` | `ADDRESS` | `address` | `localhost:8080` |

Приоритет: окружение → явно заданный флаг → JSON (`-c`, `-config`, `CONFIG`) → default.
Как и у существующих параметров, сервер учитывает пустые значения окружения,
а агент использует только непустые значения строк из окружения.

Для gRPC на сервере используется существующий `-t`, `TRUSTED_SUBNET`,
`trusted_subnet`: пустое значение отключает проверку, некорректный CIDR
останавливает запуск. Заявленный IP передаётся в metadata `x-real-ip`.
gRPC всегда использует TLS: шифруются метрики и metadata, агент проверяет
сертификат и имя сервера. HTTP-подпись, прикладное HTTP-шифрование и аудит
для gRPC не используются. TLS-настройки не влияют на HTTP и не загружаются
при выборе HTTP. Ошибки загрузки сертификатов/ключа останавливают запуск gRPC;
при ошибке TLS-соединения перехода на открытый транспорт нет.

Для локальной разработки сгенерируйте самоподписанный сертификат:

```bash
make grpc-certs
```

Создаются `tmp/certs/server.crt` и `tmp/certs/server.key` (не попадают в Git).
Существующие файлы команда не перезаписывает. Сертификат действует 365 дней
для `localhost`, `127.0.0.1` и `::1`; для другого имени/IP нужен сертификат
с соответствующим SAN. Агенту передаётся только сертификат, приватный ключ
остаётся на сервере. В рабочем окружении используйте собственные сертификаты
и защищённое хранение ключей, а не локальную пару.

Пример запуска через PostgreSQL:

```bash
DATABASE_DSN='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
  go run ./cmd/server -transport=grpc -grpc-address=localhost:8081 -t=127.0.0.0/8 \
    -grpc-tls-cert=tmp/certs/server.crt -grpc-tls-key=tmp/certs/server.key
go run ./cmd/agent -transport=grpc -grpc-address=127.0.0.1:8081 \
  -grpc-tls-cert=tmp/certs/server.crt
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
