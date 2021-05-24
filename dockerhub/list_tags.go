package dockerhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func getRepositoryTags(client *http.Client, token DockerHubToken, urlStr *string) (*DockerHubTagList, error) {
	req, buildErr := http.NewRequest("GET", *urlStr, nil)
	if buildErr != nil {
		return nil, buildErr
	}
	req.Header.Add("Authorization", "JWT"+*token)
	response, sendErr := client.Do(req)
	if sendErr != nil {
		return nil, sendErr
	}

	responseContent, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	if response.StatusCode == 200 {
		var content DockerHubTagList
		unmarshalErr := json.Unmarshal(responseContent, &content)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		} else {
			return &content, nil
		}
	} else {
		var content DockerErrorResponse
		unmarshalErr := json.Unmarshal(responseContent, &content)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		} else {
			return nil, errors.New(fmt.Sprintf("%d %s", response.StatusCode, content.Detail))
		}
	}
}

func iterateRepositoryTags(client *http.Client, token DockerHubToken, repo *DockerHubRepo, outputCh chan *TagInfo, bufferLength int) error {
	urlStr := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=%d", repo.User, repo.Name, bufferLength)
	for {
		tagList, err := getRepositoryTags(client, token, &urlStr)
		if err != nil {
			return err
		}

		for i, _ := range tagList.Results {
			info := NewTagInfo(repo, &tagList.Results[i])
			outputCh <- info
		}

		if tagList.Next != nil {
			urlStr = *tagList.Next
		} else {
			return nil
		}
	}
}

func TagsForRepoAsync(client *http.Client, token DockerHubToken, repo *DockerHubRepo, bufferLength int) (chan *TagInfo, chan error) {
	outCh := make(chan *TagInfo, bufferLength)
	errCh := make(chan error, 1)

	go func() {
		err := iterateRepositoryTags(client, token, repo, outCh, bufferLength)
		if err != nil {
			errCh <- err
			return
		} else {
			outCh <- nil
		}
	}()

	return outCh, errCh
}

func TagsForRepoStreaming(client *http.Client, token DockerHubToken, inputCh chan *DockerHubRepo, bufferLength int) (chan *TagInfo, chan error) {
	outCh := make(chan *TagInfo, bufferLength)
	errCh := make(chan error, 1)

	go func() {
		for {
			repo := <-inputCh
			if repo == nil {
				log.Print("All done, terminating")
				errCh <- nil
				return
			}

			err := iterateRepositoryTags(client, token, repo, outCh, bufferLength)
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

	return outCh, errCh
}
