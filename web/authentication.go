package web

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	urepo "github.com/thrgamon/go-utils/repo/user"
	"github.com/thrgamon/nous/database"
	"github.com/thrgamon/nous/logger"
)

var Store *sessions.CookieStore

func Init() {
	Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
}

func getUserFromSession(r *http.Request) (urepo.User, bool) {
	sessionState, err := Store.Get(r, "auth")
	if err != nil {
		println(err.Error())
	}
	userRepo := urepo.NewUserRepo(database.Database)
	userId, ok := sessionState.Values["user_id"].(string)

	if ok {
		user, _ := userRepo.Get(r.Context(), urepo.Auth0ID(userId))
		return user, true
	} else {
		return urepo.User{}, false
	}
}

func EnsureAuthed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			_, ok := getUserFromSession(r)
			if ok {
				next.ServeHTTP(w, r)
			} else {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}
		} else {
			key, present := os.LookupEnv("AUTH_KEY")
			if present && key == authHeader {
				next.ServeHTTP(w, r)
			} else if present && key != authHeader {
				http.Error(w, "Could not authenticate request", http.StatusUnauthorized)
				return
			} else {
				http.Error(w, "Could not authenticate request", http.StatusInternalServerError)
				logger.Logger.Println("AUTH_KEY not found in environment")
				return
			}
		}
	})
}
