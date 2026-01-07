# 📘 PT Besq Backend API Documentation

**Version:** 1.1.0 (Security Update)  
**Base URL:** `http://localhost:8080`  
**Content-Type:** `application/json`  
**Auth Mechanism:** JWT (Bearer Token)

---

## 🔐 1. Authentication & Users
Bagian ini bersifat **Public**. Digunakan untuk mendapatkan Token akses.

### **POST** `/api/auth/register`
Mendaftarkan user baru (Admin/Operator).
* **Body:**
```json
{
  "username": "budi_operator",
  "password": "rahasia123",
  "role": "operator" 
}

```

* **Roles:** `admin` (Full Access) atau `operator` (Input Only).

### **POST** `/api/auth/login`

Masuk sistem untuk mendapatkan Token.

* **Body:**

```json
{
  "username": "budi_operator",
  "password": "rahasia123"
}

```

* **Response (200 OK):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR...",
  "role": "operator"
}

```

> ⚠️ **PENTING:** Token ini harus dikirim di **Header** untuk semua request di bawah ini.
> **Format Header:** `Authorization: Bearer <token_anda>`

---

## 🏭 2. Process Templates (Master Data)

*Access Level: **All Logged In Users***

### **GET** `/api/templates`

Mengambil daftar menu proses.

* **Response:** List of templates (Mixing, Oven, etc).

### **GET** `/api/templates/:id/fields`

Mengambil konfigurasi form input (Form Schema).

* **Response:**

```json
{
  "template_id": 1,
  "fields": [
    { "key": "batch_code", "label": "Kode Batch", "type": "text", "required": true }
  ]
}

```

---

## 📐 3. Workflow Engine (Diagram)

*Access Level: **Mixed***

### **GET** `/api/workflows`

* **Permission:** All Users
* **Description:** Melihat diagram alur produksi.

### **POST** `/api/workflows`

* **Permission:** ⛔ **ADMIN ONLY**
* **Description:** Membuat diagram baru. Jika Operator mencoba akses, akan mendapat error `403 Forbidden`.

### **PUT** `/api/workflows/:id/layout`

* **Permission:** ⛔ **ADMIN ONLY**
* **Description:** Menyimpan perubahan posisi/layout diagram.

---

## 📝 4. Process Instances (Input Data)

*Access Level: **Operator & Admin***

### **POST** `/api/instances`

Submit data produksi harian.

* **Header:** `Authorization: Bearer <token>`
* **Body:**

```json
{
  "template_id": 1,
  "workflow_id": 1,
  "data": {
    "batch_code": "BATCH-001", 
    "rubber_weight": 50.5
  }
}

```

* **Response (201 Created):**

```json
{
  "message": "Data valid, saved & broadcasted",
  "id": 105,
  "timestamp": "..."
}

```

---

## ⚡ 5. WebSocket (Realtime Dashboard)

* **URL:** `ws://localhost:8080/ws`
* **Description:** Mendengarkan update data secara realtime.
* **Event Payload:**

```json
{
  "event": "new_data",
  "instance_id": 105,
  "workflow_id": 1,
  "template_id": 1,
  "status": "draft",
  "timestamp": "..."
}

```
---

## 📊 6. Dashboard Analytics
*Access Level: **All Logged In Users***

### **GET** `/api/dashboard/stats`
Mengambil ringkasan data produksi hari ini untuk ditampilkan dalam bentuk Grafik (Pie Chart/Bar Chart) dan Kartu Statistik.

* **Response (200 OK):**
```json
{
  "date": "today",
  "total_today": 15,   // Tampilkan di "Big Number Widget"
  "breakdown": [       // Gunakan array ini untuk source Pie Chart
    {
      "TemplateName": "Mixing",
      "Count": 10
    },
    {
      "TemplateName": "Oven Curing",
      "Count": 5
    }
  ]
}
```
---

## 🛡️ Error Dictionary

Daftar kode error yang mungkin muncul terkait keamanan:

| Status Code | Message | Penyebab |
| --- | --- | --- |
| **401 Unauthorized** | `Butuh token autentikasi` | Header Authorization kosong. |
| **401 Unauthorized** | `Token tidak valid` | Token expired atau salah. |
| **403 Forbidden** | `Akses Ditolak` | Role user tidak cukup (misal: Operator coba edit Diagram). |

