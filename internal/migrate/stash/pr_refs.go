package stash

import (
	"github.com/go-git/go-git/v5/config"
)

func (e *Export) PullRequestRefs() []config.RefSpec {
	return []config.RefSpec{"refs/pull-requests/*/from:refs/pullreq/*/head"}
}
