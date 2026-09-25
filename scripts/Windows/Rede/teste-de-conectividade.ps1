# @nome      Teste de conectividade
# @descricao Gateway, DNS e acesso à internet
# @so        windows

$gateway = (Get-NetRoute -DestinationPrefix '0.0.0.0/0' -ErrorAction SilentlyContinue |
    Sort-Object RouteMetric | Select-Object -First 1).NextHop

$alvos = @(
    @{ Nome = 'Gateway';          Host = $gateway;    Porta = $null }
    @{ Nome = 'DNS Google';       Host = '8.8.8.8';   Porta = 53 }
    @{ Nome = 'DNS Cloudflare';   Host = '1.1.1.1';   Porta = 53 }
    @{ Nome = 'HTTPS google.com'; Host = 'google.com'; Porta = 443 }
)

foreach ($a in $alvos) {
    $rotulo = ('  {0} ' -f $a.Nome).PadRight(26, '.')
    if (-not $a.Host) { Write-Host "$rotulo sem gateway" -ForegroundColor Yellow; continue }
    if ($a.Porta) {
        $ok = Test-NetConnection $a.Host -Port $a.Porta -InformationLevel Quiet -WarningAction SilentlyContinue
    } else {
        $ok = Test-Connection $a.Host -Count 1 -Quiet
    }
    if ($ok) { Write-Host "$rotulo OK" -ForegroundColor Green }
    else     { Write-Host "$rotulo FALHOU" -ForegroundColor Red }
}

Write-Host ''
try {
    $ip = (Resolve-DnsName google.com -Type A -ErrorAction Stop | Select-Object -First 1).IPAddress
    Write-Host "  Resolução de nomes: google.com -> $ip" -ForegroundColor Green
} catch {
    Write-Host '  Resolução de nomes: FALHOU' -ForegroundColor Red
}
