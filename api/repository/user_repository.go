package repository

import (
	"database/sql"
	"errors"
	"log"
	"math"
	"noda/api/data/model"
	"noda/api/data/transfer"
	"noda/failure"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) Insert(next *transfer.UserCreation) (*transfer.User, error) {
	if yes, err := r.ExistsUserWithEmail(next.Email); err != nil {
		return nil, err
	} else if yes {
		return nil, failure.ErrSameEmail
	}

	query := `
	INSERT INTO "user" ("first_name", "middle_name", "last_name", "surname", "email", "password")
	     VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING "user_id" AS "id",
		          "first_name",
		          "middle_name",
		          "last_name",
		          "surname",
		          "picture_url",
		          "email",
		          "created_at",
		          "updated_at";`
	row, err := r.db.Query(query,
		next.FirstName, next.MiddleName, next.LastName, next.Surname, next.Email, next.Password)
	if err != nil {
		var pqerr *pq.Error
		switch {
		default:
			log.Println(err)
		case errors.As(err, &pqerr):
			log.Println(failure.PQErrorToString(pqerr))
		}
		return nil, err
	}
	defer row.Close()
	user := transfer.User{}
	if err = sqlscan.ScanOne(&user, row); err != nil {
		log.Println(err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(userID string, up *transfer.UserUpdate) (bool, error) {
	if actual, err := r.SelectByID(userID); err != nil {
		return false, err
	} else if actual.FirstName == up.FirstName &&
		actual.MiddleName == up.MiddleName &&
		actual.LastName == up.LastName &&
		actual.Surname == up.Surname {
		return false, nil
	}
	query := `
	   UPDATE "user"
	      SET "first_name" = COALESCE(NULLIF(trim($2), ''), "first_name"),
	          "middle_name" = COALESCE(NULLIF(trim($3), ''), "middle_name"),
	          "last_name" = COALESCE(NULLIF(trim($4), ''), "last_name"),
	          "surname" = COALESCE(NULLIF(trim($5), ''), "surname"),
						"updated_at" = 'now()'
			WHERE "user_id" = $1;`
	result, err := r.db.Exec(query, &userID, &up.FirstName, &up.MiddleName, &up.LastName, &up.Surname)
	if err != nil {
		var pqerr *pq.Error
		switch {
		default:
			log.Println(err)
		case errors.As(err, &pqerr):
			log.Println(failure.PQErrorToString(pqerr))
		}
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		return false, err
	}
	return count >= 1, nil
}

func (r *UserRepository) ExistsUserWithEmail(email string) (bool, error) {
	// TODO: Create an index on email.
	query := `
	SELECT "user_id"
	  FROM "user"
	 WHERE lower("email") = lower($1);`
	result, err := r.db.Exec(query, &email)
	if err != nil {
		var pqerr *pq.Error
		switch {
		default:
			log.Println(err)
		case errors.As(err, &pqerr):
			log.Println(failure.PQErrorToString(pqerr))
		}
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		return false, err
	}
	return count >= 1, nil
}

func (r *UserRepository) SelectAll(limit, page int64) (*[]*transfer.User, error) {
	maxValidBeforeOverflow := (math.MaxInt64 / limit) - 1
	if page > maxValidBeforeOverflow {
		page = maxValidBeforeOverflow
	}
	query := `
	SELECT "user_id" AS "id",
	       "first_name",
	       "middle_name",
	       "last_name",
	       "surname",
	       "picture_url",
	       "email",
	       "created_at",
	       "updated_at"
	  FROM "user"
ORDER BY "created_at" DESC
   LIMIT $1
	OFFSET ($1 * ($2::BIGINT - 1));`
	rows, err := r.db.Query(query, &limit, &page)
	if err != nil {
		var pqerr *pq.Error
		switch {
		default:
			log.Println(err)
		case errors.As(err, &pqerr):
			log.Println(failure.PQErrorToString(pqerr))
		}
		return nil, err
	}
	defer rows.Close()
	users := []*transfer.User{}
	if err = sqlscan.ScanAll(&users, rows); err != nil {
		log.Println(err)
		return nil, err
	}
	return &users, nil
}

func (r *UserRepository) SelectByEmail(email string) (*transfer.User, error) {
	user, err := r.SelectWithPasswordByEmail(email)
	if err != nil {
		return nil, err
	}
	return &transfer.User{
		ID:         user.ID,
		FirstName:  user.FirstName,
		MiddleName: user.MiddleName,
		LastName:   user.LastName,
		Surname:    user.Surname,
		PictureUrl: user.PictureUrl,
		Email:      user.Email,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

func (r *UserRepository) SelectByID(id string) (*transfer.User, error) {
	query := `
	SELECT "user_id" AS "id",
	       "first_name",
	       "middle_name",
	       "last_name",
	       "surname",
	       "picture_url",
	       "email",
	       "created_at",
	       "updated_at"
	  FROM "user"
	 WHERE "user_id" = $1;`
	row, err := r.db.Query(query, &id)
	if err != nil {
		var pqerr *pq.Error
		switch {
		default:
			log.Println(err)
		case errors.As(err, &pqerr):
			log.Println(failure.PQErrorToString(pqerr))
		}
		return nil, err
	}
	defer row.Close()
	user := transfer.User{}
	if err := sqlscan.ScanOne(&user, row); err != nil {
		switch {
		default:
			log.Println(err)
			return nil, err
		case sqlscan.NotFound(err):
			return nil, failure.ErrNotFound
		}
	}
	return &user, nil
}

func (r *UserRepository) SelectWithPasswordByEmail(email string) (*model.User, error) {
	query := `
	SELECT "user_id" AS "id",
	       "first_name",
	       "middle_name",
	       "last_name",
	       "surname",
	       "picture_url",
	       "email",
				 "password",
	       "created_at",
	       "updated_at"
	  FROM "user"
	 WHERE lower("email") = lower($1);`

	row, err := r.db.Query(query, &email)
	if err != nil {
		var pqerr *pq.Error
		switch {
		default:
			log.Println(err)
		case errors.As(err, &pqerr):
			log.Println(failure.PQErrorToString(pqerr))
		}
		return nil, err
	}
	defer row.Close()

	user := model.User{}
	if err := sqlscan.ScanOne(&user, row); err != nil {
		switch {
		default:
			log.Println(err)
			return nil, err
		case sqlscan.NotFound(err):
			return nil, failure.ErrNotFound
		}
	}
	return &user, nil
}

func (r *UserRepository) Delete(id string) (string, error) {
	var (
		query = `
		DELETE FROM "user"
					WHERE "user_id" = $1
			RETURNING "user_id";`
		row           = r.db.QueryRow(query, id)
		deletedUserID = ""
	)
	if err := row.Scan(&deletedUserID); err != nil {
		var pqerr *pq.Error
		switch {
		default:
			log.Println(err)
			return "", err
		case errors.As(err, &pqerr):
			log.Println(failure.PQErrorToString(pqerr))
			return "", err
		case errors.Is(err, sql.ErrNoRows):
			return "", failure.ErrNotFound
		}
	}
	return deletedUserID, nil
}
