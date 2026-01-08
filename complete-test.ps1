# ============================================
# Quick Fix Test - With User Registration
# ============================================
$baseUrl = "http://localhost:8080"

Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  🧪 PT Besq - Testing Fixed Endpoints                 ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# ============================================
# 1. HEALTH CHECK
# ============================================
Write-Host "1️⃣  Health Check..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/api/health" -UseBasicParsing
    Write-Host "   ✅ Server: $($health.status) - $($health.system)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "   ❌ Server tidak bisa diakses!" -ForegroundColor Red
    Write-Host "   💡 Jalankan: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

# ============================================
# 2. REGISTER USER (Skip if already exists)
# ============================================
Write-Host "2️⃣  Register User..." -ForegroundColor Yellow
$registerBody = @{
    username = "test_operator"
    password = "operator123"
    email = "test@besq.com"
    full_name = "Test Operator"
    role = "operator"
} | ConvertTo-Json

try {
    $register = Invoke-RestMethod -Uri "$baseUrl/api/auth/register" `
        -Method Post `
        -Body $registerBody `
        -ContentType "application/json" `
        -UseBasicParsing
    Write-Host "   ✅ User registered: $($register.username)" -ForegroundColor Green
    Write-Host ""
} catch {
    if ($_.Exception.Message -like "*sudah dipakai*") {
        Write-Host "   ⚠️  User sudah ada, skip register" -ForegroundColor Yellow
        Write-Host ""
    } else {
        Write-Host "   ❌ Register error: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host ""
    }
}

# ============================================
# 3. LOGIN
# ============================================
Write-Host "3️⃣  Login..." -ForegroundColor Yellow
$loginBody = @{
    username = "test_operator"
    password = "operator123"
} | ConvertTo-Json

try {
    $login = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" `
        -Method Post `
        -Body $loginBody `
        -ContentType "application/json" `
        -UseBasicParsing
    
    $token = $login.token
    $headers = @{ Authorization = "Bearer $token" }
    
    Write-Host "   ✅ Login berhasil! Role: $($login.role)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "   ❌ Login gagal: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    exit 1
}

# ============================================
# 4. TEST PRODUCTION STATS (Yang Error Kemarin)
# ============================================
Write-Host "4️⃣  Testing Production Stats (YANG ERROR KEMARIN)..." -ForegroundColor Yellow
try {
    $prodStats = Invoke-RestMethod -Uri "$baseUrl/api/dashboard/production-stats" `
        -Method Get `
        -Headers $headers `
        -UseBasicParsing
    
    Write-Host "   ✅✅✅ BERHASIL! ERROR SUDAH FIX! ✅✅✅" -ForegroundColor Green
    Write-Host ""
    Write-Host "   📊 Production Statistics:" -ForegroundColor Cyan
    Write-Host "      Total Instances     : $($prodStats.data.total_instances)" -ForegroundColor White
    Write-Host "      Completed          : $($prodStats.data.completed_instances)" -ForegroundColor White
    Write-Host "      In Progress        : $($prodStats.data.in_progress_instances)" -ForegroundColor White
    Write-Host "      Rejected           : $($prodStats.data.rejected_instances)" -ForegroundColor White
    Write-Host "      Avg Duration (min) : $([math]::Round($prodStats.data.average_duration_minutes, 2))" -ForegroundColor White
    Write-Host ""
    
    if ($prodStats.data.by_template.Count -gt 0) {
        Write-Host "   📋 By Template:" -ForegroundColor Cyan
        foreach ($tmpl in $prodStats.data.by_template) {
            Write-Host "      • $($tmpl.template_name): $($tmpl.count) ($([math]::Round($tmpl.percentage, 1))%)" -ForegroundColor White
        }
        Write-Host ""
    }
    
} catch {
    Write-Host "   ❌❌❌ MASIH ERROR! ❌❌❌" -ForegroundColor Red
    Write-Host ""
    
    # Get detailed error
    try {
        $errorDetail = $_.Exception.Response
        $reader = New-Object System.IO.StreamReader($errorDetail.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "   Error Detail: $responseBody" -ForegroundColor Red
        Write-Host ""
        Write-Host "   💡 Solusi:" -ForegroundColor Yellow
        Write-Host "      1. Pastikan file enhanced_instance_repo.go sudah di-update" -ForegroundColor Yellow
        Write-Host "      2. Rebuild: docker-compose down && docker-compose build --no-cache && docker-compose up -d" -ForegroundColor Yellow
        Write-Host ""
    } catch {}
}

# ============================================
# 5. TEST NOTIFICATIONS (Yang Error Kemarin)
# ============================================
Write-Host "5️⃣  Testing Notifications (YANG ERROR KEMARIN)..." -ForegroundColor Yellow
try {
    $notifs = Invoke-RestMethod -Uri "$baseUrl/api/notifications" `
        -Method Get `
        -Headers $headers `
        -UseBasicParsing
    
    Write-Host "   ✅✅✅ BERHASIL! ERROR SUDAH FIX! ✅✅✅" -ForegroundColor Green
    Write-Host ""
    Write-Host "   🔔 Notifications:" -ForegroundColor Cyan
    Write-Host "      Unread Count: $($notifs.unread_count)" -ForegroundColor White
    Write-Host "      Total       : $($notifs.data.Count)" -ForegroundColor White
    Write-Host ""
    
    if ($notifs.data.Count -gt 0) {
        Write-Host "   📬 Recent Notifications:" -ForegroundColor Cyan
        foreach ($notif in $notifs.data | Select-Object -First 3) {
            $readStatus = if ($notif.is_read) { "✓" } else { "•" }
            Write-Host "      $readStatus [$($notif.type)] $($notif.title)" -ForegroundColor White
        }
        Write-Host ""
    }
    
} catch {
    Write-Host "   ❌❌❌ MASIH ERROR! ❌❌❌" -ForegroundColor Red
    Write-Host ""
    
    # Get detailed error
    try {
        $errorDetail = $_.Exception.Response
        $reader = New-Object System.IO.StreamReader($errorDetail.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "   Error Detail: $responseBody" -ForegroundColor Red
        Write-Host ""
    } catch {}
}

# ============================================
# 6. TEST CREATE INSTANCE (Bonus Test)
# ============================================
Write-Host "6️⃣  Bonus Test: Create New Instance..." -ForegroundColor Yellow
$instanceBody = @{
    template_id = 1
    workflow_id = 1
    data = @{
        batch_code = "TEST-FIX-$(Get-Date -Format 'HHmmss')"
        rubber_weight = 95.5
        temperature = 185
        operator_notes = "Testing after fix"
    }
} | ConvertTo-Json

try {
    $instance = Invoke-RestMethod -Uri "$baseUrl/api/instances" `
        -Method Post `
        -Body $instanceBody `
        -ContentType "application/json" `
        -Headers $headers `
        -UseBasicParsing
    
    Write-Host "   ✅ Instance created! ID: $($instance.id)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "   ⚠️  Create instance: $($_.Exception.Message)" -ForegroundColor Yellow
    Write-Host ""
}

# ============================================
# SUMMARY
# ============================================
Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║  🎉 TESTING COMPLETE!                                  ║" -ForegroundColor Green
Write-Host "╠════════════════════════════════════════════════════════╣" -ForegroundColor Green
Write-Host "║  Jika Production Stats & Notifications berhasil,       ║" -ForegroundColor Green
Write-Host "║  berarti ERROR SUDAH FIX! ✅                           ║" -ForegroundColor Green
Write-Host "║                                                        ║" -ForegroundColor Green
Write-Host "║  Jika masih error, ikuti instruksi di                 ║" -ForegroundColor Green
Write-Host "║  CARA_FIX_MANUAL.md                                    ║" -ForegroundColor Green
Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""