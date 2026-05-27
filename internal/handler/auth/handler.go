package auth

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	handlers "github.com/idalgo-2021/product-catalog-backend/internal/handler"
)

type AuthHandler struct {
	service AuthService
	logger  *zap.Logger
}

func NewAuthHandler(service AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body RegisterRequest true "Registration details"
// @Success      201  {object}  handler.Response{data=RegisterResponse}
// @Failure      400  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// TODO: DRY + нормальный валидатор
	if req.Email == "" || req.Password == "" {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Email and password are required")
		return
	}

	_, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to register user", err)
		return
	}

	handlers.Success(w, http.StatusCreated, RegisterResponse{Message: "User registered successfully"})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body LoginRequest true "Login credentials"
// @Success      200  {object}  handler.Response{data=TokenResponse}
// @Failure      400  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      401  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// TODO: DRY + нормальный валидатор
	if req.Email == "" || req.Password == "" {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Email and password are required")
		return
	}

	accessToken, refreshToken, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to login", err)
		return
	}

	handlers.Success(w, http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Refresh godoc
// @Summary      Refresh token
// @Description  Get a new access token using a refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body RefreshRequest true "Refresh token"
// @Success      200  {object}  handler.Response{data=TokenResponse}
// @Failure      400  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      401  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Refresh token is required")
		return
	}

	accessToken, refreshToken, err := h.service.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to refresh token", err)
		return
	}

	handlers.Success(w, http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Validate godoc
// @Summary      Validate token
// @Description  Check if the provided access token is valid
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body ValidateRequest true "Access token"
// @Success      200  {object}  handler.Response{data=ValidateResponse}
// @Failure      400  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Router       /auth/validate [post]
func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Token == "" {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Token is required")
		return
	}

	userID, err := h.service.ValidateToken(req.Token)
	if err != nil {
		handlers.Success(w, http.StatusOK, ValidateResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	handlers.Success(w, http.StatusOK, ValidateResponse{
		Valid:  true,
		UserID: userID,
	})
}
