package film

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"github.com/lib/pq"
)

func TestGetFilmsByGenre(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Title", "Poster"})

	expect := []models.FilmItem{
		{Id: 1, Title: "t1", Poster: "url1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Id, item.Title, item.Poster)
	}

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT film.id, film.title, poster FROM film  JOIN films_genre ON film.id = films_genre.id_film WHERE id_genre = $1 ORDER BY release_date DESC OFFSET $2 LIMIT $3")).
		WithArgs(1, 1, 2).
		WillReturnRows(rows)

	repo := &RepoPostgre{
		db: db,
	}

	films, err := repo.GetFilmsByGenre(1, 1, 2)
	if err != nil {
		t.Errorf("GetFilmsByGenre error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(films, expect) {
		t.Errorf("results not match, want %v, have %v", expect, films)
		return
	}

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT film.id, film.title, poster FROM film  JOIN films_genre ON film.id = films_genre.id_film WHERE id_genre = $1 ORDER BY release_date DESC OFFSET $2 LIMIT $3")).
		WithArgs(1, 1, 2).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.GetFilmsByGenre(1, 1, 2)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestGetFilms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Title", "Poster"})

	expect := []models.FilmItem{
		{Id: 1, Title: "t1", Poster: "url1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Id, item.Title, item.Poster)
	}

	mock.ExpectQuery("SELECT film.id, film.title, poster FROM film").WithArgs(1, 2).WillReturnRows(rows)

	repo := &RepoPostgre{
		db: db,
	}

	films, err := repo.GetFilms(1, 2)
	if err != nil {
		t.Errorf("GetFilms error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(films, expect) {
		t.Errorf("results not match, want %v, have %v", expect, films)
		return
	}

	mock.
		ExpectQuery("SELECT film.id, film.title, poster FROM film").
		WithArgs(1, 2).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.GetFilms(1, 2)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestGetFilm(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Title", "Info", "Poster", "ReleaseDate", "Country", "Mpaa"})

	expect := []models.FilmItem{
		{Id: 1, Title: "t1", Info: "i1", Poster: "url1", ReleaseDate: "date1", Country: "c1", Mpaa: "12"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Id, item.Title, item.Info, item.Poster, item.ReleaseDate, item.Country, item.Mpaa)
	}

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT * FROM film WHERE id = $1")).
		WithArgs(1).
		WillReturnRows(rows)

	repo := &RepoPostgre{
		db: db,
	}

	films, err := repo.GetFilm(1)
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(films, &expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect, films)
		return
	}

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT * FROM film WHERE id = $1")).
		WithArgs(1).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.GetFilm(1)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestGetFilmRating(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Average", "Amount"})

	expectRating := 4.2
	expectAmount := uint64(3)

	rows = rows.AddRow(expectRating, expectAmount)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT AVG(rating), COUNT(rating) FROM users_comment WHERE id_film")).
		WithArgs(1).
		WillReturnRows(rows)

	repo := &RepoPostgre{
		db: db,
	}

	rating, number, err := repo.GetFilmRating(1)
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if rating != expectRating {
		t.Errorf("results not match, want %v, have %v", expectRating, rating)
		return
	}
	if number != expectAmount {
		t.Errorf("results not match, want %v, have %v", expectAmount, number)
	}

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT AVG(rating), COUNT(rating) FROM users_comment WHERE id_film")).
		WithArgs(1).
		WillReturnError(fmt.Errorf("db_error"))

	rating, number, err = repo.GetFilmRating(1)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if rating != 0 {
		t.Errorf("expected rating 0, got %f", rating)
	}
	if number != 0 {
		t.Errorf("expected number 0, got %d", number)
	}
}

func TestFindFilm(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Title", "Id", "Poster", "Rating"})

	expectFilm := []models.FilmItem{
		{Id: 1, Title: "t1", Poster: "url1"},
	}
	expectRating := []float32{8}

	for _, item := range expectFilm {
		rows = rows.AddRow(item.Title, item.Id, item.Poster, expectRating[0])
	}

	selectStr := "SELECT DISTINCT film.title, film.id, film.poster, AVG(users_comment.rating) FROM film JOIN films_genre ON film.id = films_genre.id_film JOIN genre ON genre.id = films_genre.id_genre JOIN users_comment ON film.id = users_comment.id_film JOIN person_in_film ON film.id = person_in_film.id_film JOIN crew ON person_in_film.id_person = crew.id WHERE (CASE WHEN array_length($1::varchar[], 1)> 0 THEN genre.title = ANY ($1::varchar[]) ELSE TRUE END) AND (CASE WHEN array_length($2::varchar[], 1)> 0 THEN crew.name = ANY ($2::varchar[]) ELSE TRUE END) GROUP BY film.title, film.id, genre.title HAVING AVG(users_comment.rating) > $3 AND AVG(users_comment.rating) < $4 ORDER BY film.title"
	mock.ExpectQuery(
		regexp.QuoteMeta(selectStr)).
		WithArgs(pq.Array([]string{}), pq.Array([]string{}), float32(0), float32(10), "", "", "", "").
		WillReturnRows(rows)

	repo := &RepoPostgre{
		db: db,
	}

	film, err := repo.FindFilm("", "", "", float32(0), float32(10), "", []string{}, []string{})
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if reflect.DeepEqual(film, expectFilm) {
		t.Errorf("film results not match, want %v, have %v", expectFilm, film)
	}

	mock.ExpectQuery(
		regexp.QuoteMeta(selectStr)).
		WithArgs("", "", "", float32(0), float32(10), "", []string{}, []string{}).
		WillReturnError(fmt.Errorf("db_error"))

	film, err = repo.FindFilm("", "", "", float32(0), float32(10), "", []string{}, []string{})
	if err == mock.ExpectationsWereMet() {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	if film != nil {
		t.Errorf("expected film nil, got %v", film)
	}
}
