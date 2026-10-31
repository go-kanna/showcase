# showcase

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)

[kanna](https://github.com/go-kanna/kanna) コードジェネレータが 1 つのアプリケーションの中で働く様子を見せるために書かれた、Go の ToDo gRPC サービスです。各ジェネレータは、本来手で書くコードの一部を引き受けます。コンストラクタ、バリデーション、レイヤ間のマッピング、テストフィクスチャ、型付きのメッセージ、値オブジェクト、リポジトリの裏側のデータベースコードです。生成されたものはすべてコミットされているので、生成元と並べて読めます。

Go 1.27 以上、Docker と Compose、[buf](https://buf.build/docs/cli/installation/) CLI が必要です。

## 各ジェネレータの役割

| ツール | 場所 | 生成するもの |
|---|---|---|
| [kanna-di](https://github.com/go-kanna/di) | `internal/di`, `usecase`, `handler` | `NewInfra`, `NewPutTodo`, `NewHandler` など、各 `//kanna:container` 構造体のコンストラクタ |
| [kanna-validate](https://github.com/go-kanna/validate) | `config`, `usecase` | `validate` タグから生成する `Validate` メソッド。`rule=model.CompletedAt` も含む |
| [kanna-mapper](https://github.com/go-kanna/mapper) | `mapper` | `TodoToProto`, `TodoInputFromProto`, `ListTodoInputFromProto`。値オブジェクト、enum、oneof の直和型も含む |
| [kanna-fixture](https://github.com/go-kanna/fixture) | `model/fixture` | テスト用の、全フィールドが埋まった `fixture.Todo()` |
| [kanna-i18n](https://github.com/go-kanna/i18n) | `internal/messages` | `messages.TodoNotFound()` とバリデーションエラーコードの文言。英語と日本語 |
| [kanna-value](https://github.com/go-kanna/value) | `model` | `NewCompletedAt` と、`model.CompletedAt`（2000 年から 2100 年の間に保たれる時刻）および enum の `model.Priority` の JSON と SQL のメソッド |
| [kanna-orm](https://github.com/go-kanna/orm) | `scopes`, `postgres` | `repository.Todo` を実装する `postgres.Todo` と、`scopes.Todos().CreatedAtAsc()` のように読み取りに使うスコープ |

その他の構成要素: protobuf と gRPC に [buf](https://buf.build)、マイグレーションに [goose](https://github.com/pressly/goose)、テストごとの新しいデータベースに [pgtestdb](https://github.com/peterldowns/pgtestdb) を使っています。

## 構成

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

## 実行

PostgreSQL を起動し、マイグレーションを適用してから、サーバーを起動します。`.go` ファイルが変更されるたびに [air](https://github.com/air-verse/air) でリビルドされます。

```sh
docker compose up -d db
make db-migrate
docker compose up api
```

サーバーは `localhost:8080` で待ち受けます。リフレクションサービスは持たないので、クライアントに `proto/` のスキーマを渡してください:

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

## 開発

| コマンド | 内容 |
|---|---|
| `make gen` | すべてのジェネレータを実行する: buf を実行してから、出力が使われる順に `go generate` を実行する |
| `make gen-check` | buf または kanna のジェネレータがコミット済みの内容を変更する場合に失敗する |
| `make gen-fresh-check` | kanna のジェネレータが生成したものを削除し、1 回の実行で再生成できなければ失敗する |
| `make test` | テストを、それぞれ専用のデータベースに対して実行する |
| `make lint` | golangci-lint を実行する |
| `make sql-lint` | マイグレーションに sqlfluff を実行する |
| `make db-reset` | データベースを削除して作り直し、マイグレーションを適用する |
| `make db-describe` | マイグレーション済みのデータベースのスキーマを `kanna-orm.snapshot.json` に記録する |
| `make new-migration name=...` | マイグレーションファイルを追加する |

`make gen` は Makefile の `GEN_PACKAGES` の順に、パッケージごとにジェネレータを実行します。各ジェネレータは自身のインポート先のパッケージを読み込むので、その時点でそれらの生成コードが存在している必要があります。`//go:generate` の行を追加したパッケージは、出力を使うパッケージの後ろに並ぶようこのリストに加えてください。CI が実行する `make gen-fresh-check` は、順序のせいで 2 回の実行が必要になると失敗します。

kanna-orm はデータベースではなく `kanna-orm.snapshot.json` から生成するので、`make gen` にデータベースは不要です。マイグレーションの後は、`make db-migrate` で適用し、`make db-describe` で記録して、スナップショットをマイグレーションと一緒にコミットしてください。両者が一致しないと CI が失敗します。

ユースケースはリポジトリとトランザクタを引数で受け取るので、入力の扱いをチェックするテストは、`put_todo_use_case_test.go` のようにフェイクを渡せばよく、データベースは不要です。SQL の動作は PostgreSQL に対してテストします。そのテストには、起動中のデータベースと、`make db-migrate` で 1 度マイグレーションしておくこと（接続に使うロールもこれで作られます）、環境変数の URL が必要です:

```sh
make db-migrate
envsubst < .env.example > .env
set -a && . ./.env && set +a
make test
```

## ライセンス

[MIT](./LICENSE)
