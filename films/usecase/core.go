package usecase

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/calendar"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/crew"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/film"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/genre"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/profession"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
)

var ErrNotFound = errors.New("not found")

type ICore interface {
	GetFilmsAndGenreTitle(genreId uint64, start uint64, end uint64) ([]models.FilmItem, string, error)
	GetFilmInfo(filmId uint64) (*requests.FilmResponse, error)
	GetActorInfo(actorId uint64) (*requests.ActorResponse, error)
	GetActorsCareer(actorId uint64) ([]models.ProfessionItem, error)
	GetGenre(genreId uint64) (string, error)
	FindFilm(title string, dateFrom string, dateTo string,
		ratingFrom float32, ratingTo float32, mpaa string, genres []string, actors []string,
	) ([]models.FilmItem, error)
	FavoriteFilms(userId uint64) ([]models.FilmItem, error)
	FavoriteFilmsAdd(userId uint64, filmId uint64) error
	FavoriteFilmsRemove(userId uint64, filmId uint64) error
	GetCalendar() (*requests.CalendarResponse, error)
}

type Core struct {
	lg         *slog.Logger
	films      film.IFilmsRepo
	genres     genre.IGenreRepo
	crew       crew.ICrewRepo
	profession profession.IProfessionRepo
	calendar   calendar.ICalendarRepo
}

func GetCore(cfg_sql *configs.DbDsnCfg, lg *slog.Logger, films film.IFilmsRepo, genres genre.IGenreRepo, actors crew.ICrewRepo, professions profession.IProfessionRepo) *Core {
	core := Core{
		lg:         lg.With("module", "core"),
		films:      films,
		genres:     genres,
		crew:       actors,
		profession: professions,
	}
	return &core
}

func (core *Core) FindFilm(title string, dateFrom string, dateTo string,
	ratingFrom float32, ratingTo float32, mpaa string, genres []string, actors []string,
) ([]models.FilmItem, error) {
	films, err := core.films.FindFilm(title, dateFrom, dateTo, ratingFrom, ratingTo, mpaa, genres, actors)
	if err != nil {
		core.lg.Error("find film error", "err", err.Error())
		return nil, fmt.Errorf("find film err: %w", err)
	}

	return films, nil
}

func (core *Core) GetFilmsAndGenreTitle(genreId uint64, start uint64, end uint64) ([]models.FilmItem, string, error) {
	films := make([]models.FilmItem, 0, end-start)

	var err error

	if genreId == 0 {
		fmt.Println("go to GetFilms", start, end)
		films, err = core.films.GetFilms(start, end)
	} else {
		films, err = core.films.GetFilmsByGenre(genreId, start, end)
	}
	if err != nil {
		core.lg.Error("failed to get films from db", "err", err.Error())
		return films, "", fmt.Errorf("GetFilms err: %w", err)
	}

	genre, err := core.genres.GetGenreById(genreId)
	if err != nil {
		core.lg.Error("failed to get genre by id", "err", err.Error())
		genre = "Unknown Genre"
		return films, "", fmt.Errorf("GetFilms err: %w", err)
	}
	fmt.Println(films)
	return films, genre, nil
}

func (core *Core) GetFilmInfo(filmId uint64) (*requests.FilmResponse, error) {
	film, err := core.films.GetFilm(filmId)
	if err != nil {
		core.lg.Error("get film error", "err", err.Error())
		return nil, fmt.Errorf("get film err: %w", err)
	}
	if film.Title == "" {
		return nil, ErrNotFound
	}

	genres, err := core.genres.GetFilmGenres(filmId)
	if err != nil {
		core.lg.Error("get film genres error", "err", err.Error())
		return nil, fmt.Errorf("get film genres err: %w", err)
	}

	rating, number, err := core.films.GetFilmRating(filmId)
	if err != nil {
		core.lg.Error("get film rating error", "err", err.Error())
		return nil, fmt.Errorf("get film rating err: %w", err)
	}

	directors, err := core.crew.GetFilmDirectors(filmId)
	if err != nil {
		core.lg.Error("get film directors error", "err", err.Error())
		return nil, fmt.Errorf("get film directors err: %w", err)
	}

	scenarists, err := core.crew.GetFilmScenarists(filmId)
	if err != nil {
		core.lg.Error("get film scenarists error", "err", err.Error())
		return nil, fmt.Errorf("get film scenarists err: %w", err)
	}

	characters, err := core.crew.GetFilmCharacters(filmId)
	if err != nil {
		core.lg.Error("get film characters error", "err", err.Error())
		return nil, fmt.Errorf("get film scenarists err: %w", err)
	}

	result := requests.FilmResponse{
		Film:       *film,
		Genres:     genres,
		Rating:     rating,
		Number:     number,
		Directors:  directors,
		Scenarists: scenarists,
		Characters: characters,
	}

	return &result, nil
}

func (core *Core) GetActorInfo(actorId uint64) (*requests.ActorResponse, error) {
	actor, err := core.crew.GetActor(actorId)
	if err != nil {
		core.lg.Error("get actor error", "err", err.Error())
		return nil, fmt.Errorf("get actor err: %w", err)
	}
	if actor.Name == "" {
		return nil, ErrNotFound
	}

	career, err := core.profession.GetActorsProfessions(actorId)
	if err != nil {
		core.lg.Error("get actor profession error", "err", err.Error())
		return nil, fmt.Errorf("get actor profession err: %w", err)
	}

	result := requests.ActorResponse{
		Name:      actor.Name,
		Photo:     actor.Photo,
		BirthDate: actor.Birthdate,
		Country:   actor.Country,
		Info:      actor.Info,
		Career:    career,
	}
	return &result, nil
}

func (core *Core) GetActorsCareer(actorId uint64) ([]models.ProfessionItem, error) {
	career, err := core.profession.GetActorsProfessions(actorId)
	if err != nil {
		core.lg.Error("Get Actors Career error", "err", err.Error())
		return nil, fmt.Errorf("GetActorsCareer err: %w", err)
	}

	return career, nil
}

func (core *Core) GetGenre(genreId uint64) (string, error) {
	genre, err := core.genres.GetGenreById(genreId)
	if err != nil {
		core.lg.Error("GetGenre error", "err", err.Error())
		return "", fmt.Errorf("GetGenre err: %w", err)
	}

	return genre, nil
}

func (core *Core) FavoriteFilms(userId uint64) ([]models.FilmItem, error) {
	films, err := core.films.GetFavoriteFilms(userId)
	if err != nil {
		core.lg.Error("favorite films error", "err", err.Error())
		return nil, fmt.Errorf("favorite films err: %w", err)
	}

	return films, nil
}

func (core *Core) FavoriteFilmsAdd(userId uint64, filmId uint64) error {
	err := core.films.AddFavoriteFilm(userId, filmId)
	if err != nil {
		core.lg.Error("favorite film add error", "err", err.Error())
		return fmt.Errorf("favorite film add err: %w", err)
	}

	return nil
}

func (core *Core) FavoriteFilmsRemove(userId uint64, filmId uint64) error {
	err := core.films.RemoveFavoriteFilm(userId, filmId)
	if err != nil {
		core.lg.Error("favorite film remove error", "err", err.Error())
		return fmt.Errorf("favorite film remove err: %w", err)
	}

	return nil
}

func (core *Core) GetCalendar() (*requests.CalendarResponse, error) {
	var result *requests.CalendarResponse
	fmt.Println("go to repo Calendar")
	calendar, err := core.calendar.GetCalendar()
	if err != nil {
		core.lg.Error("get calendar error", "err", err.Error())
		return nil, fmt.Errorf("get calendar err: %w", err)
	}
	fmt.Println("return from repo Calendar")
	result.Days = calendar
	result.CurrentDay = uint8(time.Now().Day())
	result.MonthName = time.Now().Month().String()
	result.MonthText = "Новинки этого месяца"

	return result, nil
}
