# @nome      Teste de conectividade
# @descricao Gateway, DNS e acesso à internet
# @so        linux

testar() {
    rotulo=$1; shift
    if "$@" >/dev/null 2>&1; then
        printf '  %-24s OK\n' "$rotulo"
    else
        printf '  %-24s FALHOU\n' "$rotulo"
    fi
}

gateway=$(ip route 2>/dev/null | awk '/^default/ {print $3; exit}')
if [ -n "$gateway" ]; then
    testar "Gateway ($gateway)" ping -c 1 -W 2 "$gateway"
else
    printf '  %-24s sem gateway\n' "Gateway"
fi
testar "DNS Google"       timeout 3 bash -c '</dev/tcp/8.8.8.8/53'
testar "DNS Cloudflare"   timeout 3 bash -c '</dev/tcp/1.1.1.1/53'
testar "HTTPS google.com" timeout 3 bash -c '</dev/tcp/google.com/443'

echo
ip=$(getent hosts google.com | awk '{print $1; exit}')
if [ -n "$ip" ]; then
    echo "  Resolução de nomes: google.com -> $ip"
else
    echo "  Resolução de nomes: FALHOU"
    exit 1
fi
