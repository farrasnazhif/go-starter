package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	ClientID  string
	Secret    string
	BaseURL   string
	PlanID    string
	WebhookID string
	ReturnURL string
	CancelURL string
}

type Client interface {
	CreateSubscription(ctx context.Context, userID string) (*SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, subscriptionID, reason string) error
	VerifyWebhook(ctx context.Context, headers http.Header, body json.RawMessage) error
}

type SubscriptionResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	ApproveURL string `json:"approve_url"`
}

type client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) Client {
	return &client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *client) getAccessToken(ctx context.Context) (string, error) {
	data := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v1/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.cfg.ClientID, c.cfg.Secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("paypal auth failed: %s", body)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

func (c *client) CreateSubscription(ctx context.Context, userID string) (*SubscriptionResponse, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"plan_id": c.cfg.PlanID,
		"custom_id": userID,
		"application_context": map[string]any{
			"return_url": c.cfg.ReturnURL,
			"cancel_url": c.cfg.CancelURL,
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v1/billing/subscriptions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("paypal create subscription failed: %s", errBody)
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	approveURL := ""
	for _, link := range result.Links {
		if link.Rel == "approve" {
			approveURL = link.Href
			break
		}
	}

	return &SubscriptionResponse{ID: result.ID, Status: result.Status, ApproveURL: approveURL}, nil
}

func (c *client) CancelSubscription(ctx context.Context, subscriptionID, reason string) error {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{"reason": reason})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v1/billing/subscriptions/"+subscriptionID+"/cancel", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("paypal cancel failed: %s", errBody)
	}
	return nil
}

func (c *client) VerifyWebhook(ctx context.Context, headers http.Header, body json.RawMessage) error {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"auth_algo":         headers.Get("Paypal-Auth-Algo"),
		"cert_url":          headers.Get("Paypal-Cert-Url"),
		"transmission_id":   headers.Get("Paypal-Transmission-Id"),
		"transmission_sig":  headers.Get("Paypal-Transmission-Sig"),
		"transmission_time": headers.Get("Paypal-Transmission-Time"),
		"webhook_id":        c.cfg.WebhookID,
		"webhook_event":     json.RawMessage(body),
	}
	raw, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v1/notifications/verify-webhook-signature", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.VerificationStatus != "SUCCESS" {
		return fmt.Errorf("webhook verification failed: %s", result.VerificationStatus)
	}
	return nil
}
