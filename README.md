# BigB Toolbox

Lançador de scripts de suporte técnico para **Windows e Linux**, escrito em
**Go**.

O BigB não tem funcionalidade própria: **tudo o que aparece no menu vem da
pasta [`scripts/`](scripts/)**. Cada subpasta vira uma categoria e cada
script vira uma opção, navegável com as setas. Os scripts vão embutidos no
executável: o `bigb.exe` (ou `bigb`, no Linux) sozinho leva a caixa de
ferramentas inteira para qualquer máquina.

```
scripts/
  10-rede/
    flush-dns.bat               ->  Rede > Limpar cache DNS            (Windows)
    teste-de-conectividade.ps1  ->  Rede > Teste de conectividade      (Windows)
    teste-de-conectividade.sh   ->  Rede > Teste de conectividade      (Linux)
  20-sistema/
    informacoes.ps1             ->  Sistema > Informações do sistema   (Windows)
    informacoes.sh              ->  Sistema > Informações do sistema   (Linux)
```

## Formatos suportados

| Extensão | Windows | Linux |
|---|---|---|
| `.bat`, `.cmd` | `cmd.exe` | — |
| `.ps1` | `powershell.exe` (ou `pwsh.exe`) | `pwsh`, se instalado |
| `.vbs` | `cscript.exe` (VBScript em descontinuação) | — |
| `.sh` | `bash.exe`, se houver Git Bash ou WSL | `bash` (ou `sh`) |
| `.py` | `py.exe` / `python.exe`, se instalado | `python3` |

Um formato que o sistema não roda simplesmente não aparece no menu. Para
limitar um script a um sistema, use `@so windows` ou `@so linux` no
cabeçalho. Guia completo: [docs/SCRIPTS.md](docs/SCRIPTS.md).

## Como rodar

**Windows**: pede elevação sozinho (UAC).

```bat
dist\bigb.exe
dist\bigb.exe -sem-elevacao
```

**Linux**: roda como o seu usuário; use `sudo` para scripts marcados com
`@admin`.

```bash
chmod +x bigb
./bigb
sudo ./bigb
```

| Tecla | Ação |
|---|---|
| ↑ ↓ (Home, End) | Mover a seleção |
| Enter ou → | Abrir a categoria / executar o script |
| Esc, ← ou Backspace | Voltar |
| Q ou Ctrl+C | Sair |

Para testar scripts sem recompilar, aponte uma pasta externa:
`bigb -scripts <pasta>`.

Rodando do código-fonte (recompila e reembute `scripts/` a cada execução):

```bash
go run ./cmd/bigb -sem-elevacao
```

Para gerar os executáveis, veja [docs/BUILD.md](docs/BUILD.md).

## Documentação

- [docs/SCRIPTS.md](docs/SCRIPTS.md) — como adicionar e escrever scripts.
- [docs/ARQUITETURA.md](docs/ARQUITETURA.md) — pacotes, fluxo e decisões.
- [docs/BUILD.md](docs/BUILD.md) — rodar do código e gerar os executáveis.
