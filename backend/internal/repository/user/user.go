package user

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/cerr"
	"backend/pkg/postgres"
	"context"
)

type RepoUser struct {
	db *postgres.Pg
}

func InitUserRepository(db *postgres.Pg) repository.UserRepo {
	return RepoUser{db: db}
}

func (r RepoUser) Create(ctx context.Context, user models.UserCreate) (int, error) {
	var id int
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return 0, cerr.Transaction(err)
	}
	row := tx.QueryRow(ctx, `INSERT INTO users (name, sur_name, email, hashed_password) VALUES ($1, $2, $3, $4) returning id;`,
		user.Name, user.SurName, user.Email, user.PWD)

	err = row.Scan(&id)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return 0, cerr.Rollback(rbErr)
		}
		return 0, cerr.Scan(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, cerr.Commit(err)
	}
	return id, nil
}

func (r RepoUser) Get(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	row := r.db.Pool.QueryRow(ctx, `SELECT name, sur_name, email from users WHERE id = $1;`, id)
	err := row.Scan(&user.Name, &user.SurName, &user.Email)
	if err != nil {
		return nil, cerr.Scan(err)
	}
	user.ID = id
	return &user, nil
}

func (r RepoUser) GetAll(ctx context.Context) ([]models.User, error) {
	var users []models.User
	rows, err := r.db.Pool.Query(ctx, `SELECT id, name, sur_name, email from users;`)
	if err != nil {
		return nil, cerr.Execution(err)
	}
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Name, &user.SurName, &user.Email)
		if err != nil {
			return nil, cerr.Scan(err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (r RepoUser) GetPWDbyEmail(ctx context.Context, user string) (int, string, error) {
	var pwd string
	var id int
	row := r.db.Pool.QueryRow(ctx, `SELECT id,  hashed_password from users WHERE email = $1;`, user)
	err := row.Scan(&id, &pwd)
	if err != nil {
		return 0, "", cerr.Scan(err)
	}
	return id, pwd, nil
}

func (r RepoUser) ChangePWD(ctx context.Context, user models.UserChangePWD) (int, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return 0, cerr.Transaction(err)
	}
	result, err := tx.Exec(ctx, `UPDATE users SET hashed_password=$2 WHERE id=$1;`, user.ID, user.NewPWD)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return 0, cerr.Rollback(rbErr)
		}
		return 0, cerr.ExecContext(err)
	}
	count := result.RowsAffected()

	if count != 1 {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return 0, cerr.Rollback(rbErr)
		}
		return 0, cerr.NoOneRow(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, cerr.Commit(err)
	}
	return user.ID, nil
}

func (r RepoUser) Delete(ctx context.Context, id int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return cerr.Transaction(err)
	}
	result, err := tx.Exec(ctx, `DELETE FROM users WHERE id=$1;`, id)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return cerr.Rollback(rbErr)
		}
		return cerr.ExecContext(err)
	}
	count := result.RowsAffected()
	if count != 1 {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return cerr.Rollback(rbErr)
		}
		return cerr.NoOneRow(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return cerr.Commit(err)
	}
	return nil
}
