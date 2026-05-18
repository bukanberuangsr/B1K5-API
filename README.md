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
docker compose up -d --build
```

Untuk reload docker image:

```sh
docker compose down
docker compose up -d
```

## Penjelasan Rute

Semua rute API dimulai mengikuti URL dasar:

```uri
http://localhost:8080/api
```

### Autentikasi

- POST api/login

  Request

  ```json
  {
    "customer_id": "CUS-000001",
    "password": "123456"
  }
  ```

  Response
  ```json
  {
    "customer_id": "CUS-000001",
    "message": "Login success",
  }
  ```

- POST api/register

  Request
  ```json
  {
    "full_name": "Arna",
    "email": "arna@mail.com",
    "password": "123456"
  }
  // atau
  [
    {
      "full_name": "Arna",
      "email": "arna@mail.com",
      "password": "123456"
    },
    {
      "full_name": "Maya",
      "email": "maya@mail.com",
      "password": "qwerty"
    } 
  ]
  ```
  
  Response

  ```json
  {
    "accounts": [
      {
        "id": 1,
        "customer_id": "CUS-000001",
        "email": "arna@mail.com"
      },
      {
        "id": 2,
        "customer_id": "CUS-000002",
        "email": "maya@mail.com"
      }
    ],
    "message": "Register success"
  }
  ```

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
