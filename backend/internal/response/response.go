// Package response menyeragamkan bentuk error API: {"pesan": "...", "data": ...}.
// Frontend membaca field `pesan` dan status HTTP (401 memicu logout).
package response

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error adalah error yang aman ditampilkan ke klien.
type Error struct {
	Status int    `json:"-"`
	Pesan  string `json:"pesan"`
	Data   any    `json:"data,omitempty"`
}

func (e *Error) Error() string { return e.Pesan }

func New(status int, pesan string) *Error { return &Error{Status: status, Pesan: pesan} }

func WithData(status int, pesan string, data any) *Error {
	return &Error{Status: status, Pesan: pesan, Data: data}
}

func BadRequest(p string) *Error   { return New(http.StatusBadRequest, p) }
func Unauthorized(p string) *Error { return New(http.StatusUnauthorized, p) }
func Forbidden(p string) *Error    { return New(http.StatusForbidden, p) }
func NotFound(p string) *Error     { return New(http.StatusNotFound, p) }
func Conflict(p string) *Error     { return New(http.StatusConflict, p) }
func Unprocessable(p string) *Error {
	return New(http.StatusUnprocessableEntity, p)
}

// Fail menulis error ke klien. Error non-*Error dicatat dan dijawab 500 generik
// supaya detail internal (SQL, path) tidak bocor.
func Fail(c *gin.Context, err error) {
	var e *Error
	if errors.As(err, &e) {
		c.AbortWithStatusJSON(e.Status, e)
		return
	}
	log.Printf("ERROR %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, &Error{Pesan: "Terjadi kesalahan pada server."})
}

// Bind membaca body JSON; kegagalan menjadi 400.
func Bind(c *gin.Context, target any) error {
	if err := c.ShouldBindJSON(target); err != nil {
		return BadRequest("Body request tidak valid.")
	}
	return nil
}
