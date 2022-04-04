package authentication

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/thrgamon/learning_rank/repo"
)

var Store *sessions.CookieStore
var Db *pgxpool.Pool

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
		http.Error(w, "Oh No", http.StatusInternalServerError)
		return
	}

	var profile Profile
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, "Oh No", http.StatusInternalServerError)
		return
	}

  userRepo := repo.NewUserRepo(Db)
	err, exists := userRepo.Exists(r.Context(), profile.Sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		err :=userRepo.Add(r.Context(), profile.Nickname, profile.Sub)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	sessionState.Values["access_token"] = token.AccessToken
	sessionState.Values["user_id"] = profile.Sub
	if err := sessionState.Save(r, w); err != nil {
		http.Error(w, "Oh No", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
