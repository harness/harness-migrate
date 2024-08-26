package enum

// Reference of the RepoState enum is harness/gitness/types/enum/repo.go

// RepoState defines repo state.
type RepoState int

// RepoState enumeration.
const (
	RepoStateActive RepoState = iota
	RepoStateGitImport
	RepoStateMigrateGitPush
	RepoStateMigrateDataImport
)

// String returns the string representation of the RepoState.
func (state RepoState) String() string {
	switch state {
	case RepoStateActive:
		return "active"
	case RepoStateGitImport:
		return "git-import"
	case RepoStateMigrateGitPush:
		return "migrate-git-push"
	case RepoStateMigrateDataImport:
		return "migrate-data-import"
	default:
		return "undefined"
	}
}
