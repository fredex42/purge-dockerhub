package main

import (
	"flag"
	"github.com/fredex42/purge-dockerhub/dockerhub"
	"github.com/fredex42/purge-dockerhub/purge"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	log.Print("purge-dockerhub version 1.0, https://github.com/fredex42/purge-dockerhub")

	unamePtr := flag.String("uname", "", "username")
	passwdPtr := flag.String("passwd", "", "password")
	orgPtr := flag.String("org", "", "organisation")
	keepDuration := flag.Int("keepdays", 0, "Any item which has been pulled or pushed within this number of days will be kept. 0 disables the check.")
	alwaysKeep := flag.Int("keepcount", 4, "Always keep this many images regardless of age")
	reallyDelete := flag.Bool("really-delete", false, "Unless this flag is set, the app will only print what would be done and not actually delete anything")
	flag.Parse()

	httpClient := http.Client{}

	token, loginErr := dockerhub.Login(&httpClient, &dockerhub.DHLoginRequest{Username: *unamePtr, Password: *passwdPtr})
	if loginErr != nil {
		log.Fatal("ERROR Could not log in: ", loginErr)
	}

	repoOrg := *unamePtr
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
