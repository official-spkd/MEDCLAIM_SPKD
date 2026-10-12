package claims

// Katalog nama kode ICD yang umum. File V-Claim hanya memuat kode, bukan
// nama; kode di luar katalog diberi nama generik. Ganti dengan tabel master
// ICD-10/ICD-9-CM penuh bila tersedia.

var namaDx = map[string]string{
	"K29.7": "Gastritis", "K35.8": "Apendisitis akut", "J18.9": "Pneumonia",
	"E11.9": "DM tipe 2", "N18.5": "CKD stage 5", "C34.90": "Kanker paru",
	"I50.9": "Gagal jantung", "A41.9": "Sepsis", "A09": "Gastroenteritis",
	"S72.00": "Fraktur femur", "O82": "Persalinan SC", "D64.9": "Anemia",
	"I10": "Hipertensi", "J44.1": "COPD eksaserbasi", "K80.20": "Kolelitiase",
	"I50.0": "Gagal jantung kongestif", "N17.9": "Gagal ginjal akut",
	"J96.01": "Gagal napas", "E87.6": "Hipokalemia", "D69.6": "Trombositopenia",
	"A41.51": "Sepsis gram negatif", "I21.09": "Infark miokard",
}

var namaPros = map[string]string{
	"47.09": "Apendektomi", "48.29": "Kolostomi", "74.1": "Sectio caesarea inferiornya",
	"39.95": "Hemodialisa", "79.35": "Fiksasi internal fraktur", "96.04": "Intubasi",
	"88.72": "Ekokardiografi", "51.23": "Kolesistektomi",
}

func NamaDiagnosa(kode string) string {
	if n, ok := namaDx[kode]; ok {
		return n
	}
	return "Diagnosis " + kode
}

func NamaProsedur(kode string) string {
	if n, ok := namaPros[kode]; ok {
		return n
	}
	return "Prosedur " + kode
}
