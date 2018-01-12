package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	// EnvGitHubAPI is env var to set GitHub API base endpoint.
	// This is mainly used for GitHub Enterprise.
	EnvGitHubAPI = "GITHUB_API"
	// EnvGitHubToken is env var to set GitHub API token.
	EnvGitHubToken = "GITHUB_TOKEN"
	// EnvDebug is env var to handle debug mode.
	EnvDebug = "GHRD_DEBUG"
)

// Exit code are in value that represent an exit code for a particular error.
const (
	ExitCodeOK int = 0

	// Errors starts from 10
	ExitCodeParseFlagsError = 10 + iota
	ExitCodeInvalidArgs
	ExitCodeTagNotFound
	ExitCodeAssetIDNotFound
	ExitCodeOpenFileError
	ExitCodeAssetNotFound
)

const (
	defaultBaseURL = "https://api.github.com"
)

// CLI is the command line object.
type CLI struct {
	outStream, errStream io.Writer
}

// Debugf prints debug output.
func Debugf(format string, args ...interface{}) {
	if env := os.Getenv(EnvDebug); len(env) > 0 {
		log.Printf("[DEBUG]"+format+"\n", args...)
	}
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		owner    string
		repo     string
		token    string
		filepath string
		debug    bool
	)

	flags := flag.NewFlagSet("ghr-downloader", flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.StringVar(&owner, "username", "", "")
	flags.StringVar(&owner, "owner", "", "")
	flags.StringVar(&owner, "u", "", "")

	flags.StringVar(&repo, "repository", "", "")
	flags.StringVar(&repo, "r", "", "")

	flags.StringVar(&token, "token", os.Getenv(EnvGitHubToken), "")
	flags.StringVar(&token, "t", os.Getenv(EnvGitHubToken), "")

	flags.StringVar(&filepath, "path", "./", "")
	flags.StringVar(&filepath, "p", "./", "")

	flags.BoolVar(&debug, "debug", false, "")
	flags.BoolVar(&debug, "d", false, "")

	err := flags.Parse(args[1:])
	if err != nil {
		return ExitCodeParseFlagsError
	}

	if debug {
		os.Setenv(EnvDebug, "1")
		Debugf("Run as debug mode")
	}

	// Set base GitHub API. Base URL can be provided via env var, this is for GHE.
	baseURLStr := defaultBaseURL
	if urlStr := os.Getenv(EnvGitHubAPI); len(urlStr) > 0 {
		baseURLStr = urlStr
	}
	if baseURLStr[len(baseURLStr)-1] == '/' {
		baseURLStr = baseURLStr[:len(baseURLStr)-1]
	}

	gc := NewGitHubClient(owner, repo, token, baseURLStr)

	Debugf("Owner: %s", owner)
	Debugf("Repository: %s", repo)
	Debugf("GitHub API URL: %s", baseURLStr)

	var tag string
	parsedArgs := flags.Args()
	if len(parsedArgs) > 2 {
		Debugf("Error: invalid arguments")
		return ExitCodeInvalidArgs
	} else if len(parsedArgs) == 1 {
		tag = parsedArgs[0]
	}

	tag, err = gc.GetTag(tag)
	if err != nil {
		Debugf("Error: %s", err)
		return ExitCodeTagNotFound
	}
	Debugf("Tag: %s", tag)

	id, name, err := gc.GetLatestAssetID(tag)
	if err != nil {
		Debugf("Error: %s", err)
		return ExitCodeAssetIDNotFound
	}
	Debugf("Asset ID: %d", id)

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", filepath, name), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		Debugf("Error: %s", err)
		return ExitCodeOpenFileError
	}
	if err = gc.GetAsset(id, file); err != nil {
		Debugf("Error: %s", err)
		return ExitCodeAssetNotFound
	}

	return ExitCodeOK
}
