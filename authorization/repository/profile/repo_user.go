package profile

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	_ "github.com/jackc/pgx/stdlib"
)

type IUserRepo interface {
	GetUser(login string, password string) (*UserItem, bool, error)
	GetUserProfileId(login string) (int64, error)
	FindUser(login string) (bool, error)
	CreateUser(login string, password string, name string, birthDate string, email string) error
	GetUserProfile(login string) (*UserItem, error)
	EditProfile(prevLogin string, login string, password string, email string, birthDate string, photo string) error
	GetIdsAndPaths() ([]int32, []string, error) 
}

type RepoPostgre struct {
	db *sql.DB
}

func GetUserRepo(config *configs.DbDsnCfg, lg *slog.Logger) (*RepoPostgre, error) {
	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		config.User, config.DbName, config.Password, config.Host, config.Port, config.Sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		lg.Error("sql open error", "err", err.Error())
		return nil, fmt.Errorf("get user repo err: %w", err)
	}
	err = db.Ping()
	if err != nil {
		lg.Error("sql ping error", "err", err.Error())
		return nil, fmt.Errorf("get user repo err: %w", err)
	}
	db.SetMaxOpenConns(config.MaxOpenConns)

	postgreDb := RepoPostgre{db: db}

	go postgreDb.pingDb(config.Timer, lg)
	return &postgreDb, nil
}

func (repo *RepoPostgre) pingDb(timer uint32, lg *slog.Logger) {
	for {
		err := repo.db.Ping()
		if err != nil {
			lg.Error("Repo Profile db ping error", "err", err.Error())
		}

		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func (repo *RepoPostgre) GetUser(login string, password string) (*UserItem, bool, error) {
	post := &UserItem{}

	err := repo.db.QueryRow(
		"SELECT login, photo FROM profile "+
			"WHERE login = $1 AND password = $2", login, password).Scan(&post.Login, &post.Photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("GetUser err: %w", err)
	}

	return post, true, nil
}

func (repo *RepoPostgre) FindUser(login string) (bool, error) {
	post := &UserItem{}

	err := repo.db.QueryRow(
		"SELECT login FROM profile "+
			"WHERE login = $1", login).Scan(&post.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("FindUser err: %w", err)
	}

	return true, nil
}

func (repo *RepoPostgre) GetUserProfileId(login string) (int64, error) {
    var userID int64

    err := repo.db.QueryRow(
        "SELECT id FROM profile WHERE login = $1", login).Scan(&userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return 0, fmt.Errorf("User not found for login: %s", login)
        }
        return 0, fmt.Errorf("GetUserProfileID error: %w", err)
    }

    return userID, nil
}

func (repo *RepoPostgre) CreateUser(login string, password string, name string, birthDate string, email string) error {
	_, err := repo.db.Exec(
		"INSERT INTO profile(name, birth_date, photo, login, password, email, registration_date) "+
			"VALUES($1, $2, '/avatars/default.jpg', $3, $4, $5, CURRENT_TIMESTAMP)",
		name, birthDate, login, password, email)
	if err != nil {
		return fmt.Errorf("CreateUser err: %w", err)
	}

	return nil
}

func (repo *RepoPostgre) GetIdsAndPaths() ([]int32, []string, error) {
    rows, err := repo.db.Query("SELECT id, photo FROM profile")
    if err != nil {
        return nil, nil, fmt.Errorf("GetIdsAndPaths query error: %w", err)
    }
    defer rows.Close()

    var ids []int32
    var paths []string

    for rows.Next() {
        var id int32
        var path string
        if err := rows.Scan(&id, &path); err != nil {
            return nil, nil, fmt.Errorf("GetIdsAndPaths scan error: %w", err)
        }
        ids = append(ids, id)
        paths = append(paths, path)
    }

    if err := rows.Err(); err != nil {
        return nil, nil, fmt.Errorf("GetIdsAndPaths rows error: %w", err)
    }

    return ids, paths, nil
}


func (repo *RepoPostgre) GetUserProfile(login string) (*UserItem, error) {
	post := &UserItem{}

	err := repo.db.QueryRow(
		"SELECT name, birth_date, login, email, photo FROM profile "+
			"WHERE login = $1", login).Scan(&post.Name, &post.Birthdate, &post.Login, &post.Email, &post.Photo)
	if err != nil {
		return nil, fmt.Errorf("GetUserProfile err: %w", err)
	}

	return post, nil
}

func (repo *RepoPostgre) EditProfile(prevLogin string, login string, password string, email string, birthDate string, photo string) error {
	if photo == "" {
		_, err := repo.db.Exec("UPDATE profile "+
			"SET login = $1, email = $2, birth_date = $3 "+
			"WHERE login = $4", login, email, birthDate, prevLogin)
		if err != nil {
			return fmt.Errorf("failed to edit profile in db: %w", err)
		}
		return nil

	}

	if password == "" {
		_, err := repo.db.Exec("UPDATE profile "+
			"SET login = $1, photo = $2, email = $3, birth_date = $4 "+
			"WHERE login = $5", login, photo, email, birthDate, prevLogin)
		if err != nil {
			return fmt.Errorf("failed to edit profile in db: %w", err)
		}
		return nil
	}

	_, err := repo.db.Exec("UPDATE profile "+
		"SET login = $1, password = $2, photo = $3, email = $4, birth_date = $5 "+
		"WHERE login = $6", login, password, photo, email, birthDate, prevLogin)
	if err != nil {
		return fmt.Errorf("failed to edit profile in db: %w", err)
	}

	return nil
}
