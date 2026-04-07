# API Capstone C1 - Kelompok 5

API untuk Capstone Project kasus B1. Dikelola oleh Kelompok 5

## Note

Pastikan `Docker` dan `Docker Compose` sudah terpasang pada perangkat anda
sebelum menjalankannya.
Jika belum, ikuti panduan pada link berikut: <https://docs.docker.com/compose/install/>

Untuk menjalankan proyek lakukan langkah berikut:

```sh
git clone https://www.github.com/bukanberuangsr/B1K5-API.git
cd B1K5-API
docker-compose up -d 
```

## Penjelasan Rute

Semua rute API dimulai mengikuti URL dasar:

```uri
http://localhost:8080/api
```

### Autentikasi

- POST api/login
- POST api/register

### User

- GET /api/users/:id
- GET /api/users/:id/activity
- GET /api/users/:id/segment

### Personalisasi & Rekomendasi

- GET /api/personalization/:id
- GET /api/recommendation/:id

### Analitik

- GET /api/analytics/metrics
- POST /api/analytics/event

### Segment

- POST /api/segments/update
