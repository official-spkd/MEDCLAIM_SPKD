// Package auth: login/logout, sesi berbasis cookie, dan RBAC.
package auth

import "github.com/gin-gonic/gin"

const (
	CookieName = "medclaim_sesi"
	contextKey = "principal"
)

// Principal = identitas pemanggil; JSON-nya persis tipe Sesi di frontend.
type Principal struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Nama         string  `json:"nama"`
	Role         string  `json:"role"`
	HospitalID   *string `json:"hospitalId"`
	HospitalNama *string `json:"hospitalNama"`
	HospitalKode *string `json:"hospitalKode"`
	Spesialisasi *string `json:"spesialisasi"`
	DpjpKode     *string `json:"dpjpKode"`
}

func Set(c *gin.Context, p *Principal) { c.Set(contextKey, p) }

// From mengambil principal; nil bila request belum terautentikasi.
func From(c *gin.Context) *Principal {
	if v, ok := c.Get(contextKey); ok {
		if p, ok := v.(*Principal); ok {
			return p
		}
	}
	return nil
}

// Must dipakai di handler yang berada di belakang middleware.Auth.
func Must(c *gin.Context) *Principal {
	p := From(c)
	if p == nil {
		panic("auth.Must dipanggil tanpa middleware.Auth")
	}
	return p
}
