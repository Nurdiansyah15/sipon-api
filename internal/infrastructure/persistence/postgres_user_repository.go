package persistence

import (
	"context"
	"database/sql"
	"strings"
	"time"

	domainerr "sipon-api/internal/domain/errors"
	userConstant "sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/entity"
	"sipon-api/internal/domain/user/valueobject"
)

// ── Struct & Constructor ──────────────────────────────────────────────────────

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// ── Methods ───────────────────────────────────────────────────────────────────

func (r *PostgresUserRepository) Save(ctx context.Context, user *entity.User) error {
	var phone *string
	if user.PhoneNumber != nil {
		v := user.PhoneNumber.Value()
		phone = &v
	}

	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO users (id, username, fullname, email, phone, status, created_at, updated_at, last_login_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		user.ID, user.Username.Value(), user.Fullname, user.Email.Value(), phone,
		string(user.Status), user.CreatedAt, user.UpdatedAt, user.LastLoginAt, user.DeletedAt,
	)
	if err != nil {
		return domainerr.Wrap(userConstant.CodeUserPersistenceFailed, err)
	}

	if err := r.persistCredentials(ctx, user); err != nil {
		return err
	}
	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	row := execFromContext(ctx, r.db).QueryRowContext(ctx,
		`SELECT u.id, u.username, u.fullname, u.email, u.phone, u.status, u.created_at, u.updated_at, u.last_login_at, u.deleted_at, u.failed_login_attempts, u.locked_until
		 FROM users u
		 WHERE u.id = $1`, id)

	user, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	credentials, err := r.loadCredentialsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Credentials = credentials
	return user, nil
}

func (r *PostgresUserRepository) FindByLoginIdentifier(ctx context.Context, identifier valueobject.LoginIdentifier) (*entity.User, error) {
	return r.findByIdentityValue(ctx, identifier.Kind(), identifier.Value())
}

func (r *PostgresUserRepository) FindByIdentity(ctx context.Context, kind userConstant.LoginIdentifierKind, value string) (*entity.User, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil, domainerr.New(userConstant.CodeInvalidLoginIdentifier)
	}
	return r.findByIdentityValue(ctx, kind, v)
}

func (r *PostgresUserRepository) findByIdentityValue(ctx context.Context, kind userConstant.LoginIdentifierKind, value string) (*entity.User, error) {
	var userID string
	err := execFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT user_id
		FROM user_identities
		WHERE kind = $1 AND value = $2 AND deleted_at IS NULL
		ORDER BY is_primary DESC, updated_at DESC
		LIMIT 1`,
		string(kind), value,
	).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerr.New(userConstant.CodeLoginIdentityNotFound)
		}
		return nil, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
	}

	return r.FindByID(ctx, userID)
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	row := execFromContext(ctx, r.db).QueryRowContext(ctx,
		`SELECT u.id, u.username, u.fullname, u.email, u.phone, u.status, u.created_at, u.updated_at, u.last_login_at, u.deleted_at, u.failed_login_attempts, u.locked_until
		 FROM users u
		 WHERE u.username = $1`, username)

	user, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	credentials, err := r.loadCredentialsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Credentials = credentials
	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()
	var phone *string
	if user.PhoneNumber != nil {
		v := user.PhoneNumber.Value()
		phone = &v
	}

	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		UPDATE users SET fullname=$1, email=$2, phone=$3, status=$4, updated_at=$5, last_login_at=$6, deleted_at=$7,
		failed_login_attempts=$8, locked_until=$9 WHERE id=$10`,
		user.Fullname, user.Email.Value(), phone, string(user.Status), user.UpdatedAt, user.LastLoginAt, user.DeletedAt,
		user.FailedLoginAttempts, user.LockedUntil, user.ID,
	)
	if err != nil {
		return domainerr.Wrap(userConstant.CodeUserPersistenceFailed, err)
	}

	return r.persistCredentials(ctx, user)
}

func (r *PostgresUserRepository) UpdateUsername(ctx context.Context, userID, newUsername string) error {
	now := time.Now()
	_, err := execFromContext(ctx, r.db).ExecContext(ctx,
		`UPDATE users SET username=$1, updated_at=$2 WHERE id=$3`,
		newUsername, now, userID,
	)
	if err != nil {
		return domainerr.Wrap(userConstant.CodeUserPersistenceFailed, err)
	}
	_, err = execFromContext(ctx, r.db).ExecContext(ctx,
		`UPDATE user_identities SET value=$1, updated_at=$2 WHERE user_id=$3 AND kind='USERNAME' AND deleted_at IS NULL`,
		newUsername, now, userID,
	)
	if err != nil {
		return domainerr.Wrap(userConstant.CodeUserPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := execFromContext(ctx, r.db).QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1 AND deleted_at IS NULL)`, username).Scan(&exists)
	if err != nil {
		return false, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
	}
	return exists, nil
}

func (r *PostgresUserRepository) ExistsByLoginIdentity(ctx context.Context, kind userConstant.LoginIdentifierKind, value string) (bool, error) {
	var exists bool
	err := execFromContext(ctx, r.db).QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_identities WHERE kind = $1 AND value = $2 AND deleted_at IS NULL)`,
		string(kind), value,
	).Scan(&exists)
	if err != nil {
		return false, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
	}
	return exists, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanUser(row *sql.Row) (*entity.User, error) {
	var (
		id, username         string
		fullname             sql.NullString
		email                sql.NullString
		phone                sql.NullString
		status               string
		createdAt, updatedAt time.Time
		lastLoginAt          sql.NullTime
		deletedAt            sql.NullTime
		failedLoginAttempts  int
		lockedUntil          sql.NullTime
	)
	if err := row.Scan(&id, &username, &fullname, &email, &phone, &status, &createdAt, &updatedAt, &lastLoginAt, &deletedAt, &failedLoginAttempts, &lockedUntil); err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerr.New(userConstant.CodeUserNotFound)
		}
		return nil, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
	}
	return buildUserEntity(id, username, fullname, email, phone, status, createdAt, updatedAt, lastLoginAt, deletedAt, failedLoginAttempts, lockedUntil), nil
}

func buildUserEntity(id, username string, fullname sql.NullString, email, phone sql.NullString, status string, createdAt, updatedAt time.Time, lastLoginAt, deletedAt sql.NullTime, failedLoginAttempts int, lockedUntil sql.NullTime) *entity.User {
	uname, _ := valueobject.NewUsername(username)
	u := &entity.User{
		ID:                  id,
		Username:            uname,
		Status:              userConstant.UserStatus(status),
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		FailedLoginAttempts: failedLoginAttempts,
	}
	if fullname.Valid {
		v := fullname.String
		u.Fullname = &v
	}
	if email.Valid && email.String != "" {
		parsedEmail, err := valueobject.NewEmail(email.String)
		if err == nil {
			u.Email = parsedEmail
		}
	}
	if phone.Valid && phone.String != "" {
		parsedPhone, err := valueobject.NewPhoneNumber(phone.String)
		if err == nil {
			u.PhoneNumber = parsedPhone
		}
	}
	if lastLoginAt.Valid {
		t := lastLoginAt.Time
		u.LastLoginAt = &t
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		u.DeletedAt = &t
	}
	if lockedUntil.Valid {
		t := lockedUntil.Time
		u.LockedUntil = &t
	}
	return u
}

func (r *PostgresUserRepository) loadLoginIdentitiesByCredentialID(ctx context.Context, credentialID string) ([]*entity.LoginIdentity, error) {
	rows, err := execFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT id, user_id, credential_id, kind, value, status, is_primary, verified_at, created_at, updated_at, deleted_at
		FROM user_identities
		WHERE credential_id = $1 AND deleted_at IS NULL
		ORDER BY kind ASC, is_primary DESC, updated_at DESC`, credentialID)
	if err != nil {
		return nil, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
	}
	defer rows.Close()

	identities := make([]*entity.LoginIdentity, 0)
	for rows.Next() {
		var (
			id, uid, credID, kind, value, status string
			isPrimary                            bool
			verifiedAt                           sql.NullTime
			createdAt, updatedAt                 time.Time
			deletedAt                            sql.NullTime
		)

		if err := rows.Scan(&id, &uid, &credID, &kind, &value, &status, &isPrimary, &verifiedAt, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
		}

		identity := &entity.LoginIdentity{
			ID:           id,
			UserID:       uid,
			CredentialID: credID,
			Kind:         userConstant.LoginIdentifierKind(kind),
			Value:        value,
			Status:       userConstant.LoginIdentityStatus(status),
			IsPrimary:    isPrimary,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}
		if verifiedAt.Valid {
			t := verifiedAt.Time
			identity.VerifiedAt = &t
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			identity.DeletedAt = &t
		}
		identities = append(identities, identity)
	}

	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
	}

	return identities, nil
}

type rawCredential struct {
	id, uid, ctype string
	secret         sql.NullString
	lastChangedAt  sql.NullTime
	isPrimary      bool
	updatedAt      time.Time
	lastLoginAt    sql.NullTime
	deletedAt      sql.NullTime
}

func (r *PostgresUserRepository) loadCredentialsByUserID(ctx context.Context, userID string) ([]*entity.Credential, error) {
	rows, err := execFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT id, user_id, type, secret_hash, last_changed_at, is_primary, updated_at, last_login_at, deleted_at
		FROM credentials
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY is_primary DESC, updated_at DESC`, userID)
	if err != nil {
		return nil, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
	}

	// Buffer all rows before closing so that the subsequent login-identity
	// queries (one per credential) do not open a second cursor on the same
	// connection while this one is still active. Nested open cursors on a
	// single *sql.Tx connection cause "driver: bad connection" with pgx.
	var raws []rawCredential
	for rows.Next() {
		var rc rawCredential
		if err := rows.Scan(&rc.id, &rc.uid, &rc.ctype, &rc.secret, &rc.lastChangedAt, &rc.isPrimary, &rc.updatedAt, &rc.lastLoginAt, &rc.deletedAt); err != nil {
			rows.Close()
			return nil, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
		}
		raws = append(raws, rc)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(userConstant.CodeUserQueryFailed, err)
	}

	credentials := make([]*entity.Credential, 0, len(raws))
	for _, rc := range raws {
		cred := &entity.Credential{
			ID:        rc.id,
			UserID:    rc.uid,
			Type:      userConstant.CredentialType(rc.ctype),
			IsPrimary: rc.isPrimary,
			UpdatedAt: rc.updatedAt,
		}
		if rc.secret.Valid {
			hp, _ := valueobject.NewHashedPassword(rc.secret.String)
			cred.SecretHash = &hp
		}
		if rc.lastChangedAt.Valid {
			t := rc.lastChangedAt.Time
			cred.LastChangedAt = &t
		}
		if rc.lastLoginAt.Valid {
			t := rc.lastLoginAt.Time
			cred.LastLoginAt = &t
		}
		if rc.deletedAt.Valid {
			t := rc.deletedAt.Time
			cred.DeletedAt = &t
		}

		identities, err := r.loadLoginIdentitiesByCredentialID(ctx, rc.id)
		if err != nil {
			return nil, err
		}
		cred.LoginIdentities = identities

		credentials = append(credentials, cred)
	}
	return credentials, nil
}

func (r *PostgresUserRepository) persistCredentials(ctx context.Context, user *entity.User) error {
	for _, cred := range user.Credentials {
		var secret *string
		if cred.SecretHash != nil {
			s := cred.SecretHash.Value()
			secret = &s
		}
		_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
			INSERT INTO credentials (id, user_id, type, secret_hash, last_changed_at, is_primary, updated_at, last_login_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE SET
				secret_hash = EXCLUDED.secret_hash,
				last_changed_at = EXCLUDED.last_changed_at,
				is_primary = EXCLUDED.is_primary,
				updated_at = EXCLUDED.updated_at,
				last_login_at = EXCLUDED.last_login_at,
				deleted_at = EXCLUDED.deleted_at`,
			cred.ID, cred.UserID, string(cred.Type), secret,
			cred.LastChangedAt, cred.IsPrimary, cred.UpdatedAt, cred.LastLoginAt, cred.DeletedAt,
		)
		if err != nil {
			return domainerr.Wrap(userConstant.CodeUserPersistenceFailed, err)
		}

		if err := r.persistLoginIdentitiesForCredential(ctx, cred); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresUserRepository) persistLoginIdentitiesForCredential(ctx context.Context, cred *entity.Credential) error {
	for _, identity := range cred.LoginIdentities {
		_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
			INSERT INTO user_identities (id, user_id, credential_id, kind, value, status, is_primary, verified_at, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO UPDATE SET
				value = EXCLUDED.value,
				status = EXCLUDED.status,
				is_primary = EXCLUDED.is_primary,
				verified_at = EXCLUDED.verified_at,
				updated_at = EXCLUDED.updated_at,
				deleted_at = EXCLUDED.deleted_at`,
			identity.ID, identity.UserID, identity.CredentialID, string(identity.Kind), identity.Value,
			string(identity.Status), identity.IsPrimary,
			identity.VerifiedAt, identity.CreatedAt, identity.UpdatedAt, identity.DeletedAt,
		)
		if err != nil {
			return domainerr.Wrap(userConstant.CodeUserPersistenceFailed, err)
		}
	}
	return nil
}
