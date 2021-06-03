package dockerhub

import (
	"fmt"
	"time"
)

type DHLoginRequest struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type DockerLoginResponse struct {
	Token string `json:"token"`
}

type DockerErrorResponse struct {
	Detail string `json:"detail"`
}

type DockerHubRepo struct {
	User              string    `json:"user"`
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	Type              string    `json:"repository_type"`
	Status            int       `json:"status"`
	Description       string    `json:"string"`
	IsPrivate         bool      `json:"is_provate"`
	IsAutomated       bool      `json:"is_automated"`
	CanEdit           bool      `json:"can_edit"`
	StarCount         int       `json:"star_count"`
	PullCount         int       `json:"pull_count"`
	LastUpdated       time.Time `json:"last_updated"`
	IsMigrated        bool      `json:"is_migrated"`
	CollaboratorCount int       `json:"collaborator_count"`
	HubUser           string    `json:"hub_user"`
}

type DockerHubRepoList struct {
	Count    int             `json:"count"`
	Next     *string         `json:"next"`     //url for the next page
	Previous *string         `json:"previous"` //url for the previous page
	Results  []DockerHubRepo `json:"results"`
}

type DockerHubImage struct {
	Arch       *string   `json:"architecture"`
	Features   *string   `json:"features"`
	Variant    *string   `json:"variant"`
	Digest     *string   `json:"digest"`
	OS         *string   `json:"os"`
	OSFeatures *string   `json:"os_features"`
	OSVersion  *string   `json:"os_version"`
	Size       int64     `json:"size"`
	Status     string    `json:"status"`
	LastPulled time.Time `json:"last_pulled"`
	LastPushed time.Time `json:"last_pushed"`
}

type DockerHubTag struct {
	Creator             int              `json:"creator"`
	Id                  int              `json:"id"`
	ImageId             *string          `json:"image_id"`
	Images              []DockerHubImage `json:"images"`
	LastUpdated         time.Time        `json:"last_updated"`
	LastUpdater         int              `json:"last_updater"`
	LastUpdaterUsername string           `json:"last_updated_username"`
	TagName             string           `json:"name"`
	Repository          int              `json:"repository"`
	FullSize            int64            `json:"full_size"`
	V2                  bool             `json:"v2"`
	TagStatus           string           `json:"tag_status"`
	TagLastPulled       time.Time        `json:"tag_last_pulled"`
	TagLastPushed       time.Time        `json:"tag_last_pushed"`
}

type DockerHubTagList struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []DockerHubTag `json:"results"`
}

type TagInfo struct {
	Tag           *DockerHubTag
	CanonicalName string
}

func NewTagInfo(repo *DockerHubRepo, tag *DockerHubTag) *TagInfo {
	return &TagInfo{
		Tag:           tag,
		CanonicalName: repo.GetCanonicalNameForTag(tag),
	}
}

type TagSpec struct {
	Org  string
	Repo string
	Tag  string
}

func NewTagSpec(repo *DockerHubRepo, tag *DockerHubTag) *TagSpec {
	return &TagSpec{
		Org:  repo.User,
		Repo: repo.Name,
		Tag:  tag.TagName,
	}
}

func (t *TagSpec) GetCanonicalName() string {
	return fmt.Sprintf("%s/%s:%s", t.Org, t.Repo, t.Tag)
}

type DockerHubToken *string

func (repo *DockerHubRepo) GetCanonicalName() string {
	return fmt.Sprintf("%s/%s", repo.User, repo.Name)
}

func (repo *DockerHubRepo) GetCanonicalNameForTag(tag *DockerHubTag) string {
	return fmt.Sprintf("%s/%s:%s", repo.User, repo.Name, tag.TagName)
}
