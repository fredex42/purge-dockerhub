package dockerhub

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

/**
tries to log in to Docker Hub.
returns either the login token, or an error. Note that the token is stored as a pointer and must be considered
immutable.
*/
func Login(client *http.Client, rq *DHLoginRequest) (DockerHubToken, error) {
	requestBodyBytes, marshalErr := json.Marshal(rq)
	if marshalErr != nil {
		return nil, marshalErr
	}
	requestBodyReader := bytes.NewReader(requestBodyBytes)

	response, setupErr := client.Post("https://hub.docker.com/v2/users/login", "application/json", requestBodyReader)
	if setupErr != nil {
		return nil, setupErr
	}

	responseBody, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	log.Printf("DEBUG: login request returned %d %s", response.StatusCode, string(responseBody))

	if response.StatusCode == 200 {
		var response DockerLoginResponse
		unmarshalErr := json.Unmarshal(responseBody, &response)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		log.Printf("INFO Logged in to Docker Hub as %s", rq.Username)
		return &response.Token, nil
	} else {
		log.Printf("WARNING Login request returned status: %d", response.StatusCode)
		var response DockerErrorResponse
		unmarshalErr := json.Unmarshal(responseBody, &response)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		return nil, errors.New(response.Detail)
	}
}
