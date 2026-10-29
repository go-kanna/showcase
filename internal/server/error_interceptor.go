package server

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/go-kanna/i18n"
	"github.com/go-kanna/validate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/messages"
)

var kindCodes = []struct {
	kind error
	code codes.Code
}{
	{aerrors.ErrNotFound, codes.NotFound},
	{aerrors.ErrConflict, codes.AlreadyExists},
	{aerrors.ErrPrecondition, codes.FailedPrecondition},
	{aerrors.ErrUnauthenticated, codes.Unauthenticated},
	{aerrors.ErrPermissionDenied, codes.PermissionDenied},
	{aerrors.ErrInvalidArgument, codes.InvalidArgument},
	{aerrors.ErrResourceExhausted, codes.ResourceExhausted},
	{aerrors.ErrAborted, codes.Aborted},
}

func errorInterceptor(
	ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, toStatus(ctx, info.FullMethod, err)
	}
	return resp, nil
}

// toStatus tells the client what it can act on, and logs the rest.
func toStatus(ctx context.Context, method string, err error) error {
	if verrs, ok := errors.AsType[validate.Errors](err); ok {
		return status.Error(codes.InvalidArgument, validationMessage(ctx, verrs))
	}
	if aerr, ok := errors.AsType[aerrors.Error](err); ok {
		return status.Error(kindCode(aerr), message(ctx, aerr))
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return status.FromContextError(err).Err()
	}
	slog.ErrorContext(ctx, "unhandled error", "method", method, "error", err)
	return status.Error(codes.Internal, "internal error")
}

func kindCode(aerr aerrors.Error) codes.Code {
	for _, kc := range kindCodes {
		if errors.Is(aerr, kc.kind) {
			return kc.code
		}
	}
	return codes.Unknown
}

func message(ctx context.Context, aerr aerrors.Error) string {
	m, ok := aerr.Message()
	if !ok {
		return aerr.Error()
	}
	return localizer(ctx).Localize(m)
}

func validationMessage(ctx context.Context, verrs validate.Errors) string {
	return strings.Join(i18n.FullMessages(localizer(ctx), verrs), "; ")
}

func localizer(ctx context.Context) i18n.Localizer {
	return messages.Localizer(metadata.ValueFromIncomingContext(ctx, "accept-language")...)
}
