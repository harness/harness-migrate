package github

import "time"

// Error represents a Github error.
type (
	Error struct {
		Message string `json:"message"`
	}

	user struct {
		Login     string `json:"login"`
		ID        int    `json:"id"`
		AvatarURL string `json:"avatar_url"`
		Type      string `json:"type"`
		SiteAdmin bool   `json:"site_admin"`
	}

	codeComment struct {
		URL               string    `json:"url"`
		ID                int       `json:"id"`
		DiffHunk          string    `json:"diff_hunk"`
		Path              string    `json:"path"`
		CommitID          string    `json:"commit_id"`
		OriginalCommitID  string    `json:"original_commit_id"`
		User              user      `json:"user"`
		Body              string    `json:"body"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
		AuthorAssociation string    `json:"author_association"`
		StartLine         *int      `json:"start_line"`
		OriginalStartLine *int      `json:"original_start_line"`
		StartSide         *string   `json:"start_side"`
		Line              *int      `json:"line"`
		OriginalLine      *int      `json:"original_line"`
		Side              string    `json:"side"`
		InReplyToID       int       `json:"in_reply_to_id"`
		OriginalPosition  int       `json:"original_position"`
		Position          *int      `json:"position"`
		SubjectType       string    `json:"subject_type"`
	}

	HunkHeader struct {
		OldLine int
		OldSpan int
		NewLine int
		NewSpan int
		Text    string
	}

	ruleSet struct {
		ID          int         `json:"id"`
		Name        string      `json:"name"`
		Target      string      `json:"target"`
		SourceType  string      `json:"source_type"`
		Source      string      `json:"source"`
		Enforcement string      `json:"enforcement"`
		NodeID      string      `json:"node_id"`
		Links       interface{} `json:"_links"`
		CreatedAt   time.Time   `json:"created_at"`
		UpdatedAt   time.Time   `json:"updated_at"`
	}

	bypassActor struct {
		ActorID    int    `json:"actor_id"`
		ActorType  string `json:"actor_type"`
		BypassMode string `json:"bypass_mode"`
	}

	refName struct {
		Exclude []string `json:"exclude"`
		Include []string `json:"include"`
	}

	conditions struct {
		RefName refName `json:"ref_name"`
	}

	pullRequestParameters struct {
		RequiredApprovingReviewCount   int  `json:"required_approving_review_count"`
		DismissStaleReviewsOnPush      bool `json:"dismiss_stale_reviews_on_push"`
		RequireCodeOwnerReview         bool `json:"require_code_owner_review"`
		RequireLastPushApproval        bool `json:"require_last_push_approval"`
		RequiredReviewThreadResolution bool `json:"required_review_thread_resolution"`
	}

	rule struct {
		Type       string                 `json:"type"`
		Parameters map[string]interface{} `json:"parameters,omitempty"`
	}

	detailedRuleSet struct {
		ID                   int           `json:"id"`
		Name                 string        `json:"name"`
		Target               string        `json:"target"`
		SourceType           string        `json:"source_type"`
		Source               string        `json:"source"`
		Enforcement          string        `json:"enforcement"`
		Conditions           conditions    `json:"conditions"`
		Rules                []rule        `json:"rules"`
		NodeID               string        `json:"node_id"`
		CreatedAt            time.Time     `json:"created_at"`
		UpdatedAt            time.Time     `json:"updated_at"`
		BypassActors         []bypassActor `json:"bypass_actors"`
		CurrentUserCanBypass string        `json:"current_user_can_bypass"`
		Links                interface{}   `json:"_links"`
	}

	branchProtectionRulesResponse struct {
		Data struct {
			Repository struct {
				BranchProtectionRules struct {
					Edges []struct {
						Node branchProtectionRule `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"branchProtectionRules"`
			} `json:"repository"`
		} `json:"data"`
	}

	branchProtectionRule struct {
		AllowsDeletions                bool       `json:"allowsDeletions"`
		AllowsForcePushes              bool       `json:"allowsForcePushes"`
		BlocksCreations                bool       `json:"blocksCreations"`
		BypassForcePushAllowances      allowances `json:"bypassForcePushAllowances"`
		BypassPullRequestAllowances    actors     `json:"bypassPullRequestAllowances"`
		Creator                        actor      `json:"creator"`
		DatabaseID                     int        `json:"databaseId"`
		DismissesStaleReviews          bool       `json:"dismissesStaleReviews"`
		ID                             string     `json:"id"`
		IsAdminEnforced                bool       `json:"isAdminEnforced"`
		LockAllowsFetchAndMerge        bool       `json:"lockAllowsFetchAndMerge"`
		LockBranch                     bool       `json:"lockBranch"`
		Pattern                        string     `json:"pattern"`
		PushAllowances                 actors     `json:"pushAllowances"`
		RequireLastPushApproval        bool       `json:"requireLastPushApproval"`
		RequiredApprovingReviewCount   int        `json:"requiredApprovingReviewCount"`
		RequiredDeploymentEnvironments []string   `json:"requiredDeploymentEnvironments"`
		RequiresApprovingReviews       bool       `json:"requiresApprovingReviews"`
		RequiresCodeOwnerReviews       bool       `json:"requiresCodeOwnerReviews"`
		RequiresCommitSignatures       bool       `json:"requiresCommitSignatures"`
		RequiresConversationResolution bool       `json:"requiresConversationResolution"`
		RequiresDeployments            bool       `json:"requiresDeployments"`
		RequiresLinearHistory          bool       `json:"requiresLinearHistory"`
		RequiresStatusChecks           bool       `json:"requiresStatusChecks"`
		RequiresStrictStatusChecks     bool       `json:"requiresStrictStatusChecks"`
		RestrictsPushes                bool       `json:"restrictsPushes"`
		RestrictsReviewDismissals      bool       `json:"restrictsReviewDismissals"`
		ReviewDismissalAllowances      allowances `json:"reviewDismissalAllowances"`
	}

	actors struct {
		allowances
		Edges []struct {
			Node struct {
				Actor actor `json:"actor"`
			} `json:"node"`
		} `json:"edges"`
	}

	actor struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}

	allowances struct {
		TotalCount int `json:"totalCount"`
	}
)
