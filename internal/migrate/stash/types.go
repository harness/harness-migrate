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

	prCommentActivity struct {
		ID            int                `json:"id"`
		CreatedDate   int64              `json:"createdDate"`
		User          author             `json:"user"`
		Action        string             `json:"action"`
		CommentAction string             `json:"commentAction"`
		Comment       pullRequestComment `json:"comment"`
		CommentAnchor commentAnchor      `json:"commentAnchor"`
		Diff          struct {
			Source      interface{}   `json:"source"`
			Destination interface{}   `json:"destination"`
			Hunks       []interface{} `json:"hunks"`
			Truncated   bool          `json:"truncated"`
			Properties  interface{}   `json:"properties"`
		}
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
)