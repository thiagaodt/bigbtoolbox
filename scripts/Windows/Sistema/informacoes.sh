# @nome      Informações do sistema
# @descricao Distribuição, kernel, processador, memória e discos
# @so        linux

[ -r /etc/os-release ] && . /etc/os-release

echo "  Máquina ....: $(hostname)"
echo "  Sistema ....: ${PRETTY_NAME:-$(uname -s)} (kernel $(uname -r))"
echo "  Processador : $(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//')"
echo "  Memória ....: $(free -h | awk '/^Mem:/ {print $2 " total, " $7 " disponível"}')"
echo "  Ligado desde: $(uptime -s 2>/dev/null)"
echo
df -h -x tmpfs -x devtmpfs -x squashfs -x overlay 2>/dev/null
