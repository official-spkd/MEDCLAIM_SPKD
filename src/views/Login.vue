<script setup lang="ts">
// ============================================================
// MedClaim — Login v3 (split-screen, palette amber)
// ============================================================
import { ref } from "vue";
import { useRouter } from "vue-router";
import { masuk } from "../store";
import Ikon from "../components/Ikon.vue";

const router = useRouter();
const email = ref("admin@demo.medclaim.id");
const sandi = ref("MedClaim#demo2026");
const lihatSandi = ref(false);
const galat = ref("");
const memproses = ref(false);

const demoAkun = [
  { email: "admin@demo.medclaim.id", label: "Super Admin" },
  { email: "casemix@demo.medclaim.id", label: "Casemix" },
  { email: "coder@demo.medclaim.id", label: "Coder" },
  { email: "direktur@demo.medclaim.id", label: "Direktur" },
  { email: "dpjp@demo.medclaim.id", label: "DPJP" },
];

function pilih(e: string) {
  email.value = e;
  sandi.value = "MedClaim#demo2026";
}

async function submit() {
  galat.value = "";
  memproses.value = true;
  try {
    await masuk(email.value.trim(), sandi.value);
    router.push("/");
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal masuk — periksa koneksi ke layanan.";
  } finally {
    memproses.value = false;
  }
}
</script>

<template>
  <div class="login-tayang">
    <div class="login-kiri">
      <div>
        <div style="display: flex; align-items: center; gap: 13px; margin-bottom: 46px">
          <div
            style="
              width: 46px; height: 46px; border-radius: 13px;
              background: linear-gradient(140deg, #f0a640, #c07f1f);
              display: grid; place-items: center;
              box-shadow: 0 6px 20px rgba(240, 166, 64, 0.35);
            "
          >
            <svg width="24" height="24" viewBox="0 0 32 32" fill="none">
              <path d="M8 20c3-8 6-12 8-12s5 4 8 12" stroke="#2a1d0c" stroke-width="2.6" fill="none" stroke-linecap="round" />
              <circle cx="16" cy="20" r="3.2" fill="#2a1d0c" />
            </svg>
          </div>
          <div>
            <div style="font-family: var(--font-judul); font-weight: 700; font-size: 20px; color: #fff">MedClaim</div>
            <div style="font-size: 10px; letter-spacing: 2.2px; color: #b5a281">SPKD · SISTEM PELAYANAN KESEHATAN DAN DATA</div>
          </div>
        </div>
        <h1>From Hospital Data<br />to Action.</h1>
        <p class="moto" style="margin-top: 18px">
          Analitik pra-pengajuan klaim INA-CBG — deteksi kebocoran tarif sebelum klaim dikirim:
          coding loss S1, mixed loss S2, structural loss, anomali LOS, hingga recovery ledger empat mata.
        </p>

        <div style="display: flex; flex-direction: column; gap: 12px; margin-top: 30px; position: relative; z-index: 1">
          <div style="display: flex; gap: 11px; align-items: center; color: #e8d9ba; font-size: 13.5px">
            <span style="width: 30px; height: 30px; border-radius: 9px; background: rgba(240,166,64,.16); display: grid; place-items: center; color: #f0a640; flex-shrink: 0"><Ikon nama="perisai" :ukuran="15" /></span>
            20 aturan Data Quality 7 dimensi dengan gerbang ingest
          </div>
          <div style="display: flex; gap: 11px; align-items: center; color: #e8d9ba; font-size: 13.5px">
            <span style="width: 30px; height: 30px; border-radius: 9px; background: rgba(240,166,64,.16); display: grid; place-items: center; color: #f0a640; flex-shrink: 0"><Ikon nama="anomali" :ukuran="15" /></span>
            Deteksi anomali LOS, duplikat, readmisi &amp; selisih ekstrem
          </div>
          <div style="display: flex; gap: 11px; align-items: center; color: #e8d9ba; font-size: 13.5px">
            <span style="width: 30px; height: 30px; border-radius: 9px; background: rgba(240,166,64,.16); display: grid; place-items: center; color: #f0a640; flex-shrink: 0"><Ikon nama="buku" :ukuran="15" /></span>
            Recovery ledger dengan prinsip empat mata &amp; audit trail penuh
          </div>
        </div>
      </div>

      <div class="statistik">
        <div><b>20</b><span>Aturan DQ 7 dimensi</span></div>
        <div><b>5</b><span>Peran RBAC</span></div>
        <div><b>S1–S3</b><span>Strategi recovery</span></div>
        <div><b>100%</b><span>Audit trail</span></div>
      </div>
    </div>

    <div class="login-kanan">
      <form class="login-kotak" @submit.prevent="submit">
        <h2>Masuk ke MedClaim</h2>
        <p style="color: var(--text-2); margin: 0 0 24px; font-size: 13.5px">
          Gunakan akun demo untuk menjelajah sistem.
        </p>

        <div class="kolom-form" style="margin-bottom: 15px">
          <label for="email">Email</label>
          <div style="position: relative">
            <input id="email" v-model="email" type="email" required autocomplete="username" placeholder="nama@demo.medclaim.id" style="padding-left: 38px" />
            <span style="position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: var(--text-3); display: inline-flex">
              <Ikon nama="klaim" :ukuran="15" />
            </span>
          </div>
        </div>
        <div class="kolom-form" style="margin-bottom: 20px">
          <label for="sandi">Kata sandi</label>
          <div style="position: relative">
            <input
              id="sandi"
              v-model="sandi"
              :type="lihatSandi ? 'text' : 'password'"
              required
              autocomplete="current-password"
              placeholder="••••••••••"
              style="padding-left: 38px; padding-right: 42px"
            />
            <span style="position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: var(--text-3); display: inline-flex">
              <Ikon nama="perisai" :ukuran="15" />
            </span>
            <button
              type="button"
              @click="lihatSandi = !lihatSandi"
              style="position: absolute; right: 8px; top: 50%; transform: translateY(-50%); background: none; border: none; cursor: pointer; color: var(--text-3); padding: 6px; display: inline-flex"
              :aria-label="lihatSandi ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'"
            >
              <Ikon :nama="lihatSandi ? 'x' : 'matahari'" :ukuran="15" />
            </button>
          </div>
        </div>

        <div v-if="galat" class="bahaya-kecil" style="margin-bottom: 14px">{{ galat }}</div>

        <button class="btn utama" style="width: 100%; min-height: 44px; font-size: 14px" :disabled="memproses" type="submit">
          <span v-if="memproses" class="spinner" style="width: 16px; height: 16px; border-width: 2px; border-top-color: #2a1d0c"></span>
          {{ memproses ? "Memproses…" : "Masuk" }}
          <Ikon v-if="!memproses" nama="panahKanan" :ukuran="15" />
        </button>

        <div style="margin-top: 20px; font-size: 12px; color: var(--text-3); text-align: center">
          Akun demo (klik untuk mengisi) — sandi <b>MedClaim#demo2026</b>
        </div>
        <div class="demo-akun">
          <button v-for="a in demoAkun" :key="a.email" type="button" @click="pilih(a.email)">{{ a.label }}</button>
        </div>
      </form>
    </div>
  </div>
</template>
