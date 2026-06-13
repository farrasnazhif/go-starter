package request

type CancelSubscription struct {
	Reason string `json:"reason"`
}
