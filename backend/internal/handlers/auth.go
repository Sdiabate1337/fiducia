package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fiducia/backend/internal/middleware"
	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
)

func (r *Router) handleRegister(w http.ResponseWriter, req *http.Request) {
	var body models.RegisterRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.Email == "" || body.Password == "" || body.CabinetName == "" {
		writeError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	resp, err := r.authSvc.Register(req.Context(), body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	var body models.LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.Email == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "Missing email or password")
		return
	}

	resp, err := r.authSvc.Login(req.Context(), body)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) handleMe(w http.ResponseWriter, req *http.Request) {
	// User ID comes from auth middleware
	userID, ok := middleware.GetUserID(req.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Fetch user details
	// Since we don't have GetByID exposed on AuthSvc yet, we might need to add it or use Repo on Router (which is untidy)
	// Ideally AuthSvc should have GetUser method.
	// For now, I'll access repo via existing struct fields if exposed, or strictly use Service.
	// User is checking "Me", so I probably want to return fresh user info.
	// Let's add GetUser to AuthSvc in next step or just bypass for now if AuthSvc struct is not modified.
	// Actually, I can allow direct repository usage here since Router has db/repos usually.
	// But wait, Router struct in router.go doesn't have userRepo.
	// I'll skip implementing 'Me' logic using Repo for this second and rely on what I have,
	// or assume I need to implement it.

	// Implementation: Return simple success for now with Claims data if possible,
	// or I'll quickly look up the user using the Service later.
	// For MVP, just returning { "id": userID } is proof of auth.

	// Fetch user details
	userRepo := repository.NewUserRepository(r.db)
	user, err := userRepo.GetByID(req.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}
