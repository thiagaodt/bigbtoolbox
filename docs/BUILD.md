# Build — gerar os executáveis

O BigB é um programa Go: um binário único por sistema, com a pasta
`scripts/` embutida. Na máquina de destino não é preciso instalar nada.

## Pré-requisito (só na máquina de desenvolvimento)

[Go 1.26 ou mais novo](https://go.dev/dl/). O Go compila para Linux a partir
do Windows (e vice-versa) sem ferramenta extra.

## Rodar do código-fonte

```bash
go run ./cmd/bigb -sem-elevacao
```

O `go run` recompila a cada execução, então scripts novos em `scripts/` já
aparecem. No Windows, sem `-sem-elevacao`, ele pede UAC e reabre elevado.

## Gerar os executáveis

```powershell
powershell -ExecutionPolicy Bypass -File .\build\build.ps1
```

O script:

1. confere o Go;
2. roda os testes nesta máquina (`-SemTestes` pula) e o `go vet` para
   Windows **e** Linux, o que pega erro de compilação no código só-Linux;
3. compila os dois alvos:

| Arquivo | Sistema |
|---|---|
| `dist\bigb.exe` | Windows amd64 |
| `dist\bigb` | Linux amd64 (estático, `CGO_ENABLED=0`: roda em qualquer distro) |

Use `-OutDir <pasta>` para mudar o destino.

Equivalente manual:

```bash
go build -trimpath -ldflags "-s -w" -o dist/bigb.exe ./cmd/bigb
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/bigb ./cmd/bigb
```

`-trimpath` tira os caminhos da máquina de build do binário; `-s -w` removem
símbolos e DWARF. O motor ocupa pouco mais de 2 MB; o resto vem dos scripts
embutidos.

**No Linux, dê permissão de execução** depois de copiar: `chmod +x bigb`. O
NTFS não guarda o bit de execução, então o arquivo gerado no Windows chega
sem ele.

**Scripts mudaram? Recompile.** O `go:embed` copia `scripts/` para dentro do
binário no momento do build. Para testar sem recompilar, use
`bigb -scripts <pasta>`.

### Elevação

- **Windows:** não há manifesto embutido. Ao abrir sem elevação, o
  executável se relança com o verbo `runas` (UAC) e sai. Se o UAC for
  negado, segue como usuário padrão.
- **Linux:** o BigB não pede `sudo` sozinho; rode `sudo ./bigb` quando
  precisar de root.

Sem privilégio, os scripts marcados com `@admin` ficam bloqueados.

### Ícone (opcional, Windows)

O Go não embute ícone sozinho. Gere um `.syso` com
[goversioninfo](https://github.com/josephspurrier/goversioninfo) e coloque-o
em `cmd/bigb/` antes do build; o linker o inclui automaticamente.

## Dependências na máquina de destino

| Recurso | Para quê |
|---|---|
| (nada) | o menu; `.bat`/`.cmd`/`.ps1`/`.vbs` no Windows; `.sh` no Linux |
| Python | scripts `.py` |
| Git Bash ou WSL (Windows) | scripts `.sh` no Windows |
| PowerShell 7 `pwsh` (Linux) | scripts `.ps1` no Linux |
