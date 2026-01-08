# 🔧 Solusi Error: Unknown column 'i.duration_minutes'

## 📋 Masalah

Saat menjalankan API test, endpoint `/api/dashboard/production-stats` dan `/api/notifications` mengalami error 500:

```
Error 1054 (42S22): Unknown column 'i.duration_minutes' in 'SELECT'
```

## 🔍 Analisis Root Cause

Error terjadi karena di file `internal/repository/enhanced_instance_repo.go`, fungsi `GetProductionStats()` memiliki query SQL yang tidak konsisten:

### Query Bermasalah (Sebelum):
```sql
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,  -- ❌ Tidak ada alias 'i.'
    SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
    SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected,
    COALESCE(AVG(i.duration_minutes), 0) as avg_duration  -- ❌ Menggunakan 'i.' tapi lainnya tidak
FROM process_instances i
```

**Masalah:**
- Kolom `status` di query `CASE WHEN` tidak menggunakan alias tabel `i.`
- Sementara kolom `duration_minutes` menggunakan alias `i.`
- Ini menyebabkan inkonsistensi dan error pada beberapa versi MariaDB

## ✅ Solusi

### 1. Perbaiki File: `internal/repository/enhanced_instance_repo.go`

Ganti query di fungsi `GetProductionStats()` dengan:

```go
// Get total counts - FIXED: Added proper alias and handle NULL values
query := fmt.Sprintf(`
    SELECT 
        COUNT(*) as total,
        SUM(CASE WHEN i.status = 'completed' THEN 1 ELSE 0 END) as completed,
        SUM(CASE WHEN i.status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
        SUM(CASE WHEN i.status = 'rejected' THEN 1 ELSE 0 END) as rejected,
        COALESCE(AVG(CASE WHEN i.duration_minutes IS NOT NULL THEN i.duration_minutes ELSE 0 END), 0) as avg_duration
    FROM process_instances i
    %s
`, dateFilter)
```

**Perbaikan:**
1. ✅ Semua kolom sekarang menggunakan alias `i.` (konsisten)
2. ✅ Handle NULL value di `duration_minutes` dengan CASE statement
3. ✅ Query sekarang kompatibel dengan semua versi MariaDB

### 2. Perbaiki Query Lainnya

Di bagian bawah fungsi yang sama, perbaiki juga query untuk status dan priority:

```go
// Get by status
query = fmt.Sprintf(`
    SELECT i.status, COUNT(*) as count  -- ✅ Tambahkan 'i.' alias
    FROM process_instances i
    %s
    GROUP BY i.status  -- ✅ Tambahkan 'i.' alias
    ORDER BY count DESC
`, dateFilter)

// Get by priority
query = fmt.Sprintf(`
    SELECT i.priority, COUNT(*) as count  -- ✅ Tambahkan 'i.' alias
    FROM process_instances i
    %s
    GROUP BY i.priority  -- ✅ Tambahkan 'i.' alias
    ORDER BY count DESC
`, dateFilter)
```

## 🚀 Cara Deploy Perbaikan

### Opsi 1: Replace File Langsung

```bash
# Backup file lama
cp internal/repository/enhanced_instance_repo.go internal/repository/enhanced_instance_repo.go.backup

# Copy file yang sudah diperbaiki
cp /path/to/fixed/enhanced_instance_repo.go internal/repository/enhanced_instance_repo.go

# Rebuild aplikasi
go build -o server cmd/api/main.go

# Restart server
./server
```

### Opsi 2: Docker Compose (Recommended)

```bash
# Stop container yang sedang jalan
docker-compose down

# Rebuild dengan perubahan baru
docker-compose build --no-cache

# Jalankan ulang
docker-compose up -d

# Cek log
docker-compose logs -f app
```

### Opsi 3: Manual Edit

1. Buka file `internal/repository/enhanced_instance_repo.go`
2. Cari fungsi `GetProductionStats`
3. Ganti semua query seperti di atas
4. Save file
5. Rebuild dan restart

## 🧪 Testing Setelah Perbaikan

Jalankan test script PowerShell:

```powershell
.\test-api.ps1
```

Atau test manual menggunakan curl:

```bash
# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test_operator","password":"operator123"}'

# Test Production Stats (harusnya berhasil sekarang)
curl -X GET http://localhost:8080/api/dashboard/production-stats \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## ✅ Expected Result

Setelah perbaikan, endpoint akan mengembalikan response sukses:

```json
{
  "data": {
    "total_instances": 5,
    "completed_instances": 3,
    "in_progress_instances": 1,
    "rejected_instances": 1,
    "average_duration_minutes": 45.5,
    "by_template": [
      {
        "template_name": "Mixing",
        "count": 3,
        "percentage": 60.0
      },
      {
        "template_name": "Oven Curing",
        "count": 2,
        "percentage": 40.0
      }
    ],
    "by_status": [...],
    "by_priority": [...]
  }
}
```

## 📝 Catatan Penting

1. **Database Schema**: Pastikan tabel `process_instances` memiliki kolom `duration_minutes` (sudah ada di init.sql)
2. **NULL Handling**: Kolom `duration_minutes` bisa NULL, makanya perlu penanganan khusus
3. **Alias Konsisten**: Selalu gunakan alias tabel (`i.`, `t.`, dll) untuk menghindari ambiguitas
4. **Testing**: Selalu test di environment development dulu sebelum production

## 🎯 Kesimpulan

Error ini terjadi karena **inkonsistensi penggunaan alias tabel** dalam SQL query. Perbaikannya sederhana:
- Gunakan alias konsisten untuk semua kolom
- Handle NULL value dengan CASE statement
- Test ulang semua endpoint yang berkaitan

Setelah perbaikan ini, semua endpoint dashboard akan berfungsi normal! 🎉