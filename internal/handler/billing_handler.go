package handler

import (
	"io"
	"net/http"

	"github.com/farrasnazhif/go-starter/internal/dto/request"
	"github.com/farrasnazhif/go-starter/internal/helpers"
	"github.com/farrasnazhif/go-starter/internal/service"
)

type BillingHandler struct {
	billingService service.BillingService
}

func NewBillingHandler(billingService service.BillingService) *BillingHandler {
	return &BillingHandler{billingService: billingService}
}

func (h *BillingHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	userID := helpers.UserIDFromContext(r)

	result, err := h.billingService.CreateSubscription(r.Context(), userID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusInternalServerError, "failed to create subscription")
		return
	}

	helpers.SuccessResponse(w, http.StatusCreated, "Subscription created", result)
}

func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID := helpers.UserIDFromContext(r)

	var req request.CancelSubscription
	if err := helpers.ReadJSON(r, &req); err != nil {
		req.Reason = "User requested cancellation"
	}

	if err := h.billingService.CancelSubscription(r.Context(), userID, req.Reason); err != nil {
		helpers.ErrorResponse(w, http.StatusInternalServerError, "failed to cancel subscription")
		return
	}

	helpers.SuccessResponse(w, http.StatusOK, "Subscription cancelled", nil)
}

func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.billingService.HandleWebhook(r.Context(), r.Header, body); err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "webhook processing failed")
		return
	}

	helpers.SuccessResponse(w, http.StatusOK, "Webhook processed", nil)
}
