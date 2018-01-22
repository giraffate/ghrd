[![Build Status](https://travis-ci.org/giraffate/ghrd.svg?branch=master)](https://travis-ci.org/giraffate/ghrd)

# ghrd
`ghrd` download an asset from GitHub Release.

If there are some assets with the specified tag, `ghrd` download an asset with the latest id.

## Usage
After setting GitHub API token, run the following command,
```
ghrd -u [OWNER] -r [REPOSITORY] [TAG]
```

### GitHub API token
Set it via environmental variable,
```
$ export GITHUB_TOKEN=xxx
```
or using `-t` option,
```
ghrd -u [OWNER] -r [REPOSITORY] -t [TOKEN] [TAG]
```

### GitHub Enterprise
Change API endpoint via environmental variable,
```
$ export GITHUB_API=xxx
```
