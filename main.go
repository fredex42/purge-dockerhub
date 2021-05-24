package main

import (
	"flag"
	"github.com/fredex42/purge-dockerhub/dockerhub"
	"github.com/fredex42/purge-dockerhub/purge"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func loadCredentials(filepathPtr *string) *dockerhub.DHLoginRequest {
	f, openErr := os.Open(*filepathPtr)
	if openErr != nil {
		log.Fatalf("Could not load credentials from %s: %s", *filepathPtr, openErr)
	}
	defer f.Close()

	content, readErr := ioutil.ReadAll(f)
	if readErr != nil {
		log.Fatalf("Could not read credentials from %s: %s", *filepathPtr, readErr)
	}

	var result dockerhub.DHLoginRequest
	marshalErr := yaml.Unmarshal(content, &result)
	if marshalErr != nil {
		log.Fatalf("Credentials information in %s was malformatted: %s", *filepathPtr, marshalErr)
	}
	return &result
}

func main() {
	log.Print("purge-dockerhub version 1.0, https://github.com/fredex42/purge-dockerhub")

	credentialsFile := flag.String("credentails", "docker-credentials.yaml", "yaml-format file containing the credentials to log in to Docker Hub")
	orgPtr := flag.String("org", "", "organisation")
	keepDuration := flag.Int("keepdays", 0, "Any item which has been pulled or pushed within this number of days will be kept. 0 disables the check.")
	alwaysKeep := flag.Int("keepcount", 150, "Always keep this many images regardless of age")
	reallyDelete := flag.Bool("really-delete", false, "Unless this flag is set, the app will only print what would be done and not actually delete anything")
	flag.Parse()

	httpClient := http.Client{}

	credentials := loadCredentials(credentialsFile)
	token, loginErr := dockerhub.Login(&httpClient, credentials)
	if loginErr != nil {
		log.Fatal("ERROR Could not log in: ", loginErr)
	}

	repoOrg := credentials.Username
	if *orgPtr != "" {
		repoOrg = *orgPtr
	}

	var notAfter *time.Time
	if *keepDuration != 0 {
		result := time.Now().AddDate(0, 0, -*keepDuration)
		notAfter = &result
		log.Printf("INFO Images after %s will be kept", notAfter.Format(time.RFC3339))
	}

	var maybeKeepCount *int
	if *alwaysKeep != 0 {
		maybeKeepCount = alwaysKeep
		log.Printf("INFO Will keep %d images in each repo regardless of age", *alwaysKeep)
	}

	repoCh, listErrCh := dockerhub.RepositoriesAsync(&httpClient, token, repoOrg, 100)
	//tagsCh, tagsErrCh := dockerhub.TagsForRepoAsync(&httpClient, token, repoCh, 100)
	resultCh, resultErrCh := purge.ParallelPurgeDecider(repoCh, token, notAfter, maybeKeepCount, 2, 100)
	purgeErrCh := purge.AsyncPurger(resultCh, token, *reallyDelete)

	//select {
	//	case err := <-listErrCh:
	//		log.Printf("ERROR could not list repos: %s", err)
	//	case err := <-tagsErrCh:
	//		log.Printf("ERROR could not list tags: %s", err)
	//}
	for {
		select {
		case err := <-listErrCh:
			if err != nil {
				log.Printf("ERROR Could not list repos: %s", err)
				os.Exit(1)
			}
		case err := <-resultErrCh:
			if err != nil {
				log.Printf("ERROR Could not process results: %s", err)
				os.Exit(1)
			}
			os.Exit(0)
		case err := <-purgeErrCh:
			if err != nil {
				log.Printf("ERROR Could not perform deletion: %s", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}
