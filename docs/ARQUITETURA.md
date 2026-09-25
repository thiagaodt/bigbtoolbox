# Arquitetura — BigB Toolbox

O BigB é um lançador de scripts: um binário Go (um para Windows, outro para
Linux) que embute a pasta `scripts/`, monta um menu navegável pelas setas a
partir das pastas e executa o script escolhido com o interpretador certo.
Ele não tem funcionalidade própria.

## Camadas

```
+-----------------------------------------------------------+
|  cmd/bigb        flags, UAC (Windows), fonte dos scripts  |
+-----------------------------------------------------------+
              |
              v
+-----------------------------------------------------------+
|  App  (internal/app)                                      |
|  Menu por setas (ou numerado, sem terminal), navegacao,   |
|  confirmacao, admin, resultado da execucao.               |
+-----------------------------------------------------------+
              | usa
              v
+-----------------------------------------------------------+
|  Core                                                     |
|  scripts  descoberta, metadados, extracao, execucao       |
|  ui       tela, teclado cru, logo, itens de menu          |
|  config   nome/versao, pasta de logs, tema                |
|  logging  logs/toolbox_AAAAMMDD.log                       |
|  winsys   privilegio de administrador                     |
+-----------------------------------------------------------+
```

Regras:

- **`scripts` não desenha nada.** Ele devolve a árvore e o código de saída;
  quem mostra é o `app`. Por isso dá para testar o motor inteiro com um
  `fstest.MapFS`, sem console.
- **A fonte dos scripts é um `fs.FS`.** A embutida (`bigb.Scripts`) e a
  externa (`os.DirFS`, flag `-scripts`) passam pelo mesmo código.
- **Tudo que é destrutivo fica no script.** O BigB só garante a confirmação
  (`@confirmar`) e o bloqueio sem privilégio (`@admin`).
- **O que muda por sistema fica em arquivos `_windows.go` / `_linux.go`**
  (o Go escolhe pelo sufixo do nome). O resto do código é o mesmo.

## Windows x Linux

| Assunto | Windows | Linux |
|---|---|---|
| Formatos (`scripts/formatos_*.go`) | bat, cmd, ps1, vbs, py, sh | sh, py, ps1 (com `pwsh`) |
| Ajuste do console na execução | guarda e restaura o code page | nada |
| Privilégio (`winsys/admin_*.go`) | token elevado; relança com UAC (`runas`) | `euid == 0`; o usuário roda com `sudo` |
| Teclado (`ui/console_*.go`) | `ReadConsoleInputW` (códigos virtuais das teclas) | termios em modo cru + sequências ANSI (`ESC [ A`) |
| Cores, largura, título | `SetConsoleMode` (VT), `GetConsoleScreenBufferInfo` | ANSI nativo, `TIOCGWINSZ`, OSC 0 |
| Pasta de execução | `%LOCALAPPDATA%\bigb\execucao` | `~/.cache/bigb/execucao` |

Scripts com `@so` de outro sistema e formatos sem interpretador no sistema
não entram na árvore: o menu de cada executável só mostra o que ele pode
rodar.

## Fluxo

```
bigb(.exe)
   -> main (cmd/bigb)
        -> fonte: fs.Sub(bigb.Scripts, "scripts")  ou  os.DirFS(-scripts)
        -> Windows sem elevacao? winsys.RelaunchElevated (UAC) e sai
        -> scripts.Carregar(fonte)        arvore (filtra formato e @so)
        -> app.Run
             -> terminal? navegar (setas)  :  navegarNumerado (pipe)
             -> executar(script)
                  -> @admin sem privilegio? recusa
                  -> @confirmar? ui.Confirm
                  -> scripts.Executar
                       -> Interpretador (LookPath)
                       -> extrai script + irmaos + pastas "_"
                       -> normaliza encoding (CRLF, chcp, BOM, UTF-16, LF)
                       -> roda com o console herdado; devolve o codigo
```

## O menu

`navegar` redesenha a tela inteira a cada tecla dentro de `ui.Quadro`, que
junta toda a saída num buffer e manda ao terminal num único `Write`. Assim
não pisca. O cursor fica escondido enquanto o menu está na tela. Quando a
lista não cabe na janela, aparece só uma faixa em volta da seleção, com os
indicadores `^ mais N` / `v mais N`.

Se a entrada não é um terminal (pipe, automação, testes), o menu cai para o
modo numerado (`navegarNumerado`): digita-se o número do item e Enter.

## Por que o `go:embed` fica na raiz

A diretiva `//go:embed` só enxerga arquivos da pasta do próprio pacote para
baixo. Por isso existe o pacote raiz `bigb` ([scripts.go](../scripts.go)),
ao lado de `scripts/`, com `//go:embed all:scripts`. O prefixo `all:` inclui
os arquivos que começam com `_` (recursos). Se a pasta ficar sem nenhum
arquivo, a compilação falha; o `scripts/_LEIAME.md` evita isso.

## Pacotes

### `internal/scripts` — o motor

| Função/Tipo | Papel |
|---|---|
| `Carregar(fs)` | Percorre o FS e devolve a árvore `Categoria` (ordenada, sem categorias vazias, ocultos e `@so` de outro sistema de fora). |
| `Categoria`, `Script` | Árvore do menu; `Script` guarda nome, descrição, `Admin`, `Confirmar`, `Sistemas`, formato e caminho. |
| `NomeExibicao` | `10-rede-e-internet` → "Rede e internet". |
| `FormatoDe`, `FormatosSuportados` | Extensão → formato suportado neste sistema. |
| `Interpretador` | Localiza o programa que roda o formato (primeiro achado no PATH). |
| `Executar(fs, script)` | Extrai, normaliza, roda com o console herdado e devolve o código de saída. Ignora Ctrl+C no BigB. |

Para acrescentar uma extensão: uma linha na tabela `formatos` de
`formatos_windows.go` e/ou `formatos_linux.go` (interpretadores, argumentos,
função de normalização).

### `internal/app`

| Função | Papel |
|---|---|
| `Run` | Entrada do app: escolhe menu por setas ou numerado; tela de ajuda se não houver scripts. |
| `navegar` | Menu por setas de uma categoria (recursivo). Esc volta, Q/Ctrl+C sai de todos os níveis. |
| `navegarNumerado` | Mesmo menu para entrada sem terminal. |
| `executar` | Confere `@admin`/`@confirmar`, roda, mostra "Concluído" ou o código, loga. |

### `internal/ui`, `internal/config`, `internal/logging`, `internal/winsys`

- `ui`: `Quadro`, `Header`, `Logo` (arte em `ui/assets/logo.txt`), `Opcao`
  (item navegável), `MenuItem` (item numerado), `LerTecla`, `Interativo`,
  `Confirm`, `Pause`, `Ok/Warn/Err`.
- `config`: `Name`, `Version`, `Signature`, pasta de logs e tema de cores.
- `logging`: uma linha datada por evento em `logs/toolbox_AAAAMMDD.log`.
- `winsys`: `IsAdmin`, `RelaunchElevated(args)` (só Windows) e
  `DicaElevacao`.

## Estrutura de pastas

```
bigb/
  go.mod                 module bigb (unica dependencia: x/sys)
  scripts.go             pacote bigb: go:embed all:scripts
  scripts/               OS SCRIPTS (tudo o que o menu mostra)
  cmd/bigb/main.go       flags, UAC, carrega e roda
  internal/
    scripts/             motor; formatos_windows.go / formatos_linux.go
    app/                 menu e navegacao
    ui/                  tela e teclado; console_windows.go / console_linux.go
    winsys/              admin_windows.go / admin_linux.go
    config/  logging/
  build/build.ps1        testa, vet nos dois sistemas, gera bigb.exe e bigb
  docs/                  SCRIPTS.md, ARQUITETURA.md, BUILD.md
  _autosystem/           edicao AutoSystem privada (fora do build e do git)
```

## Testes

```bash
go test ./...
```

- `internal/scripts`, nos dois sistemas: nomes de exibição, metadados
  (inclusive `@so`), montagem da árvore (ordem, ocultos, categorias vazias,
  filtro por sistema) e normalização de encoding.
- Só no Windows: execução real de `.bat` (recursos extraídos, CRLF, código
  de saída) e `.ps1` (acento sem BOM lido corretamente).
- Só no Linux: execução real de `.sh` (recursos extraídos, CRLF → LF).
- `internal/ui`: tradução das sequências de teclas (setas, Home/End,
  PgUp/PgDn, Enter, Esc, Backspace, Ctrl+C).

Os testes do Linux podem ser compilados no Windows e rodados no WSL:
`GOOS=linux go test -c ./internal/scripts` gera um binário de teste.
