package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// GitHubClient is the client for GitHub API.
type GitHubClient struct {
	owner, repo, token, urlStr, tag string
	*http.Client
}

// NewGitHubClient is the constructor of GitHubClient.
func NewGitHubClient(owner, repo, token, urlStr string) *GitHubClient {
	return &GitHubClient{
		owner:  owner,
		repo:   repo,
		token:  token,
		urlStr: urlStr,
		Client: &http.Client{},
	}
}

// RepositoryTag represents a repository tag.
type RepositoryTag struct {
	Name string `json:"name,omitempty"`
}

// ListTags lists tags for specified repository.
//
// GitHub API docs: https://developer.github.com/v3/repos/#list-tags
func (gc *GitHubClient) ListTags() (*[]RepositoryTag, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/tags", gc.urlStr, gc.owner, gc.repo)

	req, err := gc.NewRequest(u)
	if err != nil {
		return nil, err
	}

	resp, err := gc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	v := []RepositoryTag{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	return &v, nil
}

// GetTag check whether the tag exists or not, and get the tag if it exists .
// If the tag is not set, get the latest tag in dictionary order.
func (gc *GitHubClient) GetTag(tag string) (string, error) {
	if tag == "" {
		latestTag, err := gc.GetLatestTag()
		if err != nil {
			return tag, err
		}
		tag = latestTag
	} else {
		err := gc.IsFoundTag(tag)
		if err != nil {
			return tag, err
		}
	}
	return tag, nil
}

// GetLatestTag get the latest tag in dictionary order.
func (gc *GitHubClient) GetLatestTag() (string, error) {
	var retTag string
	repoTags, err := gc.ListTags()
	if err != nil {
		return retTag, err
	}

	for _, repoTag := range *repoTags {
		if repoTag.Name > retTag {
			retTag = repoTag.Name
		}
	}
	if retTag == "" {
		return retTag, errors.New("tag is not found")
	}
	return retTag, nil
}

// IsFoundTag check whether tag exists or not.
func (gc *GitHubClient) IsFoundTag(tag string) error {
	repoTags, err := gc.ListTags()
	if err != nil {
		return err
	}

	for _, repoTag := range *repoTags {
		if tag == repoTag.Name {
			return nil
		}
	}
	return errors.New("tag is not found")
}

// RepositoryRelease represents a GitHub release in a repository.
type RepositoryRelease struct {
	Assets []ReleaseAsset `json:"assets,omitempty"`
}

// ReleaseAsset represents a GitHub release asset in a repository.
type ReleaseAsset struct {
	ID   int    `json:"id,omitempty"`
	URL  string `json:"url,omitempty"`
	Name string `json:"name,omitempty"`
}

// GetLatestAssetID get the latest asset id concerned with the tag.
//
// GitHub API docs: https://developer.github.com/v3/repos/releases/#get-a-release-by-tag-name
func (gc *GitHubClient) GetLatestAssetID(tag string) (int, string, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", gc.urlStr, gc.owner, gc.repo, tag)
	req, err := gc.NewRequest(u)
	if err != nil {
		return 0, "", err
	}

	resp, err := gc.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	v := RepositoryRelease{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return 0, "", err
	}

	var retID int
	var retName string
	for _, asset := range v.Assets {
		if retID < asset.ID {
			retID = asset.ID
			retName = asset.Name
		}
	}
	return retID, retName, nil
}

// GetAsset get the asset concerned with the id.
//
// GitHub API docs: https://developer.github.com/v3/repos/releases/#get-a-single-release-asset
func (gc *GitHubClient) GetAsset(id int, file *os.File) error {
	u := fmt.Sprintf("%s/repos/%s/%s/releases/assets/%d", gc.urlStr, gc.owner, gc.repo, id)
	req, err := gc.NewRequest(u)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := gc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}
	return nil
}

// NewRequest returns a new http.Request.
func (gc *GitHubClient) NewRequest(urlStr string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", gc.token))
	return req, nil
}
