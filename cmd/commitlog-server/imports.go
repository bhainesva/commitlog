package main

// this file only exists to empty import packages so they'll show up
// in the list of packages provided by `go list all`
// TODO: use some other method to list packages. could look at goimports
// for inspiration, but it's pretty complicated and all in /internal packages
import (
	_ "github.com/go-jira/jira"
)

