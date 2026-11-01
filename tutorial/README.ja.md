# kanna でブログ API を作る

[English](README.md)

Go で記事の作成・一覧・詳細・更新・削除を実装し、動くブログにコメントを追加します。[Rails 8: The Demo](https://www.youtube.com/watch?v=X_Hw9P1iZfQ) の構成を参考に、画面の代わりに JSON API を使います。モデルはタイトルと本文を持つ `Post` と、内容と記事 ID を持つ `Comment` です。どちらにも整数 ID と作成・更新日時を持たせます。

[VHS の台本](demo.tape) は、Vim で 1 文字あたり 60〜90ms で入力し、保存・生成・API 呼び出しまで実行します。[render.go](render.go) がリンク先のソースから Vim のシーンを生成します。ツール別の紹介ではなく、記事、コメント、fixture とテストの順にアプリを育てます。録画の長さは、初見の読者が完成させるまでの所要時間とは異なります。

## 準備

Go 1.27.3 以降、Git、Docker と Compose、curl 7.82 以降、Vim が必要です。録画には [VHS](https://github.com/charmbracelet/vhs)、ttyd、ffmpeg も必要です。HTTP サーバー用に `127.0.0.1:18081` を空けてください。

```sh
git clone https://github.com/go-kanna/showcase.git
cd showcase
sh tutorial/prepare.sh
cd tmp/tutorial
. ./env.sh
mkdir -p model scopes repository postgres mapper
```

準備では、独立した `example.com/blog` モジュールを作り、固定バージョンのツールを登録し、専用 PostgreSQL コンテナを localhost の自動割り当てポートで起動して、[posts.sql](posts.sql) を適用します。共通の [HTTP/JSON ヘルパー](files/http.go.txt) と [不正なリクエストのテスト](files/api_test.go.txt) も配置します。この 2 ファイルは準備済みの土台であり、生成コードでも、動画中に入力するコードでもありません。初回のダウンロードと DB 起動は録画の前に済ませます。

アプリのファイルは `tmp/tutorial` に置かれます。既存の作業ディレクトリは上書きしません。`LESSON` は、このチュートリアルのソースディレクトリです。以下の `vim` コマンドでは、リンク先の内容を入力し、`Esc`、`:wq`、`Enter` で保存してください。

## 1. 記事のモデルと DB コードを作る

[Post モデル](files/post.go.txt)、[スコープの宣言](files/scopes.go.txt)、[Repository](files/post_repository.go.txt) を書きます。Repository は CRUD の契約、`postgres` は生成される PostgreSQL 実装です。

kanna-orm は PostgreSQL と MySQL に対応しています。このチュートリアルでは PostgreSQL を使います。MySQL での設定は [MySQL のサンプル](https://github.com/go-kanna/orm/tree/main/example/mysql) を参照してください。

```sh
vim model/post.go
vim scopes/scopes.go
vim repository/post.go
go tool kanna-orm describe ./...
go generate ./scopes ./repository
```

`postgres/orm_gen.go` には、INSERT・SELECT・UPDATE・DELETE と、行をモデルへ読み込む Go コードがあります。identity 列は PostgreSQL が採番します。日時は ORM が自動設定するのではなく、アプリで設定します。その後の生成は `kanna-orm.snapshot.json` を読み、DB への接続は不要です。

## 2. 入力・レスポンス変換・API を追加する

[SavePost](files/input.go.txt)、[初期のレスポンス変換](files/post_mapper_base.go.txt)、[API の処理とコンテナ](files/api.go.txt)、[main](files/main.go.txt) を書きます。validate はタイトルの検証を、mapper はモデルから独立した JSON レスポンスへの変換を生成します。di は引数のプールから、指定した PostgreSQL の Repository を構築します。

```sh
vim input.go
go tool kanna-validate
vim mapper/mapper.go
go generate ./mapper
vim api.go
go tool kanna-di
vim main.go
go build -o blog .
./blog >server.log 2>&1 & server_pid=$!
curl --retry 20 --retry-connrefused --retry-delay 1 -fsS localhost:18081/posts
curl -sS -i --json '{"title":"Hello, kanna","body":"My first blog post."}' localhost:18081/posts
curl -fsS localhost:18081/posts/1
curl -sS -X PUT --json '{"title":"A blog with kanna","body":"Updated without writing SQL."}' localhost:18081/posts/1
curl -fsS localhost:18081/posts
curl -sS -i --json '{"title":""}' localhost:18081/posts
kill "$server_pid"; wait "$server_pid" || true
```

新しい DB では、最初の記事の ID は 1 です。作成は HTTP 201、取得・更新は 200 を返します。空のタイトルは、フィールドごとのエラーとともに 422 を返します。`main` はプールを開いてサーバーを起動するだけで、シグナル処理は省いています。

## 3. 動いているブログにコメントを追加する

コメントのスキーマを適用し、[Comment モデル](files/comment.go.txt)、[Repository](files/comment_repository.go.txt)、[入力](files/comment_input.go.txt)、[コメント投稿の処理](files/comments.go.txt) を書きます。

```sh
docker compose -p kanna-tutorial -f "$LESSON/compose.yaml" exec -T db psql -U tutorial -d tutorial -v ON_ERROR_STOP=1 < "$LESSON/comments.sql"
vim model/comment.go
vim model/post.go
vim repository/comment.go
vim comment_input.go
vim mapper/mapper.go
vim comments.go
vim api.go
```

動画と同じように、既存ファイルには次の変更を加えます。

- `Post` に `Comments []Comment` を追加します。[コメント追加後のモデル](files/post_with_comments.go.txt) が完成形です。
- `PostResponse` に `json:"comments,omitempty"` を付けた `Comments []CommentResponse` を追加し、[CommentResponse](files/comment_response.go.txt) を追記します。既存の `PostToResponse` の宣言から、mapper が入れ子の変換を生成します。
- `API` に `di:"with=postgres.NewComment"` を付けた `comments repository.Comment` を追加します。
- ルーティングに `a.createComment` を呼ぶ `POST /posts/{id}/comments` を追加します。
- `showPost` の `Get` 呼び出しに `scopes.Posts().WithComments()` を追加します。

```sh
go tool kanna-orm describe ./...
go generate ./scopes ./repository ./mapper
go tool kanna-validate
go tool kanna-di
```

[comments.sql](comments.sql) の外部キーから orm が関連を読み込みます。生成された mapper は、読み込んだコメントを入れ子のレスポンスへ変換します。`ON DELETE CASCADE` により、記事を削除するとコメントも削除されます。

## 4. fixture を生成し、テストして完成した API を使う

生成された fixture とロールバックするトランザクションを使う [Repository の統合テスト](files/repository_test.go.txt) を書きます。CRUD、コメントの関連取得、入れ子のレスポンス変換、記事削除時のコメント削除を検証します。フォーマットはすべての編集が終わった後に一度だけ実行します。

```sh
go generate ./model
vim repository/repository_test.go
gofmt -w .
go test -v ./...
KANNA_CHECK=1 go generate ./...
go build -o blog .
./blog >server.log 2>&1 & server_pid=$!
curl --retry 20 --retry-connrefused --retry-delay 1 -fsS localhost:18081/posts/1
curl -sS -i --json '{"content":"It works!"}' localhost:18081/posts/1/comments
curl -fsS localhost:18081/posts/1
curl -sS -i --json '{"content":""}' localhost:18081/posts/1/comments
curl -sS -i -X DELETE localhost:18081/posts/1
curl -sS -i localhost:18081/posts/1
kill "$server_pid"; wait "$server_pid" || true
```

コメント投稿は記事とコメントを含むレスポンスで 201 を返し、記事を取得するとコメントも返ります。空の内容は 422、削除は 204、削除後の記事取得は 404 です。

## 録画と片付け

新しい作業ディレクトリと DB を使い、showcase のルートから実行します。

```sh
sh tutorial/prepare.sh
PATH="/opt/homebrew/bin:$PATH" vhs tutorial/demo.tape
docker compose -p kanna-tutorial -f tutorial/compose.yaml down
```

VHS は profile と rc ファイルを読まない Bash を使います。Apple Silicon の macOS では PATH の指定で Homebrew の Bash を選びます。他の環境では `vhs tutorial/demo.tape` で PATH 上の Bash を使います。出力は `tmp/blog-tutorial.mp4` です。各コマンドはプロンプトを待ち、失敗するとシェルを終了します。この例では UUID の変換を追加せず、mapper によるフィールドと入れ子の構造体の変換を見せます。

録画し直す場合は、DB を停止し、`tmp/tutorial` を別の場所に移動してから、もう一度準備してください。専用 PostgreSQL のコンテナを停止するとデータは破棄されます。ソースと生成されたアプリは `tmp/tutorial` に残ります。HTTP の日時、モデルに設定する日時、実行時間は録画ごとに変わります。

## 英語キャプション

録画後、Go、ImageMagick（`magick`）、ffmpeg を使ってキャプションを追加します。

```sh
go run tutorial/caption.go
```

元動画を残し、`tmp/blog-tutorial-captioned.mp4` と、同名の `.srt`・`.vtt` 字幕ファイルを出力します。コードが隠れないよう、ターミナルの下に字幕用の帯を追加します。文面と表示時刻は [captions.json](captions.json) で変更できます。現在の時刻は 14 分 48 秒の録画に合わせているため、録り直した場合は対応する場面を確認してください。既定のフォントは macOS の Helvetica です。他の環境では `-font DejaVu-Sans` のように、ImageMagick で使えるフォントを指定してください。
