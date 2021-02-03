package api

import (
	"net/http"

	"github.com/allez-chauffe/marcel/api/auth"
	"github.com/allez-chauffe/marcel/api/commons"
)

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if auth := auth.GetAuth(r); auth == nil {
		commons.WriteResponse(w, http.StatusForbidden, "")
	}
}

func validateAdminHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.CheckPermissions(r, nil, "admin") {
		commons.WriteResponse(w, http.StatusForbidden, "")
		return
	}
}
