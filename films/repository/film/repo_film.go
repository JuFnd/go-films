package film

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"github.com/lib/pq"

	_ "github.com/jackc/pgx/stdlib"
)

type IFilmsRepo interface {
	GetFilmsByGenre(genre uint64, start uint64, end uint64) ([]models.FilmItem, error)
	GetFilms(start uint64, end uint64) ([]models.FilmItem, error)
	GetFilm(filmId uint64) (*models.FilmItem, error)
	GetFilmRating(filmId uint64) (float64, uint64, error)
	GetFavoriteFilms(userId uint64) ([]models.FilmItem, error)
	AddFavoriteFilm(userId uint64, filmId uint64) error
	RemoveFavoriteFilm(userId uint64, filmId uint64) error
	FindByActorsFilm(actors []string) ([]models.FilmItem, error)
	FindByGenresFilm(genres []string) ([]models.FilmItem, error)
	FindByRatingFilm(ratingFrom float64, ratingTo float64) ([]models.FilmItem, error)
	FindByDateFilm(dateFrom string, dateTo string) ([]models.FilmItem, error)
	FindByTitleFilm(title string) ([]models.FilmItem, error)
}

type RepoPostgre struct {
	db *sql.DB
}

func GetFilmRepo(config configs.DbDsnCfg, lg *slog.Logger) (*RepoPostgre, error) {
	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		config.User, config.DbName, config.Password, config.Host, config.Port, config.Sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		lg.Error("sql open error", "err", err.Error())
		return nil, fmt.Errorf("get film repo: %w", err)
	}
	err = db.Ping()
	if err != nil {
		lg.Error("sql ping error", "err", err.Error())
		return nil, fmt.Errorf("get film repo: %w", err)
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
			lg.Error("Repo Film db ping error", "err", err.Error())
		}

		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func (repo *RepoPostgre) GetFilmsByGenre(genre uint64, start uint64, end uint64) ([]models.FilmItem, error) {
	films := make([]models.FilmItem, 0, end-start)

	fmt.Println("in the get films by genre")
	rows, err := repo.db.Query(
		"SELECT film.id, film.title, poster FROM film "+
			"JOIN films_genre ON film.id = films_genre.id_film "+
			"WHERE id_genre = $1 "+
			"ORDER BY release_date DESC "+
			"OFFSET $2 LIMIT $3",
		genre, start, end)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetFilmsByGenre err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		post := models.FilmItem{}
		err := rows.Scan(&post.Id, &post.Title, &post.Poster)
		if err != nil {
			return nil, fmt.Errorf("GetFilmsByGenre scan err: %w", err)
		}
		films = append(films, post)
	}

	return films, nil
}

func (repo *RepoPostgre) GetFilms(start uint64, end uint64) ([]models.FilmItem, error) {
	fmt.Println("in the GetFilms")

	films := make([]models.FilmItem, 0, end-start)

	fmt.Println("start select")

	rows, err := repo.db.Query(
		"SELECT film.id, film.title, poster FROM film "+
			"ORDER BY release_date DESC "+
			"OFFSET $1 LIMIT $2",
		start, end)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fmt.Println("select error")
		return films, fmt.Errorf("GetFilms err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		post := models.FilmItem{}
		err := rows.Scan(&post.Id, &post.Title, &post.Poster)
		if err != nil {
			return films, fmt.Errorf("GetFilms scan err: %w", err)
		}
		films = append(films, post)
	}
	fmt.Println("NO ERRORS")
	return films, nil
}

func (repo *RepoPostgre) GetFilm(filmId uint64) (*models.FilmItem, error) {
	film := &models.FilmItem{}
	err := repo.db.QueryRow(
		"SELECT * FROM film "+
			"WHERE id = $1", filmId).
		Scan(&film.Id, &film.Title, &film.Info, &film.Poster, &film.ReleaseDate, &film.Country, &film.Mpaa)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return film, nil
		}

		return nil, fmt.Errorf("GetFilm err: %w", err)
	}

	return film, nil
}

func (repo *RepoPostgre) GetFilmRating(filmId uint64) (float64, uint64, error) {
	var rating sql.NullFloat64
	var number sql.NullInt64
	err := repo.db.QueryRow(
		"SELECT AVG(rating), COUNT(rating) FROM users_comment "+
			"WHERE id_film = $1", filmId).Scan(&rating, &number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("GetFilmRating err: %w", err)
	}

	return rating.Float64, uint64(number.Int64), nil
}

func (repo *RepoPostgre) FindByActorsFilm(actors []string) ([]models.FilmItem, error) {
	query := `
		SELECT film.id, film.title, film.poster
		FROM film
		JOIN person_in_film ON film.id = person_in_film.id_film
		JOIN crew ON person_in_film.id_person = crew.id
		WHERE crew.name = ANY($1)
	`
	rows, err := repo.db.Query(query, pq.Array(actors))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find by actors film error: %w", err)
	}
	defer rows.Close()

	var films []models.FilmItem
	for rows.Next() {
		var film models.FilmItem
		err := rows.Scan(&film.Id, &film.Title, &film.Poster)
		if err != nil {
			return nil, fmt.Errorf("find by actors film scan error: %w", err)
		}
		films = append(films, film)
	}

	return films, nil
}

func (repo *RepoPostgre) FindByGenresFilm(filmGenres []string) ([]models.FilmItem, error) {
	query := `
		SELECT film.id, film.title, film.poster
		FROM film
		JOIN films_genre ON film.id = films_genre.id_film
		JOIN genre ON genre.id = films_genre.id_genre
		WHERE genre.title = ANY($1)
	`
	rows, err := repo.db.Query(query, pq.Array(filmGenres))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find by genres film error: %w", err)
	}
	defer rows.Close()

	var films []models.FilmItem
	for rows.Next() {
		var film models.FilmItem
		err := rows.Scan(&film.Id, &film.Title, &film.Poster)
		if err != nil {
			return nil, fmt.Errorf("find by genres film scan error: %w", err)
		}
		films = append(films, film)
	}

	return films, nil
}

func (repo *RepoPostgre) FindByRatingFilm(ratingFrom float64, ratingTo float64) ([]models.FilmItem, error) {
	query := `
		SELECT film.id, film.title, film.poster
		FROM film
		JOIN users_comment ON film.id = users_comment.id_film
		GROUP BY film.id
		HAVING AVG(users_comment.rating) >= $1 AND AVG(users_comment.rating) <= $2
	`
	rows, err := repo.db.Query(query, ratingFrom, ratingTo)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find by rating film error: %w", err)
	}
	defer rows.Close()

	var films []models.FilmItem
	for rows.Next() {
		var film models.FilmItem
		err := rows.Scan(&film.Id, &film.Title, &film.Poster)
		if err != nil {
			return nil, fmt.Errorf("find by rating film scan error: %w", err)
		}
		films = append(films, film)
	}

	return films, nil
}

func (repo *RepoPostgre) FindByDateFilm(dateFrom string, dateTo string) ([]models.FilmItem, error) {
	query := `
		SELECT film.id, film.title, film.poster
		FROM film
		WHERE release_date >= $1 AND release_date <= $2
	`
	rows, err := repo.db.Query(query, dateFrom, dateTo)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find by date film error: %w", err)
	}
	defer rows.Close()

	var films []models.FilmItem
	for rows.Next() {
		var film models.FilmItem
		err := rows.Scan(&film.Id, &film.Title, &film.Poster)
		if err != nil {
			return nil, fmt.Errorf("find by date film scan error: %w", err)
		}
		films = append(films, film)
	}

	return films, nil
}

func (repo *RepoPostgre) FindByTitleFilm(title string) ([]models.FilmItem, error) {
	query := `
		SELECT film.id, film.title, film.poster
		FROM film
		WHERE film.title ILIKE '%' || $1 || '%'
	`
	rows, err := repo.db.Query(query, title)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find by title film error: %w", err)
	}
	defer rows.Close()

	var films []models.FilmItem
	for rows.Next() {
		var film models.FilmItem
		err := rows.Scan(&film.Id, &film.Title, &film.Poster)
		if err != nil {
			return nil, fmt.Errorf("find by title film scan error: %w", err)
		}
		films = append(films, film)
	}

	return films, nil
}

func (repo *RepoPostgre) GetFavoriteFilms(userId uint64) ([]models.FilmItem, error) {
	films := []models.FilmItem{}

	rows, err := repo.db.Query(
		"SELECT film.title, film.id, film.poster FROM film "+
			"JOIN users_favorite_film ON film.id = users_favorite_film.id_film "+
			"WHERE id_user = $1", userId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get favorite films err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		post := models.FilmItem{}
		err := rows.Scan(&post.Id, &post.Title, &post.Poster)
		if err != nil {
			return nil, fmt.Errorf("get favorite films scan err: %w", err)
		}
		films = append(films, post)
	}

	return films, nil
}

func (repo *RepoPostgre) AddFavoriteFilm(userId uint64, filmId uint64) error {
	_, err := repo.db.Exec(
		"INSERT INTO users_favorite_film(id_user, id_film) VALUES ($1, $2)", userId, filmId)
	if err != nil {
		return fmt.Errorf("add favorite film err: %w", err)
	}

	return nil
}

func (repo *RepoPostgre) RemoveFavoriteFilm(userId uint64, filmId uint64) error {
	_, err := repo.db.Exec(
		"DELETE FROM users_favorite_film "+
			"WHERE id_user = $1 AND id_film = $2", userId, filmId)
	if err != nil {
		return fmt.Errorf("remove favorite film err: %w", err)
	}

	return nil
}
