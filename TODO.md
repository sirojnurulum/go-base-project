# Daftar TODO: Membangun Backend Skala Besar

## 1. Fondasi & Struktur Proyek
- [x] **Implementasi Layered Architecture:** Pisahkan kode menjadi `handler`, `service`, dan `repository`.
- [x] **Gunakan Dependency Injection (DI) Secara Konsisten:** Inisialisasi dependensi di `main.go` dan suntikkan ke layer yang membutuhkan (misal, `db` -> `repository`, `repository` -> `service`, `service` -> `handler`).
- [x] **Implementasi Graceful Shutdown:** Pastikan aplikasi bisa mematikan koneksi (database, Redis) dan menyelesaikan request yang sedang berjalan sebelum benar-benar berhenti.
- [x] **Konfigurasi Terpusat & Aman:** Muat semua konfigurasi dari environment variables (`.env` untuk development). Buat aplikasi gagal start (*fail-fast*) jika konfigurasi krusial seperti `JWT_SECRET` tidak ada.
- [x] **Error Handling Terpusat:** Gunakan `HTTPErrorHandler` kustom di Echo untuk menangani semua jenis error secara konsisten dan mengubahnya menjadi response JSON yang standar.

## 2. Database & Persistence
- [x] **Manajemen Migrasi Terprogram:** Gunakan **Goose** untuk mengelola skema database. Jalankan migrasi secara otomatis saat aplikasi dimulai.
- [x] **Operasi Inisialisasi yang Idempoten:** Gunakan `FirstOrCreate` atau transaksi database untuk fungsi seperti `InitAdminUser` agar aman dijalankan oleh beberapa instance aplikasi secara bersamaan.
- [x] **Konfigurasi Connection Pool:** Atur `SetMaxOpenConns`, `SetMaxIdleConns`, dan `SetConnMaxLifetime` pada koneksi GORM untuk mengoptimalkan performa database.
- [x] **Implementasi Caching dengan Redis:**
    - [x] Simpan *refresh token* di Redis untuk validasi dan rotasi.
    - [x] (Opsional) Cache data yang sering diakses untuk mengurangi beban ke database utama (misal: permissions).

## 3. API, Otentikasi & Keamanan
- [x] **Implementasi JWT (Access & Refresh Token):**
    - [x] Buat *access token* dengan masa berlaku singkat (misal, 15 menit).
    - [x] Buat *refresh token* dengan masa berlaku panjang (misal, 7 hari).
    - [x] Implementasikan **Token Rotation**: Saat *refresh token* digunakan, buat *refresh token* baru dan invalidasi yang lama.
- [x] **Gunakan DTO (Data Transfer Objects):** Buat struct terpisah untuk request body (`LoginRequest`) dan response body (`LoginResponse`) untuk memisahkan model internal dari API publik.
- [x] **Validasi Request yang Solid:** Gunakan library seperti `go-playground/validator` yang terintegrasi dengan Echo untuk memvalidasi DTO secara otomatis.
- [x] **Middleware Keamanan:**
    - [x] Buat middleware JWT untuk melindungi endpoint.
    - [x] Buat middleware berbasis peran (izin) untuk otorisasi.
- [x] **Konfigurasi CORS yang Ketat:** Hanya izinkan origin yang benar-benar dibutuhkan.
- [x] **Amankan Cookie:** Gunakan flag `HttpOnly`, `Secure` (di produksi), dan `SameSite=Strict` untuk cookie *refresh token*.

## 4. Observability (Logging, Metrics, Tracing)
- [ ] **Logging Terstruktur (Structured Logging):**
    - [x] Gunakan **Zerolog** atau `slog` (Go 1.21+).
    - [x] Tambahkan *context* ke setiap log (misalnya, `request_id`, `user_id`).
    - [x] Buat middleware untuk mencatat setiap request dan response.
- [ ] **Ekspos Metrik Aplikasi:**
    - [x] Gunakan library **Prometheus** untuk Echo.
    - [x] Ekspos endpoint `/metrics` untuk memantau latensi, jumlah request, dan error rate.
- [ ] **(Advanced) Distributed Tracing:** Implementasikan **OpenTelemetry** untuk melacak alur request jika aplikasi Anda nantinya akan berkembang menjadi microservices.

## 5. Testing
- [x] **Unit Test:** Tulis tes untuk *service layer*. Gunakan *mocking* (misalnya dengan `stretchr/testify/mock`) untuk meniru *repository layer* sehingga tes tidak memerlukan koneksi database.

## 6. Deployment & Operasional
- [ ] **Containerization dengan Docker:**
    - [ ] Buat `Dockerfile` *multi-stage* untuk menghasilkan image yang kecil dan aman.
- [ ] **Lingkungan Development yang Mudah:**
    - [ ] Buat file `docker-compose.yml` untuk menjalankan seluruh stack (aplikasi, PostgreSQL, Redis) dengan satu perintah.
- [ ] **Dokumentasi API:**
    - [x] Generate dokumentasi API menggunakan Swagger/OpenAPI.
 