# @nome      Limpar temporários
# @descricao %TEMP% do usuário e Windows\Temp (pula arquivos em uso)
# @admin     sim
# @confirmar sim
# @so        windows

function Get-Tamanho($pasta) {
    [double](Get-ChildItem $pasta -Recurse -Force -File -ErrorAction SilentlyContinue |
        Measure-Object Length -Sum).Sum
}

foreach ($pasta in @($env:TEMP, (Join-Path $env:SystemRoot 'Temp'))) {
    if (-not (Test-Path $pasta)) { continue }
    $antes = Get-Tamanho $pasta
    Get-ChildItem $pasta -Force -ErrorAction SilentlyContinue |
        Remove-Item -Recurse -Force -ErrorAction SilentlyContinue
    $liberado = ($antes - (Get-Tamanho $pasta)) / 1MB
    Write-Host ('  {0}: {1:N1} MB liberados' -f $pasta, $liberado) -ForegroundColor Green
}
