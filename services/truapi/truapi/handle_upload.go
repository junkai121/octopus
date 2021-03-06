package truapi

import (
	"io/ioutil"
	"net/http"

	"github.com/TruStory/octopus/services/truapi/truapi/render"
)

// HandleUpload proxies the request from the clients to the uploader service
func (ta *TruAPI) HandleUpload(res http.ResponseWriter, req *http.Request) {

	// firing up the http client
	client := &http.Client{}

	// preparing the request
	request, err := http.NewRequest("POST", ta.APIContext.Config.App.UploadURL, req.Body)
	if err != nil {
		render.Error(res, req, err.Error(), http.StatusBadRequest)
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")

	// processing the request
	response, err := client.Do(request)
	if err != nil {
		render.Error(res, req, err.Error(), http.StatusBadRequest)
	}

	// reading the response
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		render.Error(res, req, err.Error(), http.StatusBadRequest)
	}

	// if all went well, sending back the response
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(responseBody)
	if err != nil {
		render.Error(res, req, err.Error(), http.StatusBadRequest)
	}
}
