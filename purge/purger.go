package purge

import (
	"errors"
	"fmt"
	"github.com/fredex42/purge-dockerhub/dockerhub"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func doDelete(client *http.Client, token dockerhub.DockerHubToken, result *DecisionInfo, reallyDelete bool) error {
	urlStr := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/%s/", result.TagSpec.Org, result.TagSpec.Repo, result.TagSpec.Tag)
	if reallyDelete {

		log.Printf("DEBUG AsyncPurger sending DELETE to %s", urlStr)
		req, buildErr := http.NewRequest("DELETE", urlStr, nil)
		if buildErr != nil {
			return buildErr
		}

		req.Header.Add("Authorization", "JWT "+*token)
		response, sendErr := client.Do(req)
		if sendErr != nil {
			return sendErr
		}

		responseBody, readErr := ioutil.ReadAll(response.Body)
		if readErr != nil {
			return readErr
		}

		if response.StatusCode >= 200 && response.StatusCode < 300 {
			log.Printf("DEBUG Delete was successful, server returned %s", string(responseBody))
			return nil
		} else if response.StatusCode == 429 || response.StatusCode == 502 { //rate limit exceeded, Bad Gateway
			log.Printf("INFO We are sending requests too fast, waiting before sending the next delete")
			maybeRetryAfter := response.Header.Get("Retry-After")
			if maybeRetryAfter != "" {
				log.Printf("INFO Server said to retry after %s", maybeRetryAfter)
				waitUntilEpoch, convErr := strconv.ParseInt(maybeRetryAfter, 10, 64)
				if convErr != nil {
					log.Printf("WARNING retry-after value was not a number, defaulting to 30s")
				} else {
					waitUntilTime := time.Unix(waitUntilEpoch, 0)
					log.Printf("DEBUG Waiting until %s", waitUntilTime.Format(time.RFC3339))
					time.Sleep(waitUntilTime.Sub(time.Now()))
					return doDelete(client, token, result, reallyDelete)
				}
			}
			time.Sleep(30 * time.Second)
			return doDelete(client, token, result, reallyDelete)
		} else {
			log.Printf("DEBUG Could not delete: server returned %d %s", response.StatusCode, responseBody)
			return errors.New("Could not delete item")
		}
	} else {
		log.Printf("DEBUG I would send DELETE to %s if reallyDelete was enabled", urlStr)
		return nil
	}
}

func AsyncPurger(inputCh chan *DecisionInfo, token dockerhub.DockerHubToken, reallyDelete bool) chan error {
	errCh := make(chan error, 1)

	go func() {
		httpClient := &http.Client{}
		for {
			result := <-inputCh
			if result == nil {
				log.Print("INFO AsyncPurger received null record, exiting")
				errCh <- nil
				return
			}

			if result.CanDelete() {
				err := doDelete(httpClient, token, result, reallyDelete)
				if err != nil {
					errCh <- err
					return
				}
			} else {
				log.Printf("INFO AsyncPurger keeping %s", result.GetCanonicalName())
			}
		}
	}()

	return errCh
}
