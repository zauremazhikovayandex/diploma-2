# GophKeeper Server (diploma-2)

GophKeeper — это серверная часть менеджера паролей.  
Сервис отвечает за:

- регистрацию и аутентификацию пользователей;
- хранение приватных данных (логин/пароль, тексты, бинарные данные, банковские карты);
- шифрование пользовательских данных перед записью в БД;
- выдачу и синхронизацию данных для авторизованных клиентов.

---

## Технологии

- **Go**, `gin-gonic` — HTTP-сервер и роутинг.
- **PostgreSQL**, `pgxpool` — БД и пул соединений.
- **JWT** — аутентификация и авторизация.
- **bcrypt** — безопасное хранение пароля пользователя.
- **AES-GCM + base64** — шифрование пользовательских данных (логины/пароли, и т.п.).
- **logrus** — структурированное логирование.

---

## Общая архитектура

Сервер — классический layered design:

- `app/main.go` — входная точка, сборка всех зависимостей.
- `internal/*` — бизнес-логика и HTTP-слой (auth, handlers, repositories).
- `pkg/*` — переиспользуемые инфраструктурные пакеты (config, db, logger, cryptoenc).

---

## Структура проекта

```
diploma-2/
├── app/
│   └── main.go                  # запуск HTTP-сервера
├── internal/
│   ├── auth/                    # авторизация, JWT, middleware
│   │   ├── auth.go
│   │   ├── jwt.go
│   │   └── middlewares.go
│   ├── handlers/                # HTTP-хендлеры (REST API)
│   │   ├── base_handlers.go     # health, register, login
│   │   └── passwords.go         # CRUD для пар логин/пароль
│   └── repositories/            # доступ к БД на уровне доменных сущностей
│       ├── passwords.go         # repo для паролей
│       └── users.go             # repo для получения user_id по login
├── pkg/
│   ├── config/
│   │   └── config.go            # конфиг, env/flags, параметры сервиса и БД
│   ├── cryptoenc/
│   │   └── cryptoenc.go         # AES-GCM шифрование строк и JSON
│   ├── db/
│   │   ├── connector.go         # инициализация пула PostgreSQL (pgxpool), singleton
│   │   ├── utils.go             # универсальные SELECT/EXEC с таймаутами
│   │   └── migrations.go        # CreateTables: схемы logins, passwords, и др.
│   └── logger/
│       ├── logger.go            # общий интерфейс логгера и request-логгер
│       ├── drivers/
│       │   └── stdout.go        # реализация логгера на базе logrus в stdout
│       └── message/
│           └── LogMessage.go    # формат лог-сообщений
└── internal/handlers/passwords_test.go  # юнит-тесты для хендлеров паролей
