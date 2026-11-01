# Run from the showcase root after: sh tutorial/prepare.sh
Output tmp/blog-tutorial.mp4

Require go
Require curl
Require vim
Require docker
Set Shell bash
Set Width 1440
Set Height 900
Set FontSize 22
Set Framerate 15
Set TypingSpeed 68ms
Set Theme "Catppuccin Latte"

Hide
Type "export PS1='blog> '; set -e; cd tmp/tutorial && . ./env.sh"
Enter
Wait+Line@5m /blog>/
Type `vim() { command vim -Nu NONE -n -i NONE -c 'set number showmode laststatus=2 tabstop=4 background=light noautoindent nosmartindent nocindent formatoptions=' -c 'syntax on' "$@"; }`
Enter
Wait+Line@5m /blog>/
Ctrl+L
Sleep 100ms
Show

# Start with posts and their database contract.
Type "mkdir -p model scopes repository postgres mapper"
Enter
Wait+Line@5m /blog>/
Source tmp/tutorial-tapes/post.tape
Source tmp/tutorial-tapes/scopes.tape
Source tmp/tutorial-tapes/post_repository.tape
Type `go tool kanna-orm describe ./...`
Enter
Wait+Line@5m /blog>/
Type "go generate ./scopes ./repository"
Enter
Wait+Line@5m /blog>/
Type "sed -n '55,95p' postgres/orm_gen.go"
Enter
Wait+Line@5m /blog>/
Sleep 3s
Ctrl+L

# Give the API an input and a separate JSON response.
Source tmp/tutorial-tapes/input.tape
Type "go tool kanna-validate"
Enter
Wait+Line@5m /blog>/
Source tmp/tutorial-tapes/post_mapper_base.tape
Type "go generate ./mapper"
Enter
Wait+Line@5m /blog>/
Type "cat mapper/mapper_gen.go"
Enter
Wait+Line@5m /blog>/
Sleep 3s
Ctrl+L

# Write the CRUD actions, generate the constructor, and start the blog.
Source tmp/tutorial-tapes/api.tape
Type "go tool kanna-di"
Enter
Wait+Line@5m /blog>/
Type "cat di_gen.go"
Enter
Wait+Line@5m /blog>/
Sleep 3s
Source tmp/tutorial-tapes/main.tape
Type "go build -o blog ."
Enter
Wait+Line@5m /blog>/
Type "./blog >server.log 2>&1 & server_pid=$!"
Enter
Wait+Line@5m /blog>/
Type "curl --retry 20 --retry-connrefused --retry-delay 1 -fsS localhost:18081/posts"
Enter
Wait+Line@5m /blog>/
Type `curl -sS -i --json '{"title":"Hello, kanna","body":"My first blog post."}' localhost:18081/posts`
Enter
Wait+Line@5m /blog>/
Sleep 3s
Type "curl -fsS localhost:18081/posts/1"
Enter
Wait+Line@5m /blog>/
Sleep 3s
Type `curl -sS -X PUT --json '{"title":"A blog with kanna","body":"Updated without writing SQL."}' localhost:18081/posts/1`
Enter
Wait+Line@5m /blog>/
Sleep 3s
Type "curl -fsS localhost:18081/posts"
Enter
Wait+Line@5m /blog>/
Sleep 3s
Type `curl -sS -i --json '{"title":""}' localhost:18081/posts`
Enter
Wait+Line@5m /blog>/
Sleep 3s
Type `kill "$server_pid"; wait "$server_pid" || true`
Enter
Wait+Line@5m /blog>/
Ctrl+L

# Add comments to the blog that already works.
Type `cat "$LESSON/comments.sql"`
Enter
Wait+Line@5m /blog>/
Sleep 2s
Type `docker compose -p kanna-tutorial -f "$LESSON/compose.yaml" exec -T db psql -U tutorial -d tutorial -v ON_ERROR_STOP=1 < "$LESSON/comments.sql"`
Enter
Wait+Line@5m /blog>/
Source tmp/tutorial-tapes/comment.tape
Type "vim model/post.go"
Enter
Wait+Screen@10s /model\/post.go/
Type "/UpdatedAt"
Enter
Type "o"
Tab
Type "Comments []Comment"
Escape
Type ":wq"
Enter
Wait+Line@5m /blog>/
Source tmp/tutorial-tapes/comment_repository.tape
Source tmp/tutorial-tapes/comment_input.tape

# Extend the response and let mapper generate the nested conversion.
Type "vim mapper/mapper.go"
Enter
Wait+Screen@10s /mapper\/mapper.go/
Type "/UpdatedAt"
Enter
Type "o"
Tab
Type 'Comments []CommentResponse `json:"comments,omitempty"`'
Escape
Type ":wq"
Enter
Wait+Line@5m /blog>/
Source tmp/tutorial-tapes/comment_response.tape
Source tmp/tutorial-tapes/comments.tape

# Add one dependency, one route, and eager loading to the existing API.
Type "vim api.go"
Enter
Wait+Screen@10s /api.go/
Type "/with=postgres.NewPost"
Enter
Type "o"
Tab
Type 'comments repository.Comment `di:"with=postgres.NewComment"`'
Escape
Type "/return mux"
Enter
Type "O"
Tab
Type 'mux.HandleFunc("POST /posts/{id}/comments", a.createComment)'
Escape
Type "gg"
Type "/a.posts.Get"
Enter
Type "$i"
Type ", scopes.Posts().WithComments()"
Escape
Type ":wq"
Enter
Wait+Line@5m /blog>/
Type `go tool kanna-orm describe ./...`
Enter
Wait+Line@5m /blog>/
Type "go generate ./scopes ./repository ./mapper"
Enter
Wait+Line@5m /blog>/
Type "go tool kanna-validate"
Enter
Wait+Line@5m /blog>/
Type "go tool kanna-di"
Enter
Wait+Line@5m /blog>/
Type "sed -n '8,36p' mapper/mapper_gen.go"
Enter
Wait+Line@5m /blog>/
Sleep 3s
Ctrl+L

# Generate fixtures and test the real PostgreSQL relation and response mapping.
Type "go generate ./model"
Enter
Wait+Line@5m /blog>/
Source tmp/tutorial-tapes/repository_test.tape
Type "gofmt -w ."
Enter
Wait+Line@5m /blog>/
Type "go test -v ./..."
Enter
Wait+Line@5m /blog>/
Sleep 3s
Type "KANNA_CHECK=1 go generate ./..."
Enter
Wait+Line@5m /blog>/
Type "go build -o blog ."
Enter
Wait+Line@5m /blog>/
Type "./blog >server.log 2>&1 & server_pid=$!"
Enter
Wait+Line@5m /blog>/
Type "curl --retry 20 --retry-connrefused --retry-delay 1 -fsS localhost:18081/posts/1"
Enter
Wait+Line@5m /blog>/
Type `curl -sS -i --json '{"content":"It works!"}' localhost:18081/posts/1/comments`
Enter
Wait+Line@5m /blog>/
Sleep 4s
Type "curl -fsS localhost:18081/posts/1"
Enter
Wait+Line@5m /blog>/
Sleep 4s
Type `curl -sS -i --json '{"content":""}' localhost:18081/posts/1/comments`
Enter
Wait+Line@5m /blog>/
Sleep 3s
Type "curl -sS -i -X DELETE localhost:18081/posts/1"
Enter
Wait+Line@5m /blog>/
Type "curl -sS -i localhost:18081/posts/1"
Enter
Wait+Line@5m /blog>/
Sleep 4s

Hide
Type `kill "$server_pid"; wait "$server_pid" || true`
Enter
Wait+Line@5m /blog>/
