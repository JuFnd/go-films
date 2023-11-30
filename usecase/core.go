package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"regexp"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/repository/profile"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/errors"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/comment"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/csrf"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/session"
)

type Core struct {
	sessions   session.SessionRepo
	csrfTokens csrf.CsrfRepo
	mutex      sync.RWMutex
	lg         *slog.Logger
	users      profile.IUserRepo
	comments   comment.ICommentRepo
}
type Session struct {
	Login     string
	ExpiresAt time.Time
}

func GetCore(cfg_sql configs.DbDsnCfg, cfg_csrf configs.DbRedisCfg, cfg_sessions configs.DbRedisCfg, lg *slog.Logger) (*Core, error) {
	csrf, err := csrf.GetCsrfRepo(cfg_csrf, lg)

	if err != nil {
		lg.Error("Csrf repository is not responding")
		return nil, err
	}
	session, err := session.GetSessionRepo(cfg_sessions, lg)

	if err != nil {
		lg.Error("Session repository is not responding")
		return nil, err
	}

	users, err := profile.GetUserRepo(&cfg_sql, lg)
	if err != nil {
		lg.Error("cant create repo")
		return nil, err
	}
	comments, err := comment.GetCommentRepo(cfg_sql, lg)
	if err != nil {
		lg.Error("cant create repo")
		return nil, err
	}
	core := Core{
		sessions:   *session,
		csrfTokens: *csrf,
		lg:         lg.With("module", "core"),
		users:      users,
		comments:   comments,
	}
	return &core, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (core *Core) CheckCsrfToken(ctx context.Context, token string) (bool, error) {
	core.mutex.RLock()
	found, err := core.csrfTokens.CheckActiveCsrf(ctx, token, core.lg)
	core.mutex.RUnlock()

	if err != nil {
		return false, err
	}

	return found, err
}

func (core *Core) CreateCsrfToken(ctx context.Context) (string, error) {
	sid := RandStringRunes(32)

	core.mutex.Lock()
	csrfAdded, err := core.csrfTokens.AddCsrf(
		ctx,
		models.Csrf{
			SID:       sid,
			ExpiresAt: time.Now().Add(3 * time.Hour),
		},
		core.lg,
	)
	core.mutex.Unlock()

	if !csrfAdded && err != nil {
		return "", err
	}

	if !csrfAdded {
		return "", nil
	}

	return sid, nil
}

func (core *Core) GetUserName(ctx context.Context, sid string) (string, error) {
	core.mutex.RLock()
	login, err := core.sessions.GetUserLogin(ctx, sid, core.lg)
	core.mutex.RUnlock()

	if err != nil {
		return "", err
	}

	return login, nil
}

func (core *Core) CreateSession(ctx context.Context, login string) (string, models.Session, error) {
	sid := RandStringRunes(32)

	newSession := models.Session{
		Login:     login,
		SID:       sid,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	core.mutex.Lock()
	sessionAdded, err := core.sessions.AddSession(ctx, newSession, core.lg)
	core.mutex.Unlock()

	if !sessionAdded && err != nil {
		return "", models.Session{}, err
	}

	if !sessionAdded {
		return "", models.Session{}, nil
	}

	return sid, newSession, nil
}

func (core *Core) FindActiveSession(ctx context.Context, sid string) (bool, error) {
	core.mutex.RLock()
	found, err := core.sessions.CheckActiveSession(ctx, sid, core.lg)
	core.mutex.RUnlock()

	if err != nil {
		return false, err
	}

	return found, nil
}

func (core *Core) KillSession(ctx context.Context, sid string) error {
	core.mutex.Lock()
	_, err := core.sessions.DeleteSession(ctx, sid, core.lg)
	core.mutex.Unlock()

	if err != nil {
		return err
	}

	return nil
}

func (core *Core) CreateUserAccount(login string, password string, name string, birthDate string, email string) error {
	if matched, _ := regexp.MatchString(`@`, email); !matched {
		return errors.InvalideEmail
	}
	err := core.users.CreateUser(login, password, name, birthDate, email)
	if err != nil {
		core.lg.Error("create user error", "err", err.Error())
		return fmt.Errorf("CreateUserAccount err: %w", err)
	}

	return nil
}

func (core *Core) FindUserAccount(login string, password string) (*profile.UserItem, bool, error) {
	user, found, err := core.users.GetUser(login, password)
	if err != nil {
		core.lg.Error("find user error", "err", err.Error())
		return nil, false, fmt.Errorf("FindUserAccount err: %w", err)
	}
	return user, found, nil
}

func (core *Core) FindUserByLogin(login string) (bool, error) {
	found, err := core.users.FindUser(login)
	if err != nil {
		core.lg.Error("find user error", "err", err.Error())
		return false, fmt.Errorf("FindUserByLogin err: %w", err)
	}

	return found, nil
}

func RandStringRunes(seed int) string {
	symbols := make([]rune, seed)
	for i := range symbols {
		symbols[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(symbols)
}

func (core *Core) GetFilmComments(filmId uint64, first uint64, limit uint64) ([]models.CommentItem, error) {
	comments, err := core.comments.GetFilmComments(filmId, first, limit)
	if err != nil {
		core.lg.Error("Get Film Comments error", "err", err.Error())
		return nil, fmt.Errorf("GetFilmComments err: %w", err)
	}

	return comments, nil
}

func (core *Core) AddComment(filmId uint64, userLogin string, rating uint16, text string) error {
	err := core.comments.AddComment(filmId, userLogin, rating, text)
	if err != nil {
		core.lg.Error("Add Comment error", "err", err.Error())
		return fmt.Errorf("GetActorsCareer err: %w", err)
	}

	return nil
}

func (core *Core) GetUserProfile(login string) (*profile.UserItem, error) {
	profile, err := core.users.GetUserProfile(login)
	if err != nil {
		core.lg.Error("GetUserProfile error", "err", err.Error())
		return nil, fmt.Errorf("GetUserProfile err: %w", err)
	}

	return profile, nil
}

func (core *Core) EditProfile(prevLogin string, login string, password string, email string, birthDate string, photo string) error {
	err := core.users.EditProfile(prevLogin, login, password, email, birthDate, photo)
	if err != nil {
		core.lg.Error("Edit profile error", "err", err.Error())
		return fmt.Errorf("Edit profile error: %w", err)
	}

	return nil
}

func (core *Core) HasUsersComment(login string, filmId uint64) (bool, error) {
	found, err := core.comments.HasUsersComment(login, filmId)
	if err != nil {
		core.lg.Error("find users comment error", "err", err.Error())
		return false, fmt.Errorf("find users comment error: %w", err)
	}

	return found, nil
}
