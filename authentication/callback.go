package authentication

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/thrgamon/learning_rank/repo"
)

var Store *sessions.CookieStore
var Db *pgxpool.Pool
var Log *log.Logger

type Profile struct {
	Nickname string
	Sub      string
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	sessionState, _ := Store.Get(r, "auth")
	authenticator, _ := New()

	if queryValues.Get("state") != sessionState.Values["state"] {
		http.Error(w, "Oh No", http.StatusBadRequest)
		return
	}

	token, err := authenticator.Exchange(r.Context(), queryValues["code"][0])

	if err != nil {
		http.Error(w, "Oh No", http.StatusUnauthorized)
		return
	}

	idToken, err := authenticator.VerifyIDToken(r.Context(), token)

	if err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
    Log.Println(err.Error())
    return
	}

	var profile Profile
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
    Log.Println(err.Error())
    return
	}

  userRepo := repo.NewUserRepo(Db)
	err, exists := userRepo.Exists(r.Context(), profile.Sub)
	if err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
    Log.Println(err.Error())
    return
	}

	if !exists {
		err :=userRepo.Add(r.Context(), profile.Nickname, profile.Sub)
		if err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
    Log.Println(err.Error())
    return
		}
	}

	sessionState.Values["access_token"] = token.AccessToken
	sessionState.Values["user_id"] = profile.Sub
	if err := sessionState.Save(r, w); err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
    Log.Println(err.Error())
    return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
