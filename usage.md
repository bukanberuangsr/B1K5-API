# API Usage

Panduan singkat untuk menjalankan B1K5 API dan menghubungkannya ke Flutter app seperti `b1k5-mobile`.

## 1. Jalankan API

```sh
git clone https://www.github.com/bukanberuangsr/B1K5-API.git
cd B1K5-API
docker compose up -d --build
```

Cek API:

```sh
curl http://localhost:8080/api/test
```

Jika berhasil:

```json
{
  "message": "The API is currently running!"
}
```

Service lokal:

| Service | URL |
| --- | --- |
| API | `http://localhost:8080/api` |
| Adminer | `http://localhost:8081` |
| PostgreSQL | `localhost:5433` |

Untuk restart API:

```sh
docker compose down
docker compose up -d --build
```

## 2. Pilih Base URL Flutter

Gunakan base URL sesuai target Flutter.

| Target | Base URL |
| --- | --- |
| Android Emulator | `http://10.0.2.2:8080/api` |
| iOS Simulator | `http://localhost:8080/api` |
| Flutter Desktop | `http://localhost:8080/api` |
| HP fisik | `http://<IP-LAPTOP>:8080/api` |

Contoh untuk Android Emulator:

```dart
const baseUrl = 'http://10.0.2.2:8080/api';
```

Contoh untuk HP fisik:

```dart
const baseUrl = 'http://192.168.1.20:8080/api';
```

HP fisik dan laptop harus berada di WiFi yang sama.

## 3. Tambahkan Dependency Flutter

Di `pubspec.yaml`:

```yaml
dependencies:
  http: ^1.2.2
  shared_preferences: ^2.3.2
```

Lalu jalankan:

```sh
flutter pub get
```

## 4. Buat API Service

Buat file `lib/services/api_service.dart` di Flutter app.

```dart
import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

class ApiService {
  static const String baseUrl = 'http://10.0.2.2:8080/api';

  Future<Map<String, dynamic>> login(String customerId, String password) async {
    final response = await http.post(
      Uri.parse('$baseUrl/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'customer_id': customerId,
        'password': password,
      }),
    );

    final data = _handleResponse(response);

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('token', data['token']);
    await prefs.setString('customer_id', data['customer_id']);
    await prefs.setString('role', data['role']);

    return data;
  }

  Future<Map<String, dynamic>> getUser(String customerId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/users/$customerId'),
      headers: await _authHeaders(),
    );

    return _handleResponse(response);
  }

  Future<Map<String, dynamic>> getPersonalization(String customerId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/personalization/$customerId'),
      headers: await _authHeaders(),
    );

    return _handleResponse(response);
  }

  Future<Map<String, dynamic>> getRecommendation(String customerId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/recommendation/$customerId'),
      headers: await _authHeaders(),
    );

    return _handleResponse(response);
  }

  Future<Map<String, String>> _authHeaders() async {
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString('token');

    return {
      'Content-Type': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  Map<String, dynamic> _handleResponse(http.Response response) {
    final data = jsonDecode(response.body) as Map<String, dynamic>;

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception(data['error'] ?? 'Request failed');
    }

    return data;
  }
}
```

## 5. Pakai Di Flutter

Login:

```dart
final api = ApiService();

await api.login('CUS-000001', '123456');
```

Ambil data user:

```dart
final user = await api.getUser('CUS-000001');
```

Ambil personalisasi homepage:

```dart
final personalization = await api.getPersonalization('CUS-000001');
```

Ambil rekomendasi:

```dart
final recommendation = await api.getRecommendation('CUS-000001');
```

## 6. Akun Test

Customer:

```text
customer_id: CUS-000001
password: 123456
```

Admin:

```text
customer_id: ADM-000001
password: 123456
```

## 7. Endpoint Utama

| Method | Endpoint | Auth |
| --- | --- | --- |
| `GET` | `/api/test` | Tidak |
| `POST` | `/api/auth/register` | Tidak |
| `POST` | `/api/auth/login` | Tidak |
| `GET` | `/api/users/:id` | Ya |
| `GET` | `/api/users/:id/activity` | Ya |
| `GET` | `/api/users/:id/segment` | Ya |
| `GET` | `/api/personalization/:id` | Ya |
| `GET` | `/api/recommendation/:id` | Ya |
| `POST` | `/api/analytics/event` | Ya |

Endpoint dengan `Auth: Ya` wajib mengirim header:

```http
Authorization: Bearer <token>
```

`:id` bisa memakai `customer_id`, misalnya `CUS-000001`.

## 8. Catatan Penting

Jika Android menolak HTTP lokal, tambahkan ini di `android/app/src/main/AndroidManifest.xml` untuk development:

```xml
<application
    android:usesCleartextTraffic="true"
    ...>
</application>
```

Jika Flutter Web terkena error CORS, backend perlu ditambah middleware CORS di `cmd/api/main.go`.

Jika mendapat:

| Error | Penyebab umum |
| --- | --- |
| `401` | Belum login, token kosong, atau token salah |
| `403` | User mencoba akses data milik user lain |
| Connection refused | API belum jalan atau base URL salah |
