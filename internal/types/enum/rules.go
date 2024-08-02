package enum

type RuleState string

const (
	RuleStateActive   RuleState = "active"
	RuleStateMonitor  RuleState = "monitor"
	RuleStateDisabled RuleState = "disabled"
)
