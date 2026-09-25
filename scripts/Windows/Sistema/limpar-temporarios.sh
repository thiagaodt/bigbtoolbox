# @nome      Limpar temporários
# @descricao /tmp e /var/tmp com mais de 7 dias, e o journal acima de 200 MB
# @admin     sim
# @confirmar sim
# @so        linux

for pasta in /tmp /var/tmp; do
    antes=$(du -sm "$pasta" 2>/dev/null | cut -f1)
    find "$pasta" -mindepth 1 -mtime +7 -delete 2>/dev/null
    depois=$(du -sm "$pasta" 2>/dev/null | cut -f1)
    echo "  $pasta: $((antes - depois)) MB liberados"
done

if command -v journalctl >/dev/null; then
    echo
    journalctl --vacuum-size=200M
fi
