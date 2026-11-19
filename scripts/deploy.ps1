<#
.SYNOPSIS
    Сборка и выкладка vacation-calculation на сервер.

.DESCRIPTION
    1. go vet и go test (пропуск: -SkipTests).
    2. Сборка под linux/amd64 (CGO не нужен: modernc.org/sqlite).
    3. Загрузка бинарника рядом с текущим, проверка контрольной суммы.
    4. Копия базы (/usr/local/sbin/vacation-backup) и прежнего бинарника.
    5. Замена бинарника и перезапуск сервиса.
    6. Проверка /health; если сервис не поднялся — возврат прежнего бинарника.
       Миграции при этом не откатываются: при необходимости база восстанавливается
       из копии в /var/backups/vacation-calculation.

    Нужны plink и pscp (PuTTY). Вход по ключу через Pageant или по паролю из
    переменной окружения VACATION_SSH_PASSWORD.

.EXAMPLE
    $env:VACATION_SSH_PASSWORD = '...'; .\scripts\deploy.ps1
#>
param(
    [string]$Server = '109.66.243.42',
    [string]$User = 'root',
    [string]$HostKey = 'SHA256:hp0kz6LVORSvXOollGsUAjUAjY7kMWXMCuA3BF65eiQ',
    [switch]$SkipTests
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$sshArgs = @('-batch', '-hostkey', $HostKey)
if ($env:VACATION_SSH_PASSWORD) { $sshArgs += @('-pw', $env:VACATION_SSH_PASSWORD) }
$target = "$User@$Server"

function Invoke-Remote([string]$script) {
    $file = New-TemporaryFile
    try {
        [IO.File]::WriteAllText($file, $script.Replace("`r`n", "`n"), (New-Object Text.UTF8Encoding $false))
        & plink @sshArgs -ssh $target -m $file
        if ($LASTEXITCODE -ne 0) { throw "удалённая команда завершилась с кодом $LASTEXITCODE" }
    } finally {
        Remove-Item $file -Force
    }
}

if (-not $SkipTests) {
    Write-Host '==> go vet и go test'
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw 'go vet завершился с ошибкой' }
    go test ./...
    if ($LASTEXITCODE -ne 0) { throw 'тесты не прошли' }
}

Write-Host '==> Сборка linux/amd64'
$binary = Join-Path ([IO.Path]::GetTempPath()) 'vacation-calculation-linux'
$env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
try {
    go build -trimpath -ldflags '-s -w' -o $binary ./cmd/api
    if ($LASTEXITCODE -ne 0) { throw 'сборка не удалась' }
} finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
}
$hash = (Get-FileHash $binary -Algorithm SHA256).Hash.ToLower()

Write-Host '==> Загрузка'
& pscp @sshArgs -q $binary "${target}:/opt/vacation-calculation/vacation-calculation.new"
if ($LASTEXITCODE -ne 0) { throw 'загрузка не удалась' }

Write-Host '==> Замена и перезапуск'
$remote = @'
set -e
cd /opt/vacation-calculation
echo '__HASH__  vacation-calculation.new' | sha256sum -c --quiet -
/usr/local/sbin/vacation-backup
TS=$(date +%Y%m%d-%H%M%S)
mkdir -p /root/backups
cp -p vacation-calculation /root/backups/vacation-calculation-$TS
ls -1t /root/backups/vacation-calculation-* | tail -n +6 | xargs -r rm --
chmod 755 vacation-calculation.new
mv vacation-calculation.new vacation-calculation
systemctl restart vacation-calculation

for i in 1 2 3 4 5 6 7 8 9 10; do
    sleep 1
    if curl -fsS -o /dev/null http://127.0.0.1/health; then
        echo "сервис работает, копия прежней версии: /root/backups/vacation-calculation-$TS"
        exit 0
    fi
done

echo 'сервис не ответил на /health — возвращаю прежнюю версию' >&2
journalctl -u vacation-calculation --since '-1min' --no-pager | tail -20 >&2
cp -p /root/backups/vacation-calculation-$TS vacation-calculation
systemctl restart vacation-calculation
exit 1
'@
Invoke-Remote $remote.Replace('__HASH__', $hash)

Remove-Item $binary -Force
Write-Host '==> Готово'
