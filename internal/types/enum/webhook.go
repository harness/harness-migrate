package enum

// Reference of the WebhookTrigger enum is harness/gitness/types/enum/webhook.go

// WebhookTrigger defines the different types of webhook triggers available.
type WebhookTrigger string

const (
	// WebhookTriggerBranchCreated gets triggered when a branch gets created.
	WebhookTriggerBranchCreated WebhookTrigger = "branch_created"
	// WebhookTriggerBranchUpdated gets triggered when a branch gets updated.
	WebhookTriggerBranchUpdated WebhookTrigger = "branch_updated"
	// WebhookTriggerBranchDeleted gets triggered when a branch gets deleted.
	WebhookTriggerBranchDeleted WebhookTrigger = "branch_deleted"

	// WebhookTriggerTagCreated gets triggered when a tag gets created.
	WebhookTriggerTagCreated WebhookTrigger = "tag_created"
	// WebhookTriggerTagUpdated gets triggered when a tag gets updated.
	WebhookTriggerTagUpdated WebhookTrigger = "tag_updated"
	// WebhookTriggerTagDeleted gets triggered when a tag gets deleted.
	WebhookTriggerTagDeleted WebhookTrigger = "tag_deleted"

	// WebhookTriggerPullReqCreated gets triggered when a pull request gets created.
	WebhookTriggerPullReqCreated WebhookTrigger = "pullreq_created"
	// WebhookTriggerPullReqReopened gets triggered when a pull request gets reopened.
	WebhookTriggerPullReqReopened WebhookTrigger = "pullreq_reopened"
	// WebhookTriggerPullReqBranchUpdated gets triggered when a pull request source branch gets updated.
	WebhookTriggerPullReqBranchUpdated WebhookTrigger = "pullreq_branch_updated"
	// WebhookTriggerPullReqClosed gets triggered when a pull request is closed.
	WebhookTriggerPullReqClosed WebhookTrigger = "pullreq_closed"
	// WebhookTriggerPullReqCommentCreated gets triggered when a pull request comment gets created.
	WebhookTriggerPullReqCommentCreated WebhookTrigger = "pullreq_comment_created"
	// WebhookTriggerPullReqMerged gets triggered when a pull request is merged.
	WebhookTriggerPullReqMerged WebhookTrigger = "pullreq_merged"
	// WebhookTriggerPullReqUpdated gets triggered when a pull request gets updated.
	WebhookTriggerPullReqUpdated WebhookTrigger = "pullreq_updated"
)

func ToStringSlice(vals []WebhookTrigger) []string {
	res := make([]string, len(vals))
	for i := range vals {
		res[i] = string(vals[i])
	}
	return res
}
