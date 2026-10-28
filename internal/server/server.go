package server

import (
	"net/http"
	"time"

	"google.golang.org/grpc"

	"github.com/go-kanna/showcase/internal/di"
	todo "github.com/go-kanna/showcase/internal/feature/todo/handler"
	todov1 "github.com/go-kanna/showcase/internal/gen/showcase/todo/v1"
)

func Handler(infra *di.Infra) http.Handler {
	server := grpc.NewServer(grpc.UnaryInterceptor(errorInterceptor))
	todov1.RegisterTodoServiceServer(server, todo.NewHandler(infra))
	return server
}

func New(addr string, infra *di.Infra) *http.Server {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	return &http.Server{
		Addr:              addr,
		Handler:           Handler(infra),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		Protocols:         protocols,
	}
}
