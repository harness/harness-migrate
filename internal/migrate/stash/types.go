package stash

type (
	// Error represents a Stash error.
	Error struct {
		Message string `json:"message"`
		Status  int    `json:"status-code"`
		Errors  []struct {
			Message         string `json:"message"`
			ExceptionName   string `json:"exceptionName"`
			CurrentVersion  int    `json:"currentVersion"`
			ExpectedVersion int    `json:"expectedVersion"`
		} `json:"errors"`
	}

	pagination struct {
		Start    int  `json:"start"`
		Size     int  `json:"size"`
		Limit    int  `json:"limit"`
		LastPage bool `json:"isLastPage"`
		NextPage int  `json:"nextPageStart"`
	}

	author struct {
		Name         string `json:"name"`
		EmailAddress string `json:"emailAddress"`
		ID           int    `json:"id"`
		DisplayName  string `json:"displayName"`
		Active       bool   `json:"active"`
		Slug         string `json:"slug"`
		Type         string `json:"type"`
		Links        struct {
			Self []struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
	}

	commentAnchor struct {
		FromHash string `json:"fromHash"`
		ToHash   string `json:"toHash"`
		Line     int    `json:"line"`
		LineType string `json:"lineType"`
		FileType string `json:"fileType"`
		Path     string `json:"path"`
		DiffType string `json:"diffType"`
		Orphaned bool   `json:"orphaned"`
	}

	line struct {
		Destination int    `json:"destination"`
		Source      int    `json:"source"`
		Line        string `json:"line"`
		Truncated   bool   `json:"truncated"`
		CommentIDs  []int  `json:"commentIds"`
	}

	segment struct {
		Type      string `json:"type"`
		Lines     []line `json:"lines"`
		Truncated bool   `json:"truncated"`
	}

	hunk struct {
		SourceLine      int       `json:"sourceLine"`
		SourceSpan      int       `json:"sourceSpan"`
		DestinationLine int       `json:"destinationLine"`
		DestinationSpan int       `json:"destinationSpan"`
		Segments        []segment `json:"segments"`
		Truncated       bool      `json:"truncated"`
	}

	codeDiff struct {
		Source      interface{} `json:"source"`
		Destination interface{} `json:"destination"`
		Hunks       []hunk      `json:"hunks"`
		Truncated   bool        `json:"truncated"`
		Properties  interface{} `json:"properties"`
	}

	prCommentActivity struct {
		ID            int                `json:"id"`
		CreatedDate   int64              `json:"createdDate"`
		User          author             `json:"user"`
		Action        string             `json:"action"`
		CommentAction string             `json:"commentAction"`
		Comment       pullRequestComment `json:"comment"`
		CommentAnchor commentAnchor      `json:"commentAnchor"`
		Diff          codeDiff           `json:"diff"`
	}

	activities struct {
		pagination
		Values []interface{} `json:"values"`
	}

	pullRequestComment struct {
		Properties struct {
			RepositoryID int `json:"repositoryId"`
		} `json:"properties"`
		ID                  int                  `json:"id"`
		Version             int                  `json:"version"`
		Text                string               `json:"text"`
		Author              author               `json:"author"`
		CreatedDate         int64                `json:"createdDate"`
		UpdatedDate         int64                `json:"updatedDate"`
		Comments            []pullRequestComment `json:"comments"`
		Tasks               []interface{}        `json:"tasks"`
		PermittedOperations struct {
			Editable  bool `json:"editable"`
			Deletable bool `json:"deletable"`
		} `json:"permittedOperations"`
	}

	matcher struct {
		ID        string `json:"id"`
		DisplayID string `json:"displayId"`
		Type      struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"type"`
		Active bool `json:"active"`
	}

	scope struct {
		Type       string `json:"type"`
		ResourceID int    `json:"resourceId"`
	}
	
	branchPermission struct {
		ID         int      `json:"id"`
		Scope      scope    `json:"scope"`
		Type       string   `json:"type"`
		Matcher    matcher  `json:"matcher"`
		Users      []author `json:"users"`
		Groups     []string `json:"groups"`
		AccessKeys []struct {
			Key struct {
				ID    int    `json:"id"`
				Text  string `json:"text"`
				Label string `json:"label"`
			} `json:"key"`
		} `json:"accessKeys"`
	}

	branchPermissions struct {
		pagination
		Values []*branchPermission `json:"values"`
	}

	modelBranch struct {
		RefID      string `json:"refId"`
		UseDefault bool   `json:"useDefault"`
	}

	modelCategory struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Enabled     bool   `json:"enabled"`
		Prefix      string `json:"prefix"`
	}

	branchModels struct {
		Development modelBranch     `json:"development"`
		Production  modelBranch     `json:"production"`
		Categories  []modelCategory `json:"types"`
		Scope       scope           `json:"scope"`
	}

	modelValue struct {
		modelBranch
		Prefix string
	}
)
