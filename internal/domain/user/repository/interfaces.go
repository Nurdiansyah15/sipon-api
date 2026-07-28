package repository

import (
	"context"
	"sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/entity"
	"sipon-api/internal/domain/user/valueobject"
)

// UserRepository adalah satu-satunya repository untuk User aggregate.
// Save dan Update menyimpan seluruh aggregate (LoginIdentities + Credentials).
// FindBy* mengembalikan aggregate lengkap termasuk LoginIdentities dan Credentials.
type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByLoginIdentifier(ctx context.Context, identifier valueobject.LoginIdentifier) (*entity.User, error)
	FindByIdentity(ctx context.Context, kind constant.LoginIdentifierKind, value string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByLoginIdentity(ctx context.Context, kind constant.LoginIdentifierKind, value string) (bool, error)
	UpdateUsername(ctx context.Context, userID, newUsername string) error
}

// UserScopeRepository mengelola user_scopes — assignment scope mandiri yang tidak
// terkait role. Dipakai oleh Principal Builder untuk memuat scope user.
type UserScopeRepository interface {
	Save(ctx context.Context, scope *entity.UserScope) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.UserScope, error)
	FindByUserID(ctx context.Context, userID string) ([]*entity.UserScope, error)
}
