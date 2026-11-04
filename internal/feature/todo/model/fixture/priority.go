package fixture

import "github.com/go-kanna/showcase/internal/feature/todo/model"

// Priority is the priority of the to-dos the fixtures make: normal, as most
// are, rather than low, the first constant, which kanna-fixture would pick.
func Priority() model.Priority {
	return model.PriorityNormal
}
