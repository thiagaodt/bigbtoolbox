# Scripts — como adicionar

Tudo o que o BigB faz vem da pasta `scripts/` na raiz do projeto. Colou um
script lá e recompilou, ele aparece no menu.

## Pastas viram categorias

```
scripts/
  10-rede/                  ->  categoria "Rede"
    flush-dns.bat           ->  item "Flush dns" (ou o @nome do cabeçalho)
    wifi/                   ->  subcategoria "Rede > Wifi"
      perfis.ps1
  20-sistema/               ->  categoria "Sistema"
  avulso.bat                ->  item direto no menu inicial
```

- **Ordem:** alfabética pelo nome em disco. Um prefixo numérico (`10-`,
  `20_`, `01.`) ordena e some na exibição.
- **Nome exibido:** hífens e sublinhados viram espaço e a primeira letra
  fica maiúscula (`reset-de-rede` → "Reset de rede"). Acentos e espaços
  no nome da pasta ou do arquivo funcionam.
- **Pastas sem scripts** (nem nas subpastas) não aparecem.

## Formatos

| Extensão | Windows | Linux |
|---|---|---|
| `.bat`, `.cmd` | `cmd.exe /d /c` | — |
| `.ps1` | `powershell.exe -NoProfile -ExecutionPolicy Bypass -File` (ou `pwsh.exe`) | `pwsh -NoProfile -File`, se instalado |
| `.vbs` | `cscript.exe //nologo` (VBScript em descontinuação) | — |
| `.sh` | `bash.exe` (Git Bash ou WSL) | `bash`, ou `sh` |
| `.py` | `py.exe`, `python.exe` ou `python3.exe` | `python3` ou `python` |

- Um formato marcado com — **não aparece no menu** daquele sistema.
- Qualquer outra extensão também não vira item (mas pode ser arquivo de
  apoio, veja abaixo).
- Se a máquina não tiver o interpretador (Python, `pwsh`...), o item aparece
  e o BigB avisa ao escolher.

### Mesmo script para os dois sistemas

O mesmo `scripts/` serve aos dois executáveis. Para ter a mesma função nos
dois, escreva uma versão de cada e marque com `@so`:

```
scripts/20-sistema/
  informacoes.ps1     # @so windows   (usa Get-CimInstance, que so existe no Windows)
  informacoes.sh      # @so linux
```

Com o mesmo `@nome`, o técnico vê um único "Informações do sistema" em cada
sistema. Sem `@so`, um `.ps1` também apareceria no Linux que tiver `pwsh`, e
um `.sh` no Windows que tiver Git Bash.

## Cabeçalho (metadados)

Opcional. Nas primeiras 50 linhas, em comentário do próprio formato:

```bat
@echo off
:: @nome      Reset de rede completo
:: @descricao winsock, pilha TCP/IP, renova o IP e limpa o DNS
:: @admin     sim
:: @confirmar sim
```

```powershell
# @nome      Informações do sistema
# @descricao SO, processador, memória e discos
```

| Chave | Efeito |
|---|---|
| `@nome` | Rótulo no menu (sem ele, vem do nome do arquivo). |
| `@descricao` (ou `@descrição`) | Dica ao lado do nome e no topo da execução. |
| `@admin` | Marca `(admin)` no menu; sem elevação, o BigB recusa rodar. |
| `@confirmar` | Pede `[s/N]` antes de rodar. Use em tudo que apaga ou reseta. |
| `@so` | Sistemas onde o script aparece: `windows`, `linux` ou `windows, linux`. Sem ele, aparece em todo sistema que roda o formato. |

Comentários aceitos: `::` e `REM` (batch), `#` (PowerShell, Python, bash),
`'` (VBScript). Valores de sim/não: `sim`, `s`, `true`, `yes`, `1` (ou a
chave sozinha); qualquer outra coisa é não.

## Arquivos de apoio

Itens cujo nome começa com `_` ou `.` **não aparecem no menu**, mas são
levados junto quando um script da mesma pasta roda:

```
scripts/30-impressoras/
  instalar-elgin.bat        ->  aparece no menu
  _drivers/                 ->  oculto; extraído junto
    elgin.inf
  leia-me.txt               ->  não é script: fora do menu, mas extraído junto
```

Na execução, o script roda numa pasta temporária que contém ele, os arquivos
da mesma pasta e as subpastas ocultas. Então, dentro do `instalar-elgin.bat`,
`pnputil /add-driver _drivers\elgin.inf` funciona. Subpastas visíveis (outras
categorias) não são copiadas.

## Como o script roda

- **Console herdado:** o script escreve direto na tela e pode perguntar
  coisas ao técnico (`set /p`, `Read-Host`, `input()`).
- **Pasta de trabalho:** `%LOCALAPPDATA%\bigb\execucao\<aleatória>` no
  Windows, `~/.cache/bigb/execucao/<aleatória>` no Linux; apagada no fim.
  Fica fora da pasta temporária de propósito, para que um script de limpeza
  de temporários não apague a si mesmo.
- **Código de saída:** `0` mostra "Concluído"; outro valor mostra um aviso
  com o código. Tudo vai para o log (`logs/toolbox_AAAAMMDD.log`).
- **Ctrl+C** interrompe o script, não o BigB.
- **Privilégio:** no Windows o BigB abre elevado (UAC) e os scripts herdam
  o admin; no Linux, rode `sudo ./bigb` para os scripts com `@admin`.

## Encoding: salve em UTF-8

Salve os scripts em UTF-8, que é o padrão dos editores. Na extração, o BigB
ajusta o arquivo para cada interpretador:

| Formato | Ajuste automático | Por quê |
|---|---|---|
| `.bat` / `.cmd` | fim de linha vira CRLF; com acento, ganha `@chcp 65001 >nul` na 1ª linha (se o script não tiver `chcp`) | com LF o `cmd` erra `goto`/`call :rotulo`; sem o code page UTF-8, os acentos saem trocados |
| `.ps1` | com acento e sem BOM, ganha BOM | o PowerShell 5.1 lê UTF-8 sem BOM como ANSI |
| `.vbs` | com acento, vira UTF-16LE com BOM | o `cscript` lê UTF-8 como ANSI |
| `.sh` | CRLF vira LF | o bash acusa `\r: command not found` |

Scripts salvos em ANSI (Windows-1252) ficam intactos: os interpretadores já
leem assim. No Windows, o code page do console é restaurado depois de cada
script.

## Testar sem recompilar

```bash
bigb.exe -sem-elevacao -scripts C:\caminho\da\pasta
./bigb -scripts ~/meus-scripts
```

Usa a pasta informada no lugar da embutida. Na hora de distribuir, copie os
scripts para `scripts/` e recompile: os executáveis levam tudo dentro deles.
