package authentication

import (
	"net/http"
	"net/url"
	"os"
)

// Handler for our logout.
func Logout(w http.ResponseWriter, r *http.Request) {
	logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + r.Host)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

  // Immediately expire the cookie
  sessionState, _ := Store.Get(r, "auth")
  sessionState.Options.MaxAge = -1
  sessionState.Save(r, w)

  http.Redirect(w, r, logoutUrl.String(), http.StatusTemporaryRedirect)
}
