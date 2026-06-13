package response

type Subscription struct {
	SubscriptionID string `json:"subscription_id"`
	ApproveURL     string `json:"approve_url"`
	Status         string `json:"status"`
}
