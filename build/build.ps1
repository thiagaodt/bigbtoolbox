<#
.SYNOPSIS
    Compila o BigB Toolbox para Windows e Linux.

.DESCRIPTION
    Passos:
      1. Confere se o Go esta instalado.
      2. Roda os testes nesta maquina (pule com -SemTestes) e o go vet para
         os dois sistemas (pega erro de compilacao do codigo so-Linux).
      3. Compila cmd\bigb para:
           dist\bigb.exe  Windows amd64
           dist\bigb      Linux amd64 (estatico, CGO desligado)

    A pasta scripts\ vai embutida nos dois. No Linux, depois de copiar:
    chmod +x bigb   (o NTFS nao guarda o bit de execucao).

.PARAMETER OutDir
    Pasta de saida (padrao: <raiz>\dist).

.PARAMETER SemTestes
    Nao roda "go test" antes de compilar.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File .\build\build.ps1
#>
[CmdletBinding()]
param(
    [string]$OutDir,
    [switch]$SemTestes
)

$ErrorActionPreference = 'Stop'

$raiz = Split-Path $PSScriptRoot -Parent
if (-not $OutDir) { $OutDir = Join-Path $raiz 'dist' }

# Alvos: sistema, arquitetura e nome do executavel.
$alvos = @(
    @{ GOOS = 'windows'; GOARCH = 'amd64'; Arquivo = 'bigb.exe' }
    @{ GOOS = 'linux';   GOARCH = 'amd64'; Arquivo = 'bigb' }
)

function Write-Step { param($m) Write-Host "==> $m" -ForegroundColor Cyan }

# ------------------------------------------------------------ 1. Go presente
Write-Step 'Verificando o Go...'
$go = Get-Command go -ErrorAction SilentlyContinue
if (-not $go) {
    Write-Host 'Go nao encontrado no PATH.' -ForegroundColor Red
    Write-Host 'Instale em https://go.dev/dl/ (Go 1.26 ou mais novo) e rode de novo.' -ForegroundColor Yellow
    exit 1
}
Write-Host ("    " + (& go version)) -ForegroundColor DarkGray

Push-Location $raiz
try {
    # -------------------------------------------------------- 2. testes
    if (-not $SemTestes) {
        Write-Step 'Rodando os testes...'
        & go test ./...
        if ($LASTEXITCODE -ne 0) { throw 'Testes falharam; build abortado.' }
    }

    foreach ($a in $alvos) {
        Write-Step ("go vet ({0})..." -f $a.GOOS)
        $env:GOOS = $a.GOOS; $env:GOARCH = $a.GOARCH
        & go vet ./...
        if ($LASTEXITCODE -ne 0) { throw ("go vet falhou para {0}." -f $a.GOOS) }
    }

    # ------------------------------------------------------ 3. compilar
    if (-not (Test-Path $OutDir)) { New-Item -ItemType Directory -Path $OutDir -Force | Out-Null }
    $env:CGO_ENABLED = '0'
    $gerados = @()
    foreach ($a in $alvos) {
        Write-Step ("Compilando para {0}/{1}..." -f $a.GOOS, $a.GOARCH)
        $env:GOOS = $a.GOOS; $env:GOARCH = $a.GOARCH
        $exe = Join-Path $OutDir $a.Arquivo
        # -s -w tiram tabela de simbolos e DWARF: binario bem menor.
        & go build -trimpath -ldflags '-s -w' -o $exe ./cmd/bigb
        if ($LASTEXITCODE -ne 0) { throw ("Falha na compilacao para {0}." -f $a.GOOS) }
        $gerados += $exe
    }

    Write-Host ''
    Write-Host 'Build concluido.' -ForegroundColor Green
    foreach ($exe in $gerados) {
        Write-Host ("  {0}  ({1:N1} MB)" -f $exe, ((Get-Item $exe).Length / 1MB))
    }
    Write-Host ''
    Write-Host 'No Linux, depois de copiar: chmod +x bigb' -ForegroundColor Yellow
}
finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Pop-Location
}
