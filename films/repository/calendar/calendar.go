package calendar

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
)

type ICalendarRepo interface {
	GetCalendar() ([]models.DayItem, error)
}

type RepoPostgre struct {
	db *sql.DB
}

func GetCalendarRepo(config configs.DbDsnCfg, lg *slog.Logger) (*RepoPostgre, error) {
	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		config.User, config.DbName, config.Password, config.Host, config.Port, config.Sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		lg.Error("sql open error", "err", err.Error())
		return nil, fmt.Errorf("get calendar repo: %w", err)
	}
	err = db.Ping()
	if err != nil {
		lg.Error("sql ping error", "err", err.Error())
		return nil, fmt.Errorf("get calendar repo: %w", err)
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

func (repo *RepoPostgre) GetCalendar() ([]models.DayItem, error) {
	calendar := []models.DayItem{}
	day := uint8(1)
	news := ""

	rows, err := repo.db.Query("SELECT title, release_day FROM calendar " +
		"WHERE release_month = DATE_PART('MONTH', CURRENT_DATE) " +
		"ORDER BY release_day")
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := models.DayItem{}
		err := rows.Scan(&post.DayNews, &post.DayNumber)
		if err != nil {
			return nil, fmt.Errorf("get calendar scan err: %w", err)
		}

		if post.DayNumber != day {
			calendar = append(calendar, models.DayItem{DayNumber: day, DayNews: news})
			news = ""
		} else {
			news += post.DayNews + " "
		}

	}

	return calendar, nil
}
