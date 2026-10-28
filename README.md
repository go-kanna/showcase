# showcase

A to-do gRPC service in Go, written to show the [kanna](https://github.com/go-kanna/kanna) code generators at work in
one application. Each generator takes over a piece of code you would otherwise write by hand: constructors, validation,
mappings between layers, test fixtures, and typed messages. Everything it writes is committed, so you can read it next
to what it was generated from.

Requires Go 1.27, Docker with Compose, and the [buf](https://buf.build/docs/cli/installation/) CLI.

## What each generator does here

| Tool                                                   | Where                               | What it writes                                                                                 |
|--------------------------------------------------------|-------------------------------------|------------------------------------------------------------------------------------------------|
| [kanna-di](https://github.com/go-kanna/di)             | `internal/di`, `usecase`, `handler` | `NewInfra`, `NewPutTodo`, `NewHandler`, …: the constructor of each `//kanna:container` struct  |
| [kanna-validate](https://github.com/go-kanna/validate) | `config`, `usecase`                 | `Validate` methods from the `validate` tags, `rule=model.CompletedAtRule` among them           |
| [kanna-mapper](https://github.com/go-kanna/mapper)     | `mapper`, `queries`                 | `TodoToProto` (model → protobuf) and `RowToTodo` (sqlc row → model), validated values included |
| [kanna-fixture](https://github.com/go-kanna/fixture)   | `model/fixture`                     | `fixture.Todo()`, every field filled, for tests                                                |
| [kanna-i18n](https://github.com/go-kanna/i18n)         | `internal/messages`                 | `messages.TodoNotFound()` and the texts of validation error codes, in English and Japanese     |
| [validated](https://github.com/go-kanna/validated)     | `model`                             | Nothing: a library. `model.CompletedAt` is a time held between 2000 and 2100                   |

The rest of the stack: [sqlc](https://sqlc.dev) for queries, [buf](https://buf.build) for protobuf and gRPC,
[goose](https://github.com/pressly/goose) for migrations, and [pgtestdb](https://github.com/peterldowns/pgtestdb) for
a fresh database per test.

## Layout

```text
cmd/api                   the server: gRPC over h2c, shut down on SIGTERM
config                    settings read from the environment
proto                     the TodoService definition
internal/
  di                      the containers the server is built from: Config and Infra
  feature/todo/
    model                 the domain type, and the fixtures generated for it
    repository            the interface the use cases depend on
    postgres              its implementation on sqlc
    queries               SQL, the code sqlc writes, and the row → model mapping
    usecase               one struct per operation: List, Put, SetCompletion, Delete
    mapper                the model → protobuf mapping
    handler               the gRPC handler
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
  --data '{"todo": {"id": "01890a5d-ac96-774b-bcce-b302099a8057", "title": "Buy milk"}}' \
  http://localhost:8080/showcase.todo.v1.TodoService/PutTodo

buf curl --protocol grpc --http2-prior-knowledge --schema proto \
  http://localhost:8080/showcase.todo.v1.TodoService/ListTodo
```

## Develop

| Command                        | What it does                                                      |
|--------------------------------|-------------------------------------------------------------------|
| `make gen`                     | run every generator: `go generate`, then buf, then sqlc           |
| `make gen-check`               | fail if the kanna generators would change what is committed       |
| `make test`                    | run the tests, each against its own database                      |
| `make lint`                    | run golangci-lint                                                 |
| `make sql-lint`                | run sqlfluff over the migrations                                  |
| `make db-reset`                | drop the database, create it again, and apply the migrations      |
| `make new-migration name=...`  | add a migration file                                              |

The tests need the database running, migrated once with `make db-migrate` (that creates the roles they connect as), and
its URLs in the environment:

```sh
make db-migrate
envsubst < .env.example > .env
set -a && . ./.env && set +a
make test
```

## License

[MIT](./LICENSE)
