package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/delivery"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/crew"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/film"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/genre"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/repository/profession"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/usecase"
)

func main() {
	logFile, _ := os.Create("film_log.log")
	lg := slog.New(slog.NewJSONHandler(logFile, nil))

	config, err := configs.ReadFilmConfig()
	if err != nil {
		lg.Error("read config error", "err", err.Error())
		return
	}

	var (
		films       film.IFilmsRepo
		genres      genre.IGenreRepo
		actors      crew.ICrewRepo
		professions profession.IProfessionRepo
	)
	fmt.Println(config.Films_db)

	switch config.Films_db {
	case "postgres":
		films, err = film.GetFilmRepo(*config, lg)
	}
	if err != nil {
		lg.Error("cant create repo")
		return
	}

	switch config.Genres_db {
	case "postgres":
		genres, err = genre.GetGenreRepo(*config, lg)
	}
	if err != nil {
		lg.Error("cant create repo")
		return
	}

	switch config.Crew_db {
	case "postgres":
		actors, err = crew.GetCrewRepo(*config, lg)
	}
	if err != nil {
		lg.Error("cant create repo")
		return
	}

	switch config.Profession_db {
	case "postgres":
		professions, err = profession.GetProfessionRepo(*config, lg)
	}
	if err != nil {
		lg.Error("cant create repo")
		return
	}
	fmt.Println(films)
	core := usecase.GetCore(config, lg, films, genres, actors, professions)

	fmt.Println(core)
	api := delivery.GetApi(core, lg)

	api.ListenAndServe()
}
