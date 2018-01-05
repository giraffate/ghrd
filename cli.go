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
)

// Exit code are in value that represent an exit code for a particular error.
const (
	ExitCodeOK int = 0

	// Errors starts from 10
	ExitCodeParseFlagsError = 10 + iota
	// TODO
	// Add exit code for errors
)

const (
	defaultBaseURL = "https://api.github.com"
)

// CLI is the command line object.
type CLI struct {
	outStream, errStream io.Writer
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var (
		owner    string
		repo     string
		token    string
		filepath string
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

	err := flags.Parse(args[1:])
	if err != nil {
		return ExitCodeParseFlagsError
	}

	// Set base GitHub API. Base URL can be provided via env var, this is for GHE.
	baseURLStr := defaultBaseURL
	if urlStr := os.Getenv(EnvGitHubAPI); len(urlStr) > 0 {
		baseURLStr = urlStr
	}
	// TODO
	// Remove trailing slash as default and return no error.
	if baseURLStr[len(baseURLStr)-1] == '/' {
		log.Fatalf("Remove trailing slash from base URL: %s\n", baseURLStr)
	}

	gc, err := NewGitHubClient(owner, repo, token, baseURLStr)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	var tag string
	parsedArgs := flags.Args()
	if len(parsedArgs) > 2 {
		log.Fatalln("Invalid argument: you can only set TAG.")
	} else if len(parsedArgs) == 1 {
		tag = parsedArgs[0]
	}

	tag, err = gc.GetTag(tag)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	id, name, err := gc.GetLatestAssetID(tag)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", filepath, name), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	if err = gc.GetAsset(id, file); err != nil {
		log.Fatalf("%v\n", err)
	}

	return ExitCodeOK
}
