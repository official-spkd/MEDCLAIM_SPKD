package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/response"
)

type CookieOpts struct {
	Secure   bool
	SameSite http.SameSite
}

type Handler struct {
	svc    *Service
	cookie CookieOpts
}

func NewHandler(svc *Service, cookie CookieOpts) *Handler {
	return &Handler{svc: svc, cookie: cookie}
}

func (h *Handler) setCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(CookieName, token, maxAge, "/", "", h.cookie.Secure, true)
}

// Login: POST /api/auth/login
func (h *Handler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := response.Bind(c, &body); err != nil {
		response.Fail(c, err)
		return
	}
	p, token, err := h.svc.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	h.setCookie(c, token, int(h.svc.TTL().Seconds()))
	c.JSON(http.StatusOK, p)
}

// Logout: POST /api/auth/logout (idempoten; aman dipanggil tanpa sesi).
func (h *Handler) Logout(c *gin.Context) {
	token, _ := c.Cookie(CookieName)
	if err := h.svc.Logout(c.Request.Context(), token, From(c)); err != nil {
		response.Fail(c, err)
		return
	}
	h.setCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Me: GET /api/auth/me
func (h *Handler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, Must(c))
}
