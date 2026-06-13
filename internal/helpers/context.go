package helpers

import (
	"net/http"

	"github.com/farrasnazhif/go-starter/internal/middleware"
)

func UserIDFromContext(r *http.Request) string {
	return r.Context().Value(middleware.UserIDKey).(string)
}
