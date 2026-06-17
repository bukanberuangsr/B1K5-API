# API Capstone C1 - Kelompok 5

API untuk Capstone Project kasus B1. Dikelola oleh Kelompok 5.

## Menjalankan Project

Pastikan `Docker` dan `Docker Compose` sudah terpasang.

```sh
git clone https://www.github.com/bukanberuangsr/B1K5-API.git
cd B1K5-API
docker compose up -d --build
```

Untuk reload container:

```sh
docker compose down
docker compose up -d --build
```

Base URL:

```text
http://localhost:8080/api
```

Health check:

```http
GET /api/test
```

Panduan setup API untuk Flutter app tersedia di [usage.md](usage.md).

## Auth

API memakai JWT Bearer token. Endpoint selain register/login membutuhkan header:

```http
Authorization: Bearer <token>
```

Role yang digunakan:

```text
customer
admin
```

Aturan akses:

| Role | Akses |
| --- | --- |
| `customer` | Hanya boleh mengakses data miliknya sendiri |
| `admin` | Bisa mengakses semua user dan endpoint admin |

Admin seed:

```text
customer_id: ADM-000001
password: 123456
role: admin
```

Customer seed:

```text
CUS-000001 sampai CUS-000006
password: 123456
role: customer
```

`CUS-000006` sengaja tidak memiliki data segmentasi, tetapi memiliki aktivitas `transfer` dan `topup`.

## Autentikasi

### Register

```http
POST /api/auth/register
```

Request single account:

```json
{
  "full_name": "Arna Pratama",
  "username": "arna",
  "email": "arna@mail.com",
  "password": "123456"
}
```

Request multiple accounts:

```json
[
  {
    "full_name": "Arna Pratama",
    "username": "arna",
    "email": "arna@mail.com",
    "password": "123456"
  },
  {
    "full_name": "Maya Lestari",
    "username": "maya",
    "email": "maya@mail.com",
    "password": "qwerty"
  }
]
```

Response single account:

```json
{
  "message": "Register success",
  "customer_id": "CUS-000007",
  "role": "customer"
}
```

Response multiple accounts:

```json
{
  "message": "Register success",
  "accounts": [
    {
      "id": 7,
      "customer_id": "CUS-000007",
      "email": "arna@mail.com",
      "role": "customer"
    }
  ]
}
```

Catatan: register publik selalu membuat role `customer`. Role `admin` dibuat lewat seed atau update database.

### Login

```http
POST /api/auth/login
```

Request:

```json
{
  "customer_id": "CUS-000001",
  "password": "123456"
}
```

Response:

```json
{
  "message": "Login success",
  "customer_id": "CUS-000001",
  "role": "customer",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Contoh penggunaan token:

```sh
curl http://localhost:8080/api/users/CUS-000001 \
  -H "Authorization: Bearer <token>"
```

## User

Semua endpoint user membutuhkan JWT.

### Get All Users

```http
GET /api/users/
```

Akses: `admin`.

Response:

```json
{
  "message": "all users",
  "users": [
    {
      "id": 1,
      "customer_id": "CUS-000001",
      "email": "arna@mail.com",
      "username": "arna",
      "full_name": "Arna Pratama",
      "role": "customer",
      "created_at": "2026-05-26 10:00:00"
    }
  ]
}
```

### Get User

```http
GET /api/users/:id
```

Akses: pemilik akun atau `admin`.

`:id` bisa memakai `id` internal, misalnya `1`, atau `customer_id`, misalnya `CUS-000001`.

Response:

```json
{
  "message": "user found",
  "email": "arna@mail.com",
  "username": "arna",
  "role": "customer",
  "created_at": "2026-05-26 10:00:00"
}
```

### Get User Activity

```http
GET /api/users/:id/activity
```

Akses: pemilik akun atau `admin`.

Response:

```json
{
  "message": "user activity found",
  "data": {
    "customer_id": "CUS-000006",
    "transactions": [
      {
        "trx_id": "TRX-000006",
        "type": "transfer",
        "amount": 250000,
        "status": "success",
        "created_at": "2026-05-26 10:00:00"
      }
    ],
    "frequently_used_features": [
      {
        "feature": "transfer",
        "usage_count": 2,
        "last_used_at": "2026-05-26 10:00:00"
      }
    ]
  }
}
```

### Get User Segment

```http
GET /api/users/:id/segment
```

Akses: pemilik akun atau `admin`.

Response jika user memiliki segmentasi:

```json
{
  "message": "user segment found",
  "customer_id": "CUS-000001",
  "segments": [
    {
      "id": 1,
      "name": "investor",
      "description": "Nasabah dengan sinyal ketertarikan investasi dan saldo rata-rata tinggi.",
      "confidence": 0.91,
      "assigned_at": "2026-05-26 10:00:00",
      "recommendations": [
        {
          "id": 1,
          "feature": "promo reksa dana & deposito",
          "reason": "Profil nasabah menunjukkan ketertarikan pada produk investasi dan saldo rata-rata yang mendukung pertumbuhan aset.",
          "priority": 1
        }
      ]
    }
  ]
}
```

Response jika user belum memiliki segmentasi:

```json
{
  "error": "user segment not found"
}
```

## Personalisasi

### Get Homepage Personalization

```http
GET /api/personalization/:id
```

Akses: pemilik akun atau `admin`.

Endpoint ini menghasilkan konfigurasi personalisasi untuk homepage berdasarkan aktivitas, segmentasi, dan rekomendasi user.

Response:

```json
{
  "message": "personalization config found",
  "data": {
    "customer_id": "CUS-000001",
    "homepage": {
      "primary_feature": "investment",
      "segment": {
        "id": 1,
        "name": "investor",
        "confidence": 0.91
      },
      "quick_actions": [
        {
          "feature": "investment",
          "usage_count": 1,
          "last_used_at": "2026-05-26 10:00:00"
        }
      ],
      "recommended_sections": [
        {
          "id": 1,
          "feature": "promo reksa dana & deposito",
          "reason": "Profil nasabah menunjukkan ketertarikan pada produk investasi dan saldo rata-rata yang mendukung pertumbuhan aset.",
          "priority": 1
        }
      ]
    }
  }
}
```

Untuk user tanpa segmentasi seperti `CUS-000006`, `segment` akan kosong dan `recommended_sections` akan kosong, tetapi `quick_actions` tetap bisa terisi dari `user_activities`.

## Rekomendasi

```http
GET /api/recommendation/:id
GET /api/recommendations/:id
```

Akses: pemilik akun atau `admin`.

Kedua endpoint di atas memakai handler yang sama. `:id` bisa memakai `id` internal atau `customer_id`.

Endpoint ini mengembalikan rekomendasi berdasarkan segmentasi user. Jika user belum memiliki segmentasi, rekomendasi dibuat dari fitur yang paling sering digunakan user.

Response dari rekomendasi segmentasi:

```json
{
  "message": "recommendations found",
  "customer_id": "CUS-000001",
  "recommendations": [
    {
      "id": 1,
      "feature": "promo reksa dana & deposito",
      "reason": "Profil nasabah menunjukkan ketertarikan pada produk investasi dan saldo rata-rata yang mendukung pertumbuhan aset.",
      "priority": 1,
      "source": "segment",
      "segment": "investor",
      "confidence": 0.91
    }
  ]
}
```

Response fallback dari aktivitas user:

```json
{
  "message": "recommendations found",
  "customer_id": "CUS-000006",
  "recommendations": [
    {
      "feature": "transfer",
      "reason": "Fitur ini relevan karena sering digunakan oleh user.",
      "priority": 1,
      "source": "activity",
      "usage_count": 2
    }
  ]
}
```

## Analitik

Semua endpoint analytics membutuhkan JWT.

### Get Analytics Metrics

```http
GET /api/analytics/metrics
```

Akses: `admin`.

Response:

```json
{
  "message": "analytics metrics found",
  "metrics": {
    "total_events": 10,
    "impressions": 6,
    "clicks": 2,
    "engagements": 1,
    "ctr": 33.33,
    "conversion_rate": 50.0
  },
  "top_features": [
    {
      "feature": "investment",
      "event_count": 5,
      "clicks": 2
    }
  ]
}
```

Catatan field:

| Field | Keterangan |
| --- | --- |
| `impressions` | Jumlah event `recommendation_impression` / `recommendation_viewed` |
| `clicks` | Jumlah event `recommendation_click` / `recommendation_clicked` |
| `engagements` | Jumlah event `recommendation_engaged` / `feature_used` (user benar-benar memakai fitur) |
| `ctr` | `(clicks / impressions) × 100`, dibulatkan 2 desimal |
| `conversion_rate` | `(engagements / clicks) × 100`, dibulatkan 2 desimal |

### Create Analytics Event

```http
POST /api/analytics/event
```

Akses: user login atau `admin`.

Jika `customer_id` tidak dikirim, event akan dibuat untuk user dari token. User dengan role `customer` hanya boleh membuat event untuk dirinya sendiri. Role `admin` bisa membuat event untuk user lain.

Request:

```json
{
  "event_type": "recommendation_click",
  "feature": "investment",
  "metadata": {
    "source": "homepage"
  }
}
```

Request dengan `customer_id` eksplisit:

```json
{
  "customer_id": "CUS-000001",
  "event_type": "recommendation_click",
  "feature": "investment",
  "metadata": {
    "source": "homepage"
  }
}
```

Response:

```json
{
  "message": "analytics event created",
  "event": {
    "id": 1,
    "customer_id": "CUS-000001",
    "event_type": "recommendation_click",
    "feature": "investment",
    "metadata": {
      "source": "homepage"
    },
    "created_at": "2026-05-26 10:00:00"
  }
}
```

## Dashboard

Semua endpoint dashboard membutuhkan JWT dengan role `admin`.

### Engagement Dashboard

```http
GET /api/dashboard/engagement
```

Menampilkan 4 metrik utama efektivitas rekomendasi: CTR, conversion rate, total klik minggu ini, dan top 3 fitur yang paling sering diklik.

Response:

```json
{
  "message": "engagement dashboard data retrieved successfully",
  "data": {
    "ctr": 54.55,
    "conversion_rate": 50.0,
    "total_clicks_this_week": 6,
    "top_recommended_features": [
      { "rank": 1, "feature": "tabungan_emas", "total_clicks": 3 },
      { "rank": 2, "feature": "asuransi_jiwa", "total_clicks": 1 },
      { "rank": 3, "feature": "reksa_dana",    "total_clicks": 1 }
    ],
    "period": "2026-06-10 to 2026-06-17"
  }
}
```

### Performance Dashboard

```http
GET /api/dashboard/performance?days=30
```

Query param `days` menentukan rentang waktu (default: `30`). Menampilkan breakdown per segment, hasil A/B test aktif, ROI rekomendasi, dan top 5 segment berdasarkan engagement rate.

Response:

```json
{
  "message": "dashboard performance data retrieved successfully",
  "data": {
    "overall_metrics": {
      "total_customers": 6,
      "active_customers": 5,
      "engagement_rate": 0.83,
      "avg_recommendation_ctr": 0,
      "total_recommendations": 6,
      "successful_recommendations": 0,
      "date_range": "2026-05-18 to 2026-06-17"
    },
    "engagement_by_segment": [
      {
        "segment_name": "digital_spender",
        "customer_count": 2,
        "active_customers": 2,
        "engagement_rate": 1,
        "recommendation_impressions": 0,
        "recommendation_clicks": 0,
        "ctr": 0
      }
    ],
    "ab_test_results": [],
    "recommendation_roi": {
      "total_recommendations": 9,
      "clicked_recommendations": 2,
      "roi_percentage": 22.22,
      "estimated_value_idr": 20000
    },
    "top_performing_segments": [
      {
        "segment_name": "investor",
        "engagement_rate": 1,
        "customer_count": 1,
        "rank": 1
      }
    ]
  }
}
```

## Segment

Ada dua cara update segmentasi user: batch manual (`/api/segments/update`) dan prediksi real-time per user lewat model ML (`/api/users/:id/segment/predict`).

### Update Manual / Batch

```http
POST /api/segments/update
```

Akses: `admin`.

Endpoint ini melakukan insert/update segmentasi user berdasarkan nilai yang dikirim langsung di request body. Dipakai oleh `assets/sync_segments.py` untuk sinkronisasi batch/cron.

Request:

```json
{
  "segments": [
    {
      "customer_id": "CUS-000001",
      "segment_name": "investor",
      "description": "Nasabah dengan sinyal ketertarikan investasi dan saldo rata-rata tinggi.",
      "confidence": 0.91
    }
  ]
}
```

`confidence` harus berada di rentang `0` sampai `1`.

Response:

```json
{
  "message": "user segments updated",
  "updated": 1,
  "results": [
    {
      "customer_id": "CUS-000001",
      "segment_name": "investor",
      "confidence": 0.91,
      "action": "updated"
    }
  ]
}
```

### Prediksi Real-time (ML)

```http
POST /api/users/:id/segment/predict
```

Akses: pemilik akun atau `admin`.

Endpoint ini mengambil data transaksi & fitur user dari database, mengirimkannya ke service ML (`ml-service`, lihat [Arsitektur ML](#arsitektur-ml)) untuk diproses lewat pipeline `scaler → PCA → KMeans`, lalu hasil prediksinya langsung di-upsert ke `segments` dan `user_segments` — tanpa perlu kirim `segment_name`/`confidence` manual seperti endpoint batch di atas.

Response:

```json
{
  "message": "segment predicted and updated",
  "customer_id": "CUS-000001",
  "segment_name": "investor",
  "description": "Nasabah dengan sinyal ketertarikan investasi dan saldo rata-rata tinggi. Fitur dominan: investment. Total aktivitas terdeteksi: 2.",
  "confidence": 0.47,
  "action": "updated"
}
```

`action` bernilai `"updated"` jika user sudah punya segmen sebelumnya, atau `"inserted"` jika ini segmentasi pertamanya.

## Status Code Umum

| Status | Arti |
| --- | --- |
| `200` | Request berhasil |
| `201` | Data berhasil dibuat |
| `400` | Request body tidak valid |
| `401` | Token tidak ada, token tidak valid, atau login gagal |
| `403` | Role tidak punya akses |
| `404` | Data tidak ditemukan |
| `500` | Error server/database |

## Arsitektur ML

Model segmentasi (`scaler.pkl`, `pca_model.pkl`, `kmeans_model.pkl`, `segment_map.pkl` di [machine-learning/](machine-learning/)) hasil training scikit-learn, dibungkus jadi service HTTP terpisah supaya bisa dipanggil dari Go API:

```text
Go API (b1k5-api) --HTTP--> ML Service (b1k5-ml-service, FastAPI) --joblib--> scaler/PCA/KMeans
```

Kedua container ini sudah otomatis jalan lewat `docker compose up -d --build`, satu network, dan saling kenal lewat nama container (`ML_SERVICE_URL=http://ml-service:8000`, lihat [docker-compose.yml](docker-compose.yml)).

Health check ML service:

```sh
curl http://localhost:8000/health
```

Source code ML ada di [assets/](assets/):

| File | Fungsi |
| --- | --- |
| `b1k5_capstone_model.py` | Ekstraksi fitur dari data transaksi + load pipeline scaler/PCA/KMeans + prediksi segmen |
| `ml_service.py` | FastAPI wrapper, expose `POST /predict` untuk dipanggil Go API |
| `sync_segments.py` | Script batch/cron, manggil model secara langsung (tanpa lewat Go API) lalu POST ke `/api/segments/update` |

**Catatan kualitas model**: model KMeans yang dipakai punya silhouette score ~0.14 (cluster overlap cukup tinggi, lihat `(Kasaran)_Capstone_K5_B1.ipynb`), dan 7 dari 17 fitur training (login frequency, durasi sesi, dll) tidak punya data di skema produksi sehingga diisi nilai default. Confidence yang dihasilkan endpoint prediksi karena itu cenderung lebih rendah dan kurang stabil dibanding nilai manual — cocok untuk demo, belum untuk keputusan bisnis nyata.

## Catatan Segmentasi

Segmentasi bisa berasal dari pola penggunaan fitur, misalnya `transfer`, `topup`, `bill_payment`, atau `investment`. Namun segmentasi sebaiknya tidak hanya memakai jumlah penggunaan mentah. Kombinasikan dengan frekuensi, rasio fitur, waktu terakhir aktif, nominal transaksi, login frequency, dan durasi sesi agar hasilnya lebih representatif.
