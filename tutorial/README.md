# Build a blog API with kanna

[日本語](README.ja.md)

Write a blog in Go: create, list, read, update, and delete posts, then add comments to the working app. Inspired by [Rails 8: The Demo](https://www.youtube.com/watch?v=X_Hw9P1iZfQ), this walkthrough uses a JSON API instead of browser pages. Its models are `Post` with a title and body, and `Comment` with content and a post ID. Both have integer IDs and timestamps.

The [VHS tape](demo.tape) types the code in Vim at 60–90 ms per character, saves it, runs the generators, and calls the API. [render.go](render.go) generates the Vim scenes from the linked source files. The recording follows the app's development: posts first, comments second, then fixtures and tests. Recording duration measures scripted typing, not a first-time reader's completion time.

## Prepare

Requires Go 1.27.3 or later, Git, Docker with Compose, curl 7.82 or later, and Vim. Recording also requires [VHS](https://github.com/charmbracelet/vhs), ttyd, and ffmpeg. Keep `127.0.0.1:18081` free for the HTTP server.

```sh
git clone https://github.com/go-kanna/showcase.git
cd showcase
sh tutorial/prepare.sh
cd tmp/tutorial
. ./env.sh
mkdir -p model scopes repository postgres mapper
```

Preparation creates a separate `example.com/blog` module, installs pinned tool versions, starts a dedicated PostgreSQL container on an automatically assigned localhost port, and applies [posts.sql](posts.sql). It also provides the common [HTTP/JSON helpers](files/http.go.txt) and [invalid-request tests](files/api_test.go.txt). These two files are supplied infrastructure, not generator output or code typed during the recording. Initial downloads and database startup happen before recording.

All app files live in `tmp/tutorial`. Preparation refuses to overwrite an existing workspace. `LESSON` points to this tutorial's source directory. For each `vim` command below, enter the linked file's contents and save with `Esc`, `:wq`, `Enter`.

## 1. Model posts and generate their database code

Write the [Post model](files/post.go.txt), [scopes declaration](files/scopes.go.txt), and [Repository](files/post_repository.go.txt). The Repository exposes CRUD; `postgres` is its generated PostgreSQL implementation.

kanna-orm supports PostgreSQL and MySQL. This walkthrough uses PostgreSQL; see the [MySQL example](https://github.com/go-kanna/orm/tree/main/example/mysql) for MySQL-specific setup.

```sh
vim model/post.go
vim scopes/scopes.go
vim repository/post.go
go tool kanna-orm describe ./...
go generate ./scopes ./repository
```

Read `postgres/orm_gen.go`: the INSERT, SELECT, UPDATE, DELETE, and row scanning are ordinary Go and SQL. Identity columns are assigned by PostgreSQL. Timestamps are set by the app, rather than automatically by the ORM. Subsequent generation reads `kanna-orm.snapshot.json` without connecting to the database.

## 2. Add the input, response mapping, and API

Write [SavePost](files/input.go.txt), the [initial response mapping](files/post_mapper_base.go.txt), [API actions and container](files/api.go.txt), and [main](files/main.go.txt). Validation generates the title checks. Mapper keeps the JSON response separate from the model. DI constructs the selected PostgreSQL repository from the supplied pool.

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

On a fresh database, the first post has ID 1. Creation returns HTTP 201; reading and updating return 200. An empty title returns 422 with a field-specific error. `main` only opens the pool and starts the server; the tutorial omits signal handling.

## 3. Add comments to the working blog

Apply the comment schema, then write the [Comment model](files/comment.go.txt), [Repository](files/comment_repository.go.txt), [input](files/comment_input.go.txt), and [comment action](files/comments.go.txt).

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

Make these edits to the existing files, as the tape demonstrates:

- Add `Comments []Comment` to `Post`; [the model with comments](files/post_with_comments.go.txt) shows the result.
- Add `Comments []CommentResponse` with `json:"comments,omitempty"` to `PostResponse`, and append [CommentResponse](files/comment_response.go.txt). Mapper generates the nested conversion from the existing `PostToResponse` declaration.
- Add `comments repository.Comment` with `di:"with=postgres.NewComment"` to `API`.
- Add `POST /posts/{id}/comments`, handled by `a.createComment`, to the routes.
- Add `scopes.Posts().WithComments()` to the `Get` call in `showPost`.

```sh
go tool kanna-orm describe ./...
go generate ./scopes ./repository ./mapper
go tool kanna-validate
go tool kanna-di
```

The foreign key in [comments.sql](comments.sql) gives orm the relation to load. The generated mapper converts the loaded comments into nested response values. Deleting a post also deletes its comments through `ON DELETE CASCADE`.

## 4. Generate fixtures, test, and use the complete API

Write the [Repository integration test](files/repository_test.go.txt), which uses generated fixtures and a rolled-back transaction. It checks CRUD, comment loading, nested response mapping, and cascade deletion. Run formatting once, after all edits.

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

A comment returns 201 with its post and nested comments. Reading the post includes the comment. Empty content returns 422. Deletion returns 204, and reading the deleted post returns 404.

## Record and clean up

Run from the showcase repository root, with a fresh workspace and database:

```sh
sh tutorial/prepare.sh
PATH="/opt/homebrew/bin:$PATH" vhs tutorial/demo.tape
docker compose -p kanna-tutorial -f tutorial/compose.yaml down
```

VHS uses Bash without profile or rc files. The PATH prefix selects Homebrew Bash on Apple Silicon macOS; elsewhere, `vhs tutorial/demo.tape` uses Bash on PATH. The output is `tmp/blog-tutorial.mp4`. Each command waits for the prompt, and command failures exit the shell. UUID conversion is not part of this example; mapper demonstrates field and nested-struct mapping.

Before recording again, stop the database and move `tmp/tutorial` aside, then prepare a fresh run. Stopping the dedicated PostgreSQL container discards its data. The source and generated app files remain available in `tmp/tutorial`. HTTP dates, generated timestamps, and timings vary between recordings.

## English captions

After recording, use Go, ImageMagick (`magick`), and ffmpeg to add captions:

```sh
go run tutorial/caption.go
```

This preserves the original video and writes `tmp/blog-tutorial-captioned.mp4`, plus matching `.srt` and `.vtt` subtitle files. Captions appear in an added strip below the terminal, leaving the code visible. Edit [captions.json](captions.json) to change the text or timing. The current cues match the 14:48 recording; check their alignment after recording again. The default font is macOS Helvetica; on other systems, pass an installed ImageMagick font with `-font`, such as `-font DejaVu-Sans`.

For Japanese captions, use [captions.ja.json](captions.ja.json) and a Japanese font. On macOS:

```sh
go run tutorial/caption.go -cues tutorial/captions.ja.json -output tmp/blog-tutorial-ja.mp4 -work tmp/blog-captions-ja/generated -font '/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc'
```

This writes the Japanese video and matching `.srt` and `.vtt` files without replacing the English outputs.
