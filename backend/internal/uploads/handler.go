package uploads

import (
	"embed"
	"errors"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/response"
)

// Batas ukuran body upload (isi file dikirim sebagai string JSON).
const batasBody = 32 << 20

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func bind(c *gin.Context) (Permintaan, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, batasBody)
	var req Permintaan
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			response.Fail(c, response.New(http.StatusRequestEntityTooLarge, "Ukuran file melebihi batas 32 MB."))
		} else {
			response.Fail(c, response.BadRequest("Body request tidak valid."))
		}
		return req, false
	}
	return req, true
}

// Preview: POST /api/uploads/preview
func (h *Handler) Preview(c *gin.Context) {
	req, ok := bind(c)
	if !ok {
		return
	}
	out, err := h.svc.Preview(c.Request.Context(), auth.Must(c), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Commit: POST /api/uploads/commit
func (h *Handler) Commit(c *gin.Context) {
	req, ok := bind(c)
	if !ok {
		return
	}
	out, err := h.svc.Commit(c.Request.Context(), auth.Must(c), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// List: GET /api/uploads
func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), auth.Must(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Arsip: POST /api/uploads/:id/arsip
func (h *Handler) Arsip(c *gin.Context) {
	out, err := h.svc.Arsip(c.Request.Context(), auth.Must(c), c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Hapus: DELETE /api/uploads/:id
func (h *Handler) Hapus(c *gin.Context) {
	if err := h.svc.Hapus(c.Request.Context(), auth.Must(c), c.Param("id")); err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------- File contoh ----------

//go:embed samples/*.txt
var samplesFS embed.FS

var deskripsiSample = map[string]string{
	"vclaim-basic.txt":     "Format Basic (14 kolom) — periode 2025-10, 150 baris",
	"vclaim-detail.txt":    "Format Detail (41 kolom) — periode 2025-11, 150 baris",
	"vclaim-detail-v2.txt": "Revisi Detail (6 baris berubah: LOS & severity naik) — uji versioning",
	"vclaim-broken.txt":    "File rusak: duplikat SEP, tanggal 31-02, tanggal terbalik, tarif 0 — uji gerbang DQ",
}

// Samples: GET /api/samples
func (h *Handler) Samples(c *gin.Context) {
	entries, _ := samplesFS.ReadDir("samples")
	items := []map[string]any{}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || e.IsDir() {
			continue
		}
		items = append(items, map[string]any{
			"nama": e.Name(), "ukuran": info.Size(), "deskripsi": deskripsiSample[e.Name()],
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["nama"].(string) < items[j]["nama"].(string) })
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// SampleIsi: GET /api/samples/:nama
func (h *Handler) SampleIsi(c *gin.Context) {
	nama := path.Base(c.Param("nama"))
	if !strings.HasSuffix(nama, ".txt") {
		response.Fail(c, response.BadRequest("Hanya file .txt yang diizinkan."))
		return
	}
	isi, err := samplesFS.ReadFile("samples/" + nama)
	if err != nil {
		response.Fail(c, response.NotFound("Sampel tidak ditemukan."))
		return
	}
	c.JSON(http.StatusOK, gin.H{"nama": nama, "isi": string(isi)})
}
