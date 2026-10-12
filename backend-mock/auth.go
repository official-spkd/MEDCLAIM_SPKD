// ============================================================
// MedClaim Go — Autentikasi: PBKDF2-HMAC-SHA256 (stdlib murni)
// + cookie token HMAC (port dari session.ts)
// ============================================================
package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const cookieNama = "medclaim_sesi"
const maxAgeDetik = 60 * 60 * 24 * 7
const sesiSecret = "medclaim-dev-secret-spkd-2026"

// ---------- PBKDF2-HMAC-SHA256 ----------

func pbkdf2Kunci(sandi, salt []byte, iterasi, panjang int) []byte {
	prf := hmac.New(sha256.New, sandi)
	hashLen := prf.Size()
	numBlocks := (panjang + hashLen - 1) / hashLen
	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	U := make([]byte, hashLen)
	for block := 1; block <= numBlocks; block++ {
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(buf[:], uint32(block))
		prf.Write(buf[:])
		dk = prf.Sum(dk)
		T := dk[len(dk)-hashLen:]
		copy(U, T)
		for n := 2; n <= iterasi; n++ {
			prf.Reset()
			prf.Write(U)
			U = U[:0]
			U = prf.Sum(U)
			for x := range U {
				T[x] ^= U[x]
			}
		}
	}
	return dk[:panjang]
}

const iterasiPBKDF2 = 210000

func hashSandi(sandi string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	kunci := pbkdf2Kunci([]byte(sandi), salt, iterasiPBKDF2, 32)
	return "pbkdf2$" + strconv.Itoa(iterasiPBKDF2) + "$" + hex.EncodeToString(salt) + "$" + hex.EncodeToString(kunci)
}

func verifikasiSandi(sandi, tersimpan string) bool {
	parts := strings.Split(tersimpan, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2" {
		return false
	}
	iterasi, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	hash, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}
	coba := pbkdf2Kunci([]byte(sandi), salt, iterasi, len(hash))
	return subtle.ConstantTimeCompare(coba, hash) == 1
}

// ---------- Token sesi ----------

func tanda(payload string) string {
	m := hmac.New(sha256.New, []byte(sesiSecret))
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func buatToken(userID string) string {
	payload, _ := json.Marshal(map[string]interface{}{
		"uid": userID,
		"exp": time.Now().Add(time.Duration(maxAgeDetik) * time.Second).UnixMilli(),
	})
	p := base64.RawURLEncoding.EncodeToString(payload)
	return p + "." + tanda(p)
}

func bacaToken(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return ""
	}
	harapan := tanda(parts[0])
	if subtle.ConstantTimeCompare([]byte(harapan), []byte(parts[1])) != 1 {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ""
	}
	var data struct {
		UID string `json:"uid"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return ""
	}
	if data.Exp > 0 && data.Exp < time.Now().UnixMilli() {
		return ""
	}
	return data.UID
}

func setCookieSesi(w http.ResponseWriter, userID string) {
	http.SetCookie(w, &http.Cookie{
		Name: cookieNama, Value: buatToken(userID), Path: "/",
		MaxAge: maxAgeDetik, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

func hapusCookieSesi(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: cookieNama, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

// getSesi — baca cookie dari request, verifikasi, ambil pengguna.
func getSesi(r *http.Request) *Sesi {
	c, err := r.Cookie(cookieNama)
	if err != nil || c.Value == "" {
		return nil
	}
	uid := bacaToken(c.Value)
	if uid == "" {
		return nil
	}
	u := store.cariUser(uid)
	if u == nil || !u.Aktif {
		return nil
	}
	return buatSesi(u)
}

func butuhSesi(r *http.Request) (*Sesi, *RespError) {
	s := getSesi(r)
	if s == nil {
		return nil, &RespError{Kode: http.StatusUnauthorized, Pesan: "Tidak terautentikasi — silakan masuk kembali."}
	}
	return s, nil
}

// ---------- ID unik ----------

func idBaru(prefix string) string {
	b := make([]byte, 6)
	rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}
