package dockerhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getRepositoryPage(client *http.Client, token DockerHubToken, urlStr *string) (*DockerHubRepoList, error) {
	req, buildErr := http.NewRequest("GET", *urlStr, nil)
	if buildErr != nil {
		return nil, buildErr
	}
	req.Header.Add("Authorization", "JWT "+*token)
	response, setupErr := client.Get(*urlStr)
	if setupErr != nil {
		return nil, setupErr
	}
	responseBody, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	var content DockerHubRepoList
	marshalErr := json.Unmarshal(responseBody, &content)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return &content, nil
}

func RepositoriesAsync(client *http.Client, token DockerHubToken, org string, bufferLength int) (chan *DockerHubRepo, chan error) {
	outCh := make(chan *DockerHubRepo, bufferLength)
	errCh := make(chan error, 1)

	go func() {
		urlStr := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/?page_size=%d", org, bufferLength)
		for {
			repoList, err := getRepositoryPage(client, token, &urlStr)
			if err != nil {
				errCh <- err
				return
			}
			for _, repo := range repoList.Results {
				copiedData := repo
				outCh <- &copiedData
			}
			if repoList.Next == nil {
				outCh <- nil
				return
			} else {
				urlStr = *repoList.Next
			}
		}
	}()

	return outCh, errCh
}
