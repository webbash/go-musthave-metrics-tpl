# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

PProf (17 Increment)

❯ go tool pprof -top -diff_base=profiles/base.mem profiles/result.mem
File: get_value_list.test
Type: alloc_space
Time: 2026-08-17 01:17:38 EEST
Showing nodes accounting for -128412.31MB, 84.39% of 152173.01MB total
Dropped 54 nodes (cum <= 760.87MB)
flat flat% sum% cum cum%
-151892.31MB 99.82% 99.82% -152168.01MB 100%
github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list.buildHTMLConcat
16326.87MB 10.73% 89.09% 16326.87MB 10.73% strings.(*Builder).WriteString (inline)
3620.58MB 2.38% 86.71% 3517.41MB 2.31% fmt.Sprintf
3533.05MB 2.32% 84.39% 23653.03MB 15.54%
github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list.buildHTMLBuilder
-0.50MB 0.00033% 84.39% -128515.97MB 84.45% testing.(*B).runN
0 0% 84.39% -152168.01MB 100%
github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list.BenchmarkHandler_buildHTML.func1
0 0% 84.39% 23653.03MB 15.54%
github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list.BenchmarkHandler_buildHTML.func2
0 0% 84.39% -128461.31MB 84.42% testing.(*B).launch