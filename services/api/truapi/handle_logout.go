package truapi

import (
	"net/http"
	"os"

	"github.com/TruStory/octopus/services/api/truapi/cookies"
)

// Logout deletes a session and redirects the logged in user to the correct page
func Logout() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		cookie := cookies.GetLogoutCookie()
		http.SetCookie(w, cookie)
		http.Redirect(w, req, os.Getenv("AUTH_LOGOUT_REDIR"), http.StatusFound)
	}
	return http.HandlerFunc(fn)
}