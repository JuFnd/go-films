package crew

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"

	_ "github.com/jackc/pgx/stdlib"
)

type ICrewRepo interface {
	GetFilmDirectors(filmId uint64) ([]models.CrewItem, error)
	GetFilmScenarists(filmId uint64) ([]models.CrewItem, error)
	GetFilmCharacters(filmId uint64) ([]models.Character, error)
	GetActor(actorId uint64) (*models.CrewItem, error)
}

type RepoPostgre struct {
	db *sql.DB
}

func GetCrewRepo(config configs.DbDsnCfg, lg *slog.Logger) (*RepoPostgre, error) {
	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		config.User, config.DbName, config.Password, config.Host, config.Port, config.Sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		lg.Error("sql open error", "err", err.Error())
		return nil, fmt.Errorf("get crew repo: %w", err)
	}
	err = db.Ping()
	if err != nil {
		lg.Error("sql ping error", "err", err.Error())
		return nil, fmt.Errorf("get crew repo: %w", err)
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
			lg.Error("Repo Crew db ping error", "err", err.Error())
		}

		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func (repo *RepoPostgre) GetFilmDirectors(filmId uint64) ([]models.CrewItem, error) {
	directors := []models.CrewItem{}

	rows, err := repo.db.Query(
		"SELECT crew.id, name, photo  FROM crew "+
			"JOIN person_in_film ON crew.id = person_in_film.id_person "+
			"WHERE id_film = $1 AND id_profession = "+
			"(SELECT id FROM profession WHERE title = 'режиссёр')", filmId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetFilmDirectors err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		post := models.CrewItem{}
		err := rows.Scan(&post.Id, &post.Name, &post.Photo)
		if err != nil {
			return nil, fmt.Errorf("GetFilmDirectors scan err: %w", err)
		}
		directors = append(directors, post)
	}

	return directors, nil
}

func (repo *RepoPostgre) GetFilmScenarists(filmId uint64) ([]models.CrewItem, error) {
	scenarists := []models.CrewItem{}

	rows, err := repo.db.Query(
		"SELECT crew.id, name, photo  FROM crew "+
			"JOIN person_in_film ON crew.id = person_in_film.id_person "+
			"WHERE id_film = $1 AND id_profession = "+
			"(SELECT id FROM profession WHERE title = 'сценарист')", filmId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetFilmScenarists err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		post := models.CrewItem{}
		err := rows.Scan(&post.Id, &post.Name, &post.Photo)
		if err != nil {
			return nil, fmt.Errorf("GetFilmScenarists scan err: %w", err)
		}
		scenarists = append(scenarists, post)
	}

	return scenarists, nil
}

func (repo *RepoPostgre) GetFilmCharacters(filmId uint64) ([]models.Character, error) {
	characters := []models.Character{}

	rows, err := repo.db.Query(
		"SELECT crew.id, name, photo, person_in_film.character_name FROM crew "+
			"JOIN person_in_film ON crew.id = person_in_film.id_person "+
			"WHERE id_film = $1 AND id_profession = "+
			"(SELECT id FROM profession WHERE title = 'актёр')", filmId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetFilmCharacters err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		post := models.Character{}
		err := rows.Scan(&post.IdActor, &post.NameActor, &post.ActorPhoto, &post.NameCharacter)
		if err != nil {
			return nil, fmt.Errorf("GetFilmCharacters scan err: %w", err)
		}
		characters = append(characters, post)
	}

	return characters, nil
}

func (repo *RepoPostgre) GetActor(actorId uint64) (*models.CrewItem, error) {
	actor := &models.CrewItem{}

	err := repo.db.QueryRow(
		"SELECT id, name, birth_date, photo, info FROM crew "+
			"WHERE id = $1", actorId).
		Scan(&actor.Id, &actor.Name, &actor.Birthdate, &actor.Photo, &actor.Info)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return actor, nil
		}
		return nil, fmt.Errorf("GetActor err: %w", err)
	}

	return actor, nil
}
