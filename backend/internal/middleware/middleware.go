// Package middleware: autentikasi sesi, RBAC, CORS.
package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/response"
)

// Auth mewajibkan sesi valid; Principal disimpan di context.
func Auth(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, err := resolve(c, svc)
		if err != nil {
			response.Fail(c, err)
			return
		}
		if p == nil {
			response.Fail(c, response.Unauthorized("Tidak terautentikasi — silakan masuk kembali."))
			return
		}
		auth.Set(c, p)
		c.Next()
	}
}

// OptionalAuth mengisi Principal bila ada sesi valid, tanpa menolak request.
func OptionalAuth(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if p, err := resolve(c, svc); err == nil && p != nil {
			auth.Set(c, p)
		}
		c.Next()
	}
}

func resolve(c *gin.Context, svc *auth.Service) (*auth.Principal, error) {
	token, err := c.Cookie(auth.CookieName)
	if err != nil || token == "" {
		return nil, nil
	}
	return svc.Resolve(c.Request.Context(), token)
}

// RequireAccess menolak (403) peran yang tidak punya akses baca ke grup.
func RequireAccess(grup, pesan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !auth.BolehAkses(grup, auth.Must(c).Role) {
			response.Fail(c, response.Forbidden(pesan))
			return
		}
		c.Next()
	}
}

// RequireMutate menolak (403) peran yang tidak boleh menulis di grup.
func RequireMutate(grup, pesan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !auth.BolehMutasi(grup, auth.Must(c).Role) {
			response.Fail(c, response.Forbidden(pesan))
			return
		}
		c.Next()
	}
}

// RequireRole menolak (403) peran di luar daftar.
func RequireRole(pesan string, peran ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slices.Contains(peran, auth.Must(c).Role) {
			response.Fail(c, response.Forbidden(pesan))
			return
		}
		c.Next()
	}
}

// CORS mengizinkan origin terdaftar dengan credentials (cookie sesi).
func CORS(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && slices.Contains(origins, origin) {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Add("Vary", "Origin")
			if c.Request.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				reqHeaders := c.GetHeader("Access-Control-Request-Headers")
				if reqHeaders == "" {
					reqHeaders = "Content-Type"
				}
				h.Set("Access-Control-Allow-Headers", reqHeaders)
				h.Set("Access-Control-Max-Age", "600")
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}
		c.Next()
	}
}

// SecurityHeaders: header dasar untuk API JSON.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "no-store")
		c.Next()
	}
}
