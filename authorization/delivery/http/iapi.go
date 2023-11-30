package delivery

import (
	"net/http"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
)


type IApi interface {
	SendResponse(w http.ResponseWriter, response requests.Response)
    Signin(w http.ResponseWriter, r *http.Request)
    Signup(w http.ResponseWriter, r *http.Request)
    LogoutSession(w http.ResponseWriter, r *http.Request)
	AuthAccept(w http.ResponseWriter, r *http.Request)
}