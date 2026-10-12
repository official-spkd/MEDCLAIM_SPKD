// Package users memegang model pengguna yang dipakai modul auth dan admin.
package users

import "time"

// User sekaligus DTO publik (PasswordHash tidak pernah diserialisasi).
type User struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Nama         string     `json:"nama"`
	Role         string     `json:"role"`
	HospitalID   *string    `gorm:"type:uuid" json:"hospitalId"`
	Spesialisasi *string    `json:"spesialisasi"`
	DpjpKode     *string    `json:"dpjpKode"`
	Aktif        bool       `json:"aktif"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
}

func (User) TableName() string { return "users" }

var SemuaPeran = []string{"SUPERADMIN", "CASEMIX", "CODER", "DIREKTUR", "DPJP"}

func PeranValid(r string) bool {
	for _, p := range SemuaPeran {
		if p == r {
			return true
		}
	}
	return false
}
