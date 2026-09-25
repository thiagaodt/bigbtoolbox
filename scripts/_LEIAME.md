# Pasta de scripts do BigB

Tudo o que aparece no menu do BigB vem daqui, no Windows e no Linux. Guia
completo: [docs/SCRIPTS.md](../docs/SCRIPTS.md).

- Cada subpasta vira uma categoria; o prefixo numérico (`10-rede`) só ordena.
- Windows: `.bat`, `.cmd`, `.ps1`, `.vbs` (e `.py`/`.sh` se houver
  interpretador). Linux: `.sh`, `.py`, `.ps1` (com `pwsh`).
- Cabeçalho opcional em comentário: `@nome`, `@descricao`, `@admin`,
  `@confirmar`, `@so` (`windows` ou `linux`).
- Itens com `_` ou `.` no início não aparecem no menu, mas vão junto com os
  scripts da mesma pasta (use para instaladores e arquivos de apoio).

Este arquivo também garante que a pasta nunca fique vazia: o `go:embed`
recusa compilar uma pasta sem arquivos.
