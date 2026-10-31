# showcase

[![日本語](https://img.shields.io/badge/lang-%E6%97%A5%E6%9C%AC%E8%AA%9E-blue.svg)](README.ja.md)

A to-do gRPC service in Go, written to show the [kanna](https://github.com/go-kanna/kanna) code generators at work in
one application. Each generator takes over a piece of code you would otherwise write by hand: constructors, validation,
mappings between layers, test fixtures, typed messages, value types, and the database code behind a repository.
Everything it writes is committed, so you can read it next to what it was generated from.

Requires Go 1.27, Docker with Compose, and the [buf](https://buf.build/docs/cli/installation/) CLI.

## What each generator does here

| Tool                                                   | Where                               | What it writes                                                                                                                            |
|--------------------------------------------------------|-------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| [kanna-di](https://github.com/go-kanna/di)             | `internal/di`, `usecase`, `handler` | `NewInfra`, `NewPutTodo`, `NewHandler`, …: the constructor of each `//kanna:container` struct                                             |
| [kanna-validate](https://github.com/go-kanna/validate) | `config`, `usecase`                 | `Validate` methods from the `validate` tags, `rule=model.CompletedAt` among them                                                          |
| [kanna-mapper](https://github.com/go-kanna/mapper)     | `mapper`                            | `TodoToProto`, `TodoInputFromProto`, and `ListTodoInputFromProto`, value types, enums, and a oneof's sum type included                    |
| [kanna-fixture](https://github.com/go-kanna/fixture)   | `model/fixture`                     | `fixture.Todo()`, every field filled, for tests                                                                                           |
| [kanna-i18n](https://github.com/go-kanna/i18n)         | `internal/messages`                 | `messages.TodoNotFound()` and the texts of validation error codes, in English and Japanese                                                |
| [kanna-value](https://github.com/go-kanna/value)       | `model`                             | `NewCompletedAt` and the JSON and SQL methods of `model.CompletedAt`, a time held between 2000 and 2100, and of `model.Priority`, an enum |
| [kanna-orm](https://github.com/go-kanna/orm)           | `scopes`, `postgres`                | `postgres.Todo`, which implements `repository.Todo`, and the scopes it reads through, as `scopes.Todos().CreatedAtAsc()`                  |

The rest of the stack: [buf](https://buf.build) for protobuf and gRPC, [goose](https://github.com/pressly/goose) for
migrations, and [pgtestdb](https://github.com/peterldowns/pgtestdb) for a fresh database per test.

## Layout

```text
cmd/api                   the server: gRPC over h2c, shut down on SIGTERM
config                    settings read from the environment
proto                     the TodoService definition
internal/
  di                      the containers the server is built from: Config and Infra
  feature/todo/
    model                 the domain type, and the fixtures generated for it
    scopes                the scope types the repository reads through
    repository            the interface the use cases depend on
    postgres              its implementation, which kanna-orm writes, and the SQL written by hand for it
    usecase               one struct per operation, given the repository and transactor it works through
    mapper                the model → protobuf mapping
    handler               the gRPC handler, which puts the feature together and chooses the repository
  gen                     the code buf writes from proto/
  messages                the messages compiled from locales/
  infra/storage           pgx pools, transactions, and migrations
test/infra                a database per test, migrated from a template
```

## Run

Start PostgreSQL, apply the migrations, then start the server. It rebuilds with [air](https://github.com/air-verse/air)
whenever a `.go` file changes.

```sh
docker compose up -d db
make db-migrate
docker compose up api
```

The server listens on `localhost:8080`. It has no reflection service, so give the client the schema in `proto/`:

```sh
buf curl --protocol grpc --http2-prior-knowledge --schema proto \
  --data '{"todo": {"id": "01890a5d-ac96-774b-bcce-b302099a8057", "title": "Buy milk", "priority": "PRIORITY_HIGH"}}' \
  http://localhost:8080/showcase.todo.v1.TodoService/PutTodo

buf curl --protocol grpc --http2-prior-knowledge --schema proto \
  http://localhost:8080/showcase.todo.v1.TodoService/ListTodo

buf curl --protocol grpc --http2-prior-knowledge --schema proto \
  --data '{"priority": {"priority": "PRIORITY_HIGH"}}' \
  http://localhost:8080/showcase.todo.v1.TodoService/ListTodo
```

## Develop

| Command                       | What it does                                                                    |
|-------------------------------|---------------------------------------------------------------------------------|
| `make gen`                    | run every generator: buf, then `go generate` in the order outputs are used      |
| `make gen-check`              | fail if buf or the kanna generators would change what is committed              |
| `make gen-fresh-check`        | delete what the kanna generators wrote, and fail unless one run writes it again |
| `make test`                   | run the tests, each against its own database                                    |
| `make lint`                   | run golangci-lint                                                               |
| `make sql-lint`               | run sqlfluff over the migrations                                                |
| `make db-reset`               | drop the database, create it again, and apply the migrations                    |
| `make db-describe`            | record the schema of the migrated database in `kanna-orm.snapshot.json`         |
| `make new-migration name=...` | add a migration file                                                            |

`make gen` runs the generators package by package, in the order of `GEN_PACKAGES` in the Makefile: each loads the
packages its own imports, whose generated code has to be there by then. A package that gains a `//go:generate` line
goes into that list, after the packages whose output it uses; `make gen-fresh-check`, which CI runs, fails when the
order needs a second run.

kanna-orm generates from `kanna-orm.snapshot.json`, not from a database, so `make gen` needs none. After a migration,
apply it with `make db-migrate`, record it with `make db-describe`, and commit the snapshot along with the migration;
CI fails when the two disagree.

A use case takes its repository and transactor as arguments, so a test that checks what it does with its input gives
it fakes and needs no database, as `put_todo_use_case_test.go` does; what the SQL does is tested against PostgreSQL.
Those tests need the database running, migrated once with `make db-migrate` (that creates the roles they connect as),
and its URLs in the environment:

```sh
make db-migrate
envsubst < .env.example > .env
set -a && . ./.env && set +a
make test
```

## License

[MIT](./LICENSE)
