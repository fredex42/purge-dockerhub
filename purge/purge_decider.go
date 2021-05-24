package purge

import (
	"github.com/fredex42/purge-dockerhub/dockerhub"
	"log"
	"net/http"
	"sync"
	"time"
)

func makeDecisions(repo *dockerhub.DockerHubRepo, tag *dockerhub.TagInfo, returnedPlace int, notAfter *time.Time, alwaysKeepCount *int) *DecisionInfo {
	info := DecisionInfo{
		TagSpec:            dockerhub.NewTagSpec(repo, tag.Tag),
		LastPulledOverTime: false,
		LastPushedOverTime: false,
		OutsideKeepRecent:  false,
	}

	if notAfter == nil { //if the "notAfter" check is disabled then default to delete (i.e. both true) or nothing will get deleted at all
		info.LastPushedOverTime = true
		info.LastPushedOverTime = true
	}

	if notAfter != nil && tag.Tag.TagLastPulled.Before(*notAfter) {
		info.LastPulledOverTime = true
	}
	if notAfter != nil && tag.Tag.TagLastPushed.Before(*notAfter) {
		info.LastPushedOverTime = true
	}
	if alwaysKeepCount != nil && returnedPlace > *alwaysKeepCount {
		info.OutsideKeepRecent = true
	}
	return &info
}

func buildListForRepo(client *http.Client, token dockerhub.DockerHubToken, repo *dockerhub.DockerHubRepo, notAfter *time.Time, alwaysKeepCount *int, bufferLength int, outputCh chan *DecisionInfo) error {
	tagsCh, errCh := dockerhub.TagsForRepoAsync(client, token, repo, bufferLength)

	i := 0
	for {
		select {
		case tag := <-tagsCh:
			if tag == nil {
				log.Printf("Finished receiving %d tags for %s", i, repo.GetCanonicalName())
				return nil
			}
			info := makeDecisions(repo, tag, i, notAfter, alwaysKeepCount)
			outputCh <- info
			i += 1
		case err := <-errCh:
			return err
		}
	}
}

func PurgeDeciderAsync(inputCh chan *dockerhub.DockerHubRepo, outputCh chan *DecisionInfo, errCh chan error, token dockerhub.DockerHubToken, notAfter *time.Time, alwaysKeepCount *int, bufferLength int, wg *sync.WaitGroup) {
	httpClient := &http.Client{}

	go func() {
		for {
			repo := <-inputCh
			if repo == nil {
				log.Print("INFO PurgeDeciderAsync received nil record, exiting")
				wg.Done()
				return
			}

			err := buildListForRepo(httpClient, token, repo, notAfter, alwaysKeepCount, bufferLength, outputCh)
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

}

func ParallelPurgeDecider(inputCh chan *dockerhub.DockerHubRepo, token dockerhub.DockerHubToken, notAfter *time.Time, alwaysKeepCount *int, parallel int, bufferLength int) (chan *DecisionInfo, chan error) {
	outputCh := make(chan *DecisionInfo, bufferLength)
	duplicatedInputCh := make(chan *dockerhub.DockerHubRepo, bufferLength)
	errCh := make(chan error, 1)
	waitGroup := sync.WaitGroup{}

	for i := 0; i < parallel; i++ {
		PurgeDeciderAsync(duplicatedInputCh, outputCh, errCh, token, notAfter, alwaysKeepCount, bufferLength, &waitGroup)
	}
	waitGroup.Add(parallel)

	go func() {
		for {
			repo := <-inputCh
			if repo == nil {
				for i := 0; i < parallel; i++ {
					duplicatedInputCh <- nil
				}
				log.Print("INFO ParallelPurgeDecider waiting for threads to exit")
				waitGroup.Wait()
				log.Print("INFO ParallelPurgeDecider Done")
				outputCh <- nil
			} else {
				duplicatedInputCh <- repo
			}
		}
	}()

	return outputCh, errCh
}
