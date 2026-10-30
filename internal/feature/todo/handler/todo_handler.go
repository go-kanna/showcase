package handler

import (
	"context"

	"github.com/go-kanna/showcase/internal/di"
	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/mapper"
	"github.com/go-kanna/showcase/internal/feature/todo/usecase"
	todov1 "github.com/go-kanna/showcase/internal/gen/showcase/todo/v1"
	"github.com/go-kanna/showcase/internal/lib/xslices"
)

//go:generate go tool kanna-di
//kanna:container
type Handler struct {
	todov1.UnimplementedTodoServiceServer

	_ *di.Infra `di:"arg"`

	delete        *usecase.DeleteTodo
	list          *usecase.ListTodo
	put           *usecase.PutTodo
	setCompletion *usecase.SetTodoCompletion
}

func (h Handler) DeleteTodo(
	ctx context.Context, request *todov1.DeleteTodoRequest,
) (*todov1.DeleteTodoResponse, error) {
	_, err := h.delete.Do(ctx, usecase.DeleteTodoInput{TodoID: request.GetId()})
	if err != nil {
		return nil, err
	}
	return &todov1.DeleteTodoResponse{}, nil
}

func (h Handler) ListTodo(ctx context.Context, request *todov1.ListTodoRequest) (*todov1.ListTodoResponse, error) {
	out, err := h.list.Do(ctx, usecase.ListTodoInput{})
	if err != nil {
		return nil, err
	}
	return &todov1.ListTodoResponse{
		Todos: xslices.Map(out.Todos, mapper.TodoToProto),
	}, nil
}

func (h Handler) PutTodo(ctx context.Context, request *todov1.PutTodoRequest) (*todov1.PutTodoResponse, error) {
	in, err := mapper.TodoInputFromProto(request.GetTodo())
	if err != nil {
		// A priority of a number the service does not know.
		return nil, aerrors.InvalidArgument(err.Error())
	}
	if ts := request.GetCompletedAt(); ts != nil {
		in.CompletedAt = new(ts.AsTime())
	}
	if _, err := h.put.Do(ctx, in); err != nil {
		return nil, err
	}
	return &todov1.PutTodoResponse{}, nil
}

func (h Handler) SetTodoCompletion(
	ctx context.Context, request *todov1.SetTodoCompletionRequest,
) (*todov1.SetTodoCompletionResponse, error) {
	_, err := h.setCompletion.Do(ctx, usecase.SetTodoCompletionInput{
		TodoID:    request.GetId(),
		Completed: request.GetCompleted(),
	})
	if err != nil {
		return nil, err
	}
	return &todov1.SetTodoCompletionResponse{}, nil
}

var _ todov1.TodoServiceServer = (*Handler)(nil)
