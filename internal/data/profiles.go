package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	ProfileUser    uuid.UUID
	ProfileName    string
	ProfilePicture string
	IsIndexed      bool
	IsDeleted      bool
	Version        int
}

type ProfileModel struct {
	DB *sql.DB
}

func (m ProfileModel) Get(id uuid.UUID) (*Profile, error) {
	query := `
        SELECT id, created_at, profile_user, profile_name,
					profile_picture, is_indexed, is_deleted, version
        FROM profiles
        WHERE id = $1`

	var profile Profile

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&profile.ID,
		&profile.CreatedAt,
		&profile.ProfileUser,
		&profile.ProfileName,
		&profile.ProfilePicture,
		&profile.IsIndexed,
		&profile.IsDeleted,
		&profile.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &profile, nil
}

func (m ProfileModel) IsIndexedTrue(profile *Profile) error {
	query := `
        UPDATE profile
        SET is_indexed = true
				WHERE id = $1 AND version = $2
        RETURNING version`

	args := []interface{}{
		profile.ID,
		profile.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&profile.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m ProfileModel) Delete(profile *Profile) error {
	query := `
        DELETE FROM profiles
				WHERE id = $1`

	args := []interface{}{
		profile.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
