# showcase

[![日本語](https://img.shields.io/badge/lang-%E6%97%A5%E6%9C%AC%E8%AA%9E-blue.svg)](README.ja.md)

A to-do gRPC service in Go, written to show the [kanna](https://go-kanna.github.io/) code generators at work in one
application. Each generator takes over a piece of code you would otherwise write by hand: constructors, reading
settings, validation, mappings between layers, test fixtures, typed messages, value types, and the database code behind
a repository. Everything it writes is committed, so you can read it next to what it was generated from.

To build a blog JSON API step by step with orm, validate, mapper, di, and fixture, follow the [tutorial](tutorial/README.md). Its VHS tape records the same walkthrough without manual screen capture.

Requires Go 1.27, Docker with Compose, and the [buf](https://buf.build/docs/cli/installation/) CLI.

## What each generator does here

| Tool                                                   | Where                               | What it writes                                                                                                                                                    |
|--------------------------------------------------------|-------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [kanna-di](https://github.com/go-kanna/di)             | `internal/di`, `usecase`, `handler` | `NewInfra`, `NewPutTodo`, `NewHandler`, …: the constructor of each `//kanna:container` struct                                                                     |
| [kanna-validate](https://github.com/go-kanna/validate) | `usecase`                           | `Validate` methods from the `validate` tags, `rule=model.CompletedAt` among them                                                                                  |
| [kanna-config](https://github.com/go-kanna/config)     | `config`                            | `LoadAPI` and `LoadDatabase`, which read flags and the environment and check the `validate` tags, and `DefineAPIFlags` and `DefineDatabaseFlags`                  |
| [kanna-mapper](https://github.com/go-kanna/mapper)     | `mapper`                            | `TodoToProto`, `CommentToProto`, `CommentSummaryToProto`, `TodoInputFromProto`, and `ListTodoInputFromProto`, value types, enums, and a oneof's sum type included |
| [kanna-fixture](https://github.com/go-kanna/fixture)   | `model/fixture`                     | `fixture.Todo()`, for tests: every value filled, a pointer left nil, and the priority from `fixture.Priority()`, written by hand                                  |
| [kanna-i18n](https://github.com/go-kanna/i18n)         | `internal/messages`                 | `messages.TodoNotFound()` and the texts of validation error codes, in English and Japanese                                                                        |
| [kanna-value](https://github.com/go-kanna/value)       | `model`                             | `NewCompletedAt` and the JSON and SQL methods of `model.CompletedAt`, a time held between 2000 and 2100, and of `model.Priority`, an enum                         |
| [kanna-orm](https://github.com/go-kanna/orm)           | `scopes`, `postgres`                | `postgres.Todo` and `postgres.Comment`, which implement the repositories, and the scopes they read and page through, as `scopes.Todos().WithComments(...)`        |

The rest of the stack: [buf](https://buf.build) for protobuf and gRPC, [goose](https://github.com/pressly/goose) for
migrations, and [pgtestdb](https://github.com/peterldowns/pgtestdb) for a fresh database per test.

## Layout

```text
cmd/api                   the server: gRPC over h2c, shut down on SIGTERM
config                    settings read from flags and the environment
proto                     the TodoService definition
internal/
  di                      the containers the server is built from: Config and Infra
  feature/todo/
    model                 the domain types, and the fixtures generated for them
    scopes                the scope types the repositories read through
    repository            the interfaces the use cases depend on
    postgres              their implementations, which kanna-orm writes, and the SQL written by hand for them
    usecase               one struct per operation, given the repositories and transactor it works through
    mapper                the mappings between protobuf and the model or the use case inputs
    handler               the gRPC handler, which puts the feature together and chooses the repositories
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
  --data '{"todoId": "01890a5d-ac96-774b-bcce-b302099a8057", "body": "Whole milk"}' \
  http://localhost:8080/showcase.todo.v1.TodoService/AddComment

buf curl --protocol grpc --http2-prior-knowledge --schema proto \
  http://localhost:8080/showcase.todo.v1.TodoService/ListRecentComments

buf curl --protocol grpc --http2-prior-knowledge --schema proto \
  http://localhost:8080/showcase.todo.v1.TodoService/ListTodo

buf curl --protocol grpc --http2-prior-knowledge --schema proto \
  --data '{"priority": {"priority": "PRIORITY_HIGH"}}' \
  http://localhost:8080/showcase.todo.v1.TodoService/ListTodo

buf curl --protocol grpc --http2-prior-knowledge --schema proto \
  --data '{"pageSize": 1}' \
  http://localhost:8080/showcase.todo.v1.TodoService/ListTodo
```

Each to-do of a list comes with its three latest comments, and with `commentCount`, which a query written by hand in
`todo_queries.sql` adds to it. kanna-orm reads the query's rows as those of `todos`, so the scopes of ListTodo narrow,
order, and page them as they would the table. A page ends with a `nextPageToken` while to-dos follow it; pass it as
`pageToken` to read the next page.

`ListRecentComments` gives the latest comments, each with the to-do it is on, which kanna-orm joins in the same query
from the nested struct of `repository.CommentSummary`.

## Develop

| Command                       | What it does                                                                    |
|-------------------------------|---------------------------------------------------------------------------------|
| `make gen`                    | run every generator: buf, then the kanna generators with `go tool kanna gen`    |
| `make gen-check`              | fail if buf or the kanna generators would change what is committed              |
| `make test`                   | run the tests, each against its own database                                    |
| `make lint`                   | run golangci-lint                                                               |
| `make vet`                    | check the code calling what the generators write with `go tool kanna vet`       |
| `make sql-lint`               | run sqlfluff over the migrations                                                |
| `make db-reset`               | drop the database, create it again, and apply the migrations                    |
| `make db-describe`            | record the schema of the migrated database in `kanna-orm.snapshot.json`         |
| `make new-migration name=...` | add a migration file                                                            |

`make gen` runs the kanna generators with `go tool kanna gen`, which works out their order from what each reads and
writes, so a package that gains a `//go:generate` line needs nothing else. `go tool kanna gen -plan` shows the order and
why.

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
