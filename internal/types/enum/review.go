package enum

type ReviewDecision string

const (
	ReviewDecisionPending   ReviewDecision = "pending"
	ReviewDecisionReviewed  ReviewDecision = "reviewed"
	ReviewDecisionApproved  ReviewDecision = "approved"
	ReviewDecisionChangeReq ReviewDecision = "changereq"
)