package truapi

import (
	"encoding/json"
	"net/http"

	"github.com/TruStory/octopus/services/truapi/truapi/cookies"
	"github.com/TruStory/octopus/services/truapi/truapi/render"
)

// AuthenticationRequest represents the http request to authenticate a user's account
type AuthenticationRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// TruErrors for user auth
var (
	ErrUnverifiedEmail = render.TruError{Code: 300, Message: "Please verify your email."}
)

// HandleUserAuthentication authenticates users using email/username and password combination
func (ta *TruAPI) HandleUserAuthentication(w http.ResponseWriter, r *http.Request) {
	// only support POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request AuthenticationRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	user, err := ta.DBClient.GetAuthenticatedUser(request.Identifier, request.Password)
	if err != nil || user == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if (*user).VerifiedAt.IsZero() {
		render.LoginError(w, r, ErrUnverifiedEmail, http.StatusBadRequest)
		return
	}

	cookie, err := cookies.GetLoginCookie(ta.APIContext, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	render.Response(w, r, user, http.StatusOK)
}
