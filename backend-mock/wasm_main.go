//go:build js && wasm

// MedClaim — mode demo tanpa backend: seluruh API Go dijalankan di browser (WebAssembly).
// Data bersumber dari embed/medclaim.json (dummy) dan hanya hidup di memori tab.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall/js"
)

func buatMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", handleLogin)
	mux.HandleFunc("POST /api/auth/logout", handleLogout)
	mux.HandleFunc("GET /api/auth/me", handleMe)
	mux.HandleFunc("GET /api/bootstrap", handleBootstrap)
	mux.HandleFunc("GET /api/dashboard", handleDashboard)
	mux.HandleFunc("GET /api/anomali", handleAnomali)
	mux.HandleFunc("GET /api/financial", handleFinancial)
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/episodes", handleEpisodes)
	mux.HandleFunc("GET /api/episodes/{id}", handleEpisodeDetail)
	mux.HandleFunc("GET /api/uploads", handleUploads)
	mux.HandleFunc("POST /api/uploads/preview", handleUploadPreview)
	mux.HandleFunc("POST /api/uploads/commit", handleUploadCommit)
	mux.HandleFunc("POST /api/uploads/{id}/arsip", handleUploadArsip)
	mux.HandleFunc("DELETE /api/uploads/{id}", handleUploadDelete)
	mux.HandleFunc("GET /api/ledger", handleLedgerList)
	mux.HandleFunc("POST /api/ledger", handleLedgerBuat)
	mux.HandleFunc("GET /api/ledger/{id}", handleLedgerDetail)
	mux.HandleFunc("POST /api/ledger/{id}/transisi", handleLedgerTransisi)
	mux.HandleFunc("GET /api/queries", handleQueries)
	mux.HandleFunc("POST /api/queries", handleQueryBuat)
	mux.HandleFunc("POST /api/queries/{id}/answer", handleQueryJawab)
	mux.HandleFunc("GET /api/templates", handleTemplates)
	mux.HandleFunc("GET /api/admin/users", handleAdminUsers)
	mux.HandleFunc("POST /api/admin/users", handleAdminUserBuat)
	mux.HandleFunc("POST /api/admin/users/{id}/toggle", handleAdminUserToggle)
	mux.HandleFunc("GET /api/admin/audit", handleAdminAudit)
	mux.HandleFunc("GET /api/samples", handleSamples)
	mux.HandleFunc("GET /api/samples/{nama}", handleSampleIsi)
	return mux
}

func main() {
	store.file = "data/medclaim.json"
	if !store.muat() {
		seed(store)
	}
	mux := buatMux()

	// medclaimFetch(method, path, body, cookie) -> Promise<{status, body, setCookie}>
	js.Global().Set("medclaimFetch", js.FuncOf(func(this js.Value, args []js.Value) any {
		method, path, body, cookie := args[0].String(), args[1].String(), args[2].String(), args[3].String()
		handler := js.FuncOf(func(_ js.Value, pa []js.Value) any {
			resolve := pa[0]
			go func() {
				req := httptest.NewRequest(method, path, strings.NewReader(body))
				if cookie != "" {
					req.Header.Set("Cookie", cookie)
				}
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, req)
				out, _ := json.Marshal(map[string]any{
					"status":    rec.Code,
					"body":      rec.Body.String(),
					"setCookie": rec.Header().Get("Set-Cookie"),
				})
				resolve.Invoke(string(out))
			}()
			return nil
		})
		return js.Global().Get("Promise").New(handler)
	}))
	js.Global().Set("medclaimSiap", true)
	select {}
}
