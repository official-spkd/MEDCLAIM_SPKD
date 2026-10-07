//go:build !js

// ============================================================
// MedClaim Go — Server utama (port 3030, stdlib murni)
// Menyajikan API /api/* + SPA Vue build di /app/
// ============================================================
package main

import (
        "fmt"
        "log"
        "net/http"
        "os"
        "path/filepath"
        "strings"
)

func main() {
        // Lokasi data & dist relatif terhadap file eksekusi
        exe, err := os.Executable()
        if err != nil {
                exe = "."
        }
        dir := filepath.Dir(exe)
        store.file = filepath.Join(dir, "data", "medclaim.json")

        if err := os.MkdirAll(filepath.Join(dir, "data"), 0755); err != nil {
                log.Fatal("Gagal membuat folder data:", err)
        }

        // Seed berjalan sekali saat startup (single-threaded, sebelum server listen).
        // JANGAN memegang write lock di sini — simpan() memakai RLock (bukan reentrant).
        if !store.muat() {
                log.Println("Database kosong — menjalankan seed demo…")
                seed(store)
                store.simpan()
                log.Printf("Seed selesai: %d episode, %d upload, %d ledger", len(store.Episodes), len(store.Uploads), len(store.Ledger))
        }

        mux := http.NewServeMux()

        // Auth
        mux.HandleFunc("POST /api/auth/login", handleLogin)
        mux.HandleFunc("POST /api/auth/logout", handleLogout)
        mux.HandleFunc("GET /api/auth/me", handleMe)

        // Bootstrap & analitik
        mux.HandleFunc("GET /api/bootstrap", handleBootstrap)
        mux.HandleFunc("GET /api/dashboard", handleDashboard)
        mux.HandleFunc("GET /api/anomali", handleAnomali)
        mux.HandleFunc("GET /api/financial", handleFinancial)
        mux.HandleFunc("GET /api/health", handleHealth)

        // Episodes
        mux.HandleFunc("GET /api/episodes", handleEpisodes)
        mux.HandleFunc("GET /api/episodes/{id}", handleEpisodeDetail)

        // Uploads
        mux.HandleFunc("GET /api/uploads", handleUploads)
        mux.HandleFunc("POST /api/uploads/preview", handleUploadPreview)
        mux.HandleFunc("POST /api/uploads/commit", handleUploadCommit)
        mux.HandleFunc("POST /api/uploads/{id}/arsip", handleUploadArsip)
        mux.HandleFunc("DELETE /api/uploads/{id}", handleUploadDelete)

        // Ledger
        mux.HandleFunc("GET /api/ledger", handleLedgerList)
        mux.HandleFunc("POST /api/ledger", handleLedgerBuat)
        mux.HandleFunc("GET /api/ledger/{id}", handleLedgerDetail)
        mux.HandleFunc("POST /api/ledger/{id}/transisi", handleLedgerTransisi)

        // Queries
        mux.HandleFunc("GET /api/queries", handleQueries)
        mux.HandleFunc("POST /api/queries", handleQueryBuat)
        mux.HandleFunc("POST /api/queries/{id}/answer", handleQueryJawab)
        mux.HandleFunc("GET /api/templates", handleTemplates)

        // Admin
        mux.HandleFunc("GET /api/admin/users", handleAdminUsers)
        mux.HandleFunc("POST /api/admin/users", handleAdminUserBuat)
        mux.HandleFunc("POST /api/admin/users/{id}/toggle", handleAdminUserToggle)
        mux.HandleFunc("GET /api/admin/audit", handleAdminAudit)

        // Sampel demo
        mux.HandleFunc("GET /api/samples", handleSamples)
        mux.HandleFunc("GET /api/samples/{nama}", handleSampleIsi)

        // SPA Vue (build dist)
        dist := filepath.Join(dir, "dist")
        fs := http.FileServer(http.Dir(dist))
        mux.Handle("GET /app", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                http.Redirect(w, r, "/app/", http.StatusMovedPermanently)
        }))
        mux.Handle("GET /app/", http.StripPrefix("/app/", spaHandler(fs)))

        port := os.Getenv("PORT")
        if port == "" {
                port = "3030"
        }
        log.Printf("MedClaim Go API berjalan di port %s (dist: %s)", port, dist)
        if err := http.ListenAndServe(":"+port, mux); err != nil {
                log.Fatal(err)
        }
}

// spaHandler — fallback ke index.html untuk aset SPA yang tidak ditemukan.
func spaHandler(fs http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                path := r.URL.Path
                if !strings.Contains(path, ".") && path != "/" {
                        // rute virtual SPA → kembalikan index.html
                        index := filepath.Join(".", "dist", "index.html")
                        if _, err := os.Stat(index); err == nil {
                                http.ServeFile(w, r, index)
                                return
                        }
                }
                fs.ServeHTTP(w, r)
        })
}

var _ = fmt.Sprintf
