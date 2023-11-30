package delivery

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/usecase"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
)

type API struct {
	core usecase.ICore
	lg   *slog.Logger
	mx   *http.ServeMux
}

func GetApi(c *usecase.Core, l *slog.Logger) *API {
	api := &API{
		core: c,
		lg:   l.With("module", "api"),
	}
	mx := http.NewServeMux()
	mx.HandleFunc("/api/v1/films", api.Films)
	mx.HandleFunc("/api/v1/film", api.Film)
	mx.HandleFunc("/api/v1/actor", api.Actor)
	mx.HandleFunc("/api/v1/favorite/films", api.FavoriteFilms)
	mx.HandleFunc("/api/v1/favorite/film/add", api.FavoriteFilmsAdd)
	mx.HandleFunc("/api/v1/favorite/film/remove", api.FavoriteFilmsRemove)
	mx.HandleFunc("/api/v1/favorite/actors", api.FavoriteActors)
	mx.HandleFunc("/api/v1/favorite/actor/add", api.FavoriteActorsAdd)
	mx.HandleFunc("/api/v1/favorite/actor/remove", api.FavoriteActorsRemove)
	mx.HandleFunc("/api/v1/calendar", api.Calendar)

	api.mx = mx

	return api
}

func (a *API) ListenAndServe() {
	err := http.ListenAndServe(":8081", a.mx)
	if err != nil {
		a.lg.Error("ListenAndServe error", "err", err.Error())
	}
}

func (a *API) Films(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}

	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	title := r.URL.Query().Get("title")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")
	ratingFromStr := r.URL.Query().Get("rating_from")
	ratingToStr := r.URL.Query().Get("rating_to")
	mpaa := r.URL.Query().Get("mpaa")
	genres := strings.Split(r.URL.Query().Get("genre"), ",")
	actors := strings.Split(r.URL.Query().Get("actors"), ",")

	var ratingFrom, ratingTo float32
	var err error
	if ratingFromStr != "" {
		ratingFloat, err := strconv.ParseFloat(ratingFromStr, 32)
		if err != nil {
			response.Status = http.StatusBadRequest
			requests.SendResponse(w, response, a.lg)
			return
		}
		ratingFrom = float32(ratingFloat)
	}
	if ratingToStr != "" {
		ratingToFloat, err := strconv.ParseFloat(ratingToStr, 32)
		if err != nil {
			response.Status = http.StatusBadRequest
			requests.SendResponse(w, response, a.lg)
			return
		}
		ratingTo = float32(ratingToFloat)
	}

	films, err := a.core.FindFilm(title, dateFrom, dateTo, ratingFrom, ratingTo, mpaa, genres, actors)
	if err != nil {
		a.lg.Error("find film error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = films
	requests.SendResponse(w, response, a.lg)
}

func (a *API) Film(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	filmId, err := strconv.ParseUint(r.URL.Query().Get("film_id"), 10, 64)
	if err != nil {
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	film, err := a.core.GetFilmInfo(filmId)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			response.Status = http.StatusNotFound
			requests.SendResponse(w, response, a.lg)
			return
		}
		a.lg.Error("film error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = film

	requests.SendResponse(w, response, a.lg)
}

func (a *API) Actor(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	actorId, err := strconv.ParseUint(r.URL.Query().Get("actor_id"), 10, 64)
	if err != nil {
		a.lg.Error("actor error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	actor, err := a.core.GetActorInfo(actorId)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			response.Status = http.StatusNotFound
			requests.SendResponse(w, response, a.lg)
			return
		}
		a.lg.Error("actor error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = actor
	requests.SendResponse(w, response, a.lg)
}

func (a *API) FindFilm(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
	var request requests.FindFilmRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.lg.Error("find film error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		a.lg.Error("find film error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	films, err := a.core.FindFilm(request.Title, request.DateFrom, request.DateTo, request.RatingFrom, request.RatingTo, request.Mpaa, request.Genres, request.Actors)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			response.Status = http.StatusNotFound
			requests.SendResponse(w, response, a.lg)
			return
		}
		a.lg.Error("find film error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = films
	requests.SendResponse(w, response, a.lg)
}

func (a *API) FavoriteFilmsAdd(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) FavoriteFilmsRemove(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) FavoriteFilms(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) FavoriteActorsAdd(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) FavoriteActorsRemove(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) FavoriteActors(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) Calendar(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	calendar, err := a.core.GetCalendar()
	if err != nil {
		a.lg.Error("calendar error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = calendar
	requests.SendResponse(w, response, a.lg)
}
