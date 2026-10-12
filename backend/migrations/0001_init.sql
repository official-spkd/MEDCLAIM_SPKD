-- MedClaim SPKD — skema awal (diturunkan dari erd-medclaim.drawio)
-- Tambahan di luar ERD: episodes.hospital_id (isolasi tenant) dan tabel sessions.

-- ---------- Enum ----------
CREATE TYPE peran         AS ENUM ('SUPERADMIN', 'CASEMIX', 'CODER', 'DIREKTUR', 'DPJP');
CREATE TYPE format_upload AS ENUM ('BASIC', 'DETAIL');
CREATE TYPE status_upload AS ENUM ('PROSES', 'SELESAI', 'ARSIP', 'DIHAPUS');
CREATE TYPE tipe_layanan  AS ENUM ('RANAP', 'RJTP');
CREATE TYPE kelas_rugi    AS ENUM ('SURPLUS', 'S1_CODING', 'S2_MIXED', 'STRUCTURAL', 'RAHASIA');
CREATE TYPE status_ledger AS ENUM ('KANDIDAT', 'DIPERIKSA', 'DIKOREKSI', 'DIAJUKAN', 'HASIL', 'TIDAK_BERUBAH');
CREATE TYPE status_query  AS ENUM ('DRAFT', 'TERKIRIM', 'DIJAWAB', 'TERLAMBAT');

-- ---------- Master ----------
CREATE TABLE hospitals (
    id    varchar PRIMARY KEY,
    kode  varchar NOT NULL UNIQUE,
    nama  varchar NOT NULL,
    kota  varchar NOT NULL,
    kelas char(1) NOT NULL
);

CREATE TABLE users (
    id            varchar PRIMARY KEY,
    email         varchar NOT NULL UNIQUE,
    password_hash varchar NOT NULL,
    nama          varchar NOT NULL,
    role          peran   NOT NULL,
    hospital_id   varchar REFERENCES hospitals (id),
    spesialisasi  varchar,
    dpjp_kode     varchar,
    aktif         boolean     NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    last_login_at timestamptz,
    -- SUPERADMIN lintas RS (hospital_id NULL); peran lain wajib terikat ke satu RS
    CONSTRAINT users_hospital_sesuai_peran CHECK ((role = 'SUPERADMIN') = (hospital_id IS NULL))
);
CREATE INDEX users_hospital_idx ON users (hospital_id);

-- Sesi login (cookie). Tidak ada di ERD; dibutuhkan untuk /api/auth/*.
CREATE TABLE sessions (
    id         varchar PRIMARY KEY,
    user_id    varchar     NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);
CREATE INDEX sessions_user_idx    ON sessions (user_id);
CREATE INDEX sessions_expires_idx ON sessions (expires_at);

CREATE TABLE ref_inacbg (
    kode      varchar PRIMARY KEY,
    deskripsi varchar NOT NULL,
    tarif     bigint  NOT NULL DEFAULT 0
);

CREATE TABLE query_templates (
    id       varchar PRIMARY KEY,
    nama     varchar NOT NULL,
    isi      text    NOT NULL,
    kategori varchar NOT NULL,
    aktif    boolean NOT NULL DEFAULT true
);

CREATE TABLE settings (
    key   varchar PRIMARY KEY,
    value text NOT NULL
);

-- Referensi logis ke users.email (tanpa FK) agar log tetap utuh walau user dihapus.
CREATE TABLE audit_logs (
    id          varchar PRIMARY KEY,
    user_email  varchar     NOT NULL,
    aksi        varchar     NOT NULL,
    entitas     varchar     NOT NULL,
    entitas_id  varchar,
    meta        text,
    waktu       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_logs_waktu_idx ON audit_logs (waktu DESC);

-- ---------- Upload & episode ----------
CREATE TABLE upload_batches (
    id               varchar PRIMARY KEY,
    hospital_id      varchar       NOT NULL REFERENCES hospitals (id),
    uploaded_by_id   varchar       NOT NULL REFERENCES users (id),
    nama_file        varchar       NOT NULL,
    periode          char(7)       NOT NULL,           -- 'YYYY-MM'
    format           format_upload NOT NULL,
    dq_score         int           NOT NULL DEFAULT 0,
    dq_grade         char(1)       NOT NULL DEFAULT 'D',
    dq_report        jsonb,
    status           status_upload NOT NULL DEFAULT 'PROSES',
    total_baris      int           NOT NULL DEFAULT 0,
    baris_valid      int           NOT NULL DEFAULT 0,
    baris_gagal      int           NOT NULL DEFAULT 0,
    baris_baru       int           NOT NULL DEFAULT 0,
    baris_berubah    int           NOT NULL DEFAULT 0,
    baris_unchanged  int           NOT NULL DEFAULT 0,
    override_used    boolean       NOT NULL DEFAULT false,
    alasan_override  text,
    created_at       timestamptz   NOT NULL DEFAULT now()
);
CREATE INDEX upload_batches_hospital_idx ON upload_batches (hospital_id, created_at DESC);

CREATE TABLE episodes (
    id              varchar PRIMARY KEY,               -- 'noSep#vN'
    hospital_id     varchar      NOT NULL REFERENCES hospitals (id),   -- tambahan (denormalisasi dari upload_batches)
    upload_batch_id varchar      NOT NULL REFERENCES upload_batches (id),
    kode_icbg       varchar      NOT NULL REFERENCES ref_inacbg (kode),
    no_sep          varchar      NOT NULL,
    no_kartu        varchar      NOT NULL,
    no_mr           varchar      NOT NULL,
    nama_pasien     varchar      NOT NULL,
    tgl_lahir       date         NOT NULL,
    jenis_kelamin   char(1)      NOT NULL,
    tgl_masuk       date         NOT NULL,
    tgl_keluar      date         NOT NULL,
    los             int          NOT NULL,
    kelas_rawat     int          NOT NULL,
    tipe_layanan    tipe_layanan NOT NULL,
    cara_pulang     varchar,
    dpjp_kode       varchar,
    dpjp_nama       varchar,
    severity_level  int          NOT NULL DEFAULT 0,
    idrg_kode       varchar,
    total_billing   bigint       NOT NULL DEFAULT 0,
    tarif_inacbg    bigint       NOT NULL DEFAULT 0,
    selisih         bigint       NOT NULL DEFAULT 0,
    pct_rugi        numeric(8,2) NOT NULL DEFAULT 0,
    kelas_rugi      kelas_rugi   NOT NULL,
    dq_score        int          NOT NULL DEFAULT 0,
    row_hash        varchar      NOT NULL,
    version_no      int          NOT NULL DEFAULT 1,
    superseded      boolean      NOT NULL DEFAULT false,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    UNIQUE (hospital_id, no_sep, version_no)
);
-- Hanya satu versi aktif per SEP per RS
CREATE UNIQUE INDEX episodes_aktif_uq ON episodes (hospital_id, no_sep) WHERE NOT superseded;
CREATE INDEX episodes_hospital_tgl_idx ON episodes (hospital_id, tgl_keluar) WHERE NOT superseded;
CREATE INDEX episodes_upload_idx       ON episodes (upload_batch_id);
CREATE INDEX episodes_no_mr_idx        ON episodes (hospital_id, no_mr);
CREATE INDEX episodes_dpjp_idx         ON episodes (hospital_id, dpjp_kode);

CREATE TABLE episode_diagnoses (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    episode_id varchar NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    urutan     int     NOT NULL,
    kode_icd10 varchar NOT NULL,
    nama       varchar NOT NULL DEFAULT '',
    UNIQUE (episode_id, urutan)
);

CREATE TABLE episode_procedures (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    episode_id varchar NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    kode_icd9  varchar NOT NULL,
    nama       varchar NOT NULL DEFAULT ''
);
CREATE INDEX episode_procedures_episode_idx ON episode_procedures (episode_id);

CREATE TABLE episode_billing_components (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    episode_id varchar NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    komponen   varchar NOT NULL,
    nilai      bigint  NOT NULL DEFAULT 0,
    UNIQUE (episode_id, komponen)
);

CREATE TABLE episode_addons (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    episode_id varchar NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    kode       varchar NOT NULL,
    jenis      varchar NOT NULL,
    tarif      bigint  NOT NULL DEFAULT 0,
    status     varchar NOT NULL
);
CREATE INDEX episode_addons_episode_idx ON episode_addons (episode_id);

-- ---------- Ledger ----------
CREATE TABLE ledger_entries (
    id           varchar PRIMARY KEY,
    episode_id   varchar       NOT NULL REFERENCES episodes (id),
    hospital_id  varchar       NOT NULL REFERENCES hospitals (id),
    sep          varchar       NOT NULL,
    kelas_rugi   kelas_rugi    NOT NULL,
    estimasi     bigint        NOT NULL DEFAULT 0,
    realisasi    bigint,
    status       status_ledger NOT NULL DEFAULT 'KANDIDAT',
    strategy     varchar,
    referensi_rm varchar,
    catatan      text,
    created_at   timestamptz   NOT NULL DEFAULT now(),
    updated_at   timestamptz   NOT NULL DEFAULT now()
);
CREATE INDEX ledger_entries_hospital_idx ON ledger_entries (hospital_id, status);
CREATE INDEX ledger_entries_episode_idx  ON ledger_entries (episode_id);

CREATE TABLE ledger_events (
    id          varchar PRIMARY KEY,
    ledger_id   varchar       NOT NULL REFERENCES ledger_entries (id) ON DELETE CASCADE,
    dari_status status_ledger,
    ke_status   status_ledger,
    aktor_nama  varchar       NOT NULL,
    catatan     text,
    waktu       timestamptz   NOT NULL DEFAULT now()
);
CREATE INDEX ledger_events_ledger_idx ON ledger_events (ledger_id, waktu);

-- ---------- Query dokumen ----------
CREATE TABLE document_queries (
    id            varchar PRIMARY KEY,
    hospital_id   varchar      NOT NULL REFERENCES hospitals (id),
    sep           varchar      NOT NULL,
    nomor_mr      varchar      NOT NULL,
    inisial       varchar      NOT NULL,
    tgl_masuk     date,
    tgl_keluar    date,
    dpjp_id       varchar      NOT NULL REFERENCES users (id),
    dpjp_nama     varchar      NOT NULL,
    template_nama varchar,
    pertanyaan    text         NOT NULL,
    konteks       text         NOT NULL DEFAULT '',
    status        status_query NOT NULL DEFAULT 'DRAFT',
    deadline      timestamptz  NOT NULL,
    sent_at       timestamptz,
    answered_at   timestamptz,
    jawaban       text,
    sla_hari      int          NOT NULL DEFAULT 3,
    created_at    timestamptz  NOT NULL DEFAULT now()
);
CREATE INDEX document_queries_hospital_idx ON document_queries (hospital_id, status);
CREATE INDEX document_queries_dpjp_idx    ON document_queries (dpjp_id, status);
