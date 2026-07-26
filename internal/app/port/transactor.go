package port

import "context"

// Transactor memungkinkan usecase membungkus beberapa operasi dalam satu DB transaction.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
