package usecase

import "github.com/go-kanna/showcase/internal/feature/todo/repository"

// The rows the repositories read stay in the use cases, which return types of
// their own.
//
//kanna:mapper
type _ interface {
	commentSummaryFromRepository(c repository.CommentSummary) CommentSummary
}
