package handler

import (
	"net/http"

	"github.com/farrasnazhif/go-starter/internal/helpers"
	"github.com/farrasnazhif/go-starter/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := helpers.UserIDFromContext(r)

	result, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.SuccessResponse(w, http.StatusOK, "User retrieved successfully", result)
}

func (h *UserHandler) GetBilling(w http.ResponseWriter, r *http.Request) {
	userID := helpers.UserIDFromContext(r)

	result, err := h.userService.GetBilling(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.SuccessResponse(w, http.StatusOK, "User billing retrieved successfully", result)
}
