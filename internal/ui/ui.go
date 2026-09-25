// Package ui reune as primitivas de interface no terminal (tema, logo,
// cabecalho, itens de menu, teclado).
//
// Toda a camada visual do BigB. Nao contem regra de negocio: apenas
// desenha. As cores vem do tema em config.
//
// Tudo o que e desenhado passa pela variavel out; Quadro troca out por um
// buffer para mandar a tela inteira ao terminal de uma vez (sem piscar ao
// navegar com as setas).
//
// Camada: Core / Apresentacao.
// Depende de: config (tema), logging (Status alimenta o log), winsys (admin).
package ui

import (
	"bufio"
	"bytes"
	"cmp"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"bigb/internal/config"
	"bigb/internal/logging"
	"bigb/internal/winsys"
)

// Arte do logo (mantida fora do codigo para nao brigar com escapes).
//
//go:embed assets/logo.txt
var logo string

var (
	out          io.Writer = os.Stdout
	coresLigadas bool
	entrada      = bufio.NewReader(os.Stdin)
	entradaFim   bool // stdin chegou ao EOF: nao adianta pedir mais nada
)

// Init prepara o terminal (cores, titulo e tamanho da janela).
func Init() {
	coresLigadas = EnableVirtualTerminal()
	cfg := config.Get()
	host, _ := os.Hostname()
	SetTitle(fmt.Sprintf("%s v%s %s - %s", config.Name, cfg.Version, cfg.Signature, host))
	Resize(132, 42, 3000)
}

// Quadro desenha uma tela inteira de uma vez: o que desenhar escrever vai
// para um buffer, enviado ao terminal num unico Write.
func Quadro(desenhar func()) {
	var buf bytes.Buffer
	out = &buf
	defer func() {
		out = os.Stdout
		os.Stdout.Write(buf.Bytes())
	}()
	desenhar()
}

// CursorVisivel mostra ou esconde o cursor do terminal.
func CursorVisivel(visivel bool) {
	if !coresLigadas {
		return
	}
	if visivel {
		fmt.Fprint(out, "\x1b[?25h")
	} else {
		fmt.Fprint(out, "\x1b[?25l")
	}
}

// ------------------------------------------------------------------ escrita

func paint(cor config.Color, texto string) string {
	if !coresLigadas || cor == config.ColorNone {
		return texto
	}
	return string(cor) + texto + "\x1b[0m"
}

// Write imprime texto colorido sem quebrar linha.
func Write(cor config.Color, format string, args ...any) {
	fmt.Fprint(out, paint(cor, sprintf(format, args...)))
}

// Line imprime texto colorido com quebra de linha.
func Line(cor config.Color, format string, args ...any) {
	fmt.Fprintln(out, paint(cor, sprintf(format, args...)))
}

// Blank imprime uma linha em branco.
func Blank() { fmt.Fprintln(out) }

// Plain imprime texto sem cor, com quebra de linha.
func Plain(format string, args ...any) { fmt.Fprintln(out, sprintf(format, args...)) }

func sprintf(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

// Clear limpa a tela (inclusive o historico de rolagem).
func Clear() {
	if coresLigadas {
		fmt.Fprint(out, "\x1b[H\x1b[2J\x1b[3J")
		return
	}
	fmt.Fprint(out, strings.Repeat("\n", 50))
}

// ------------------------------------------------------------------ logo

// Logo desenha o logo com a assinatura alinhada a direita dele.
func Logo() {
	tema := config.GetTheme()
	cfg := config.Get()
	Blank()

	largura := 0
	for l := range strings.SplitSeq(strings.TrimRight(logo, "\r\n"), "\n") {
		l = strings.TrimRight(l, "\r")
		largura = max(largura, len([]rune(l)))
		Line(tema.Logo, "  %s", l)
	}
	Line(tema.Credit, "  %s", padLeft(cfg.Signature, largura))
	Blank()
}

// LinhasCabecalho informa quantas linhas o Header ocupa (para o menu saber
// quantos itens cabem na tela).
func LinhasCabecalho() int {
	return strings.Count(strings.TrimRight(logo, "\r\n"), "\n") + 1 + 7
}

// Header limpa a tela e desenha o cabecalho (logo + identificacao).
func Header() {
	tema := config.GetTheme()
	cfg := config.Get()
	Clear()
	Logo()

	admin, corAdmin := "[USUARIO PADRAO]", tema.Warn
	if winsys.IsAdmin() {
		admin, corAdmin = "[ADMIN]", tema.Ok
	}
	host, _ := os.Hostname()
	usuario := cmp.Or(os.Getenv("USERNAME"), os.Getenv("USER"))

	Write(tema.Accent, "  %s v%s   ", config.Name, cfg.Version)
	Line(corAdmin, "%s", admin)
	Line(tema.Muted, "  %s \\ %s   |   %s", host, usuario, time.Now().Format("02/01/2006 15:04"))
	Blank()
	Rule("")
}

// Rule desenha uma linha separadora, opcionalmente com titulo a esquerda.
func Rule(titulo string) {
	tema := config.GetTheme()
	w := min(max(Width()-1, 10), 120)
	if titulo == "" {
		Line(tema.Muted, "%s", strings.Repeat("-", w))
		return
	}
	esq := "-- " + titulo + " "
	resto := max(w-len([]rune(esq)), 0)
	Line(tema.Muted, "%s%s", esq, strings.Repeat("-", resto))
}

// ------------------------------------------------------------------ status

// Level e a severidade visual de uma linha de status.
type Level int

// Niveis de Status.
const (
	LevelInfo Level = iota
	LevelOk
	LevelWarn
	LevelErr
)

// Status escreve uma linha de status com etiqueta colorida e alimenta o log.
func Status(nivel Level, format string, args ...any) {
	tema := config.GetTheme()
	msg := sprintf(format, args...)

	var etiqueta string
	var cor config.Color
	var nivelLog logging.Level
	switch nivel {
	case LevelOk:
		etiqueta, cor, nivelLog = "[ OK ]", tema.Ok, logging.Info
	case LevelWarn:
		etiqueta, cor, nivelLog = "[ !! ]", tema.Warn, logging.Warn
	case LevelErr:
		etiqueta, cor, nivelLog = "[ XX ]", tema.Err, logging.Error
	default:
		etiqueta, cor, nivelLog = "[ .. ]", tema.Accent, logging.Info
	}
	Write(cor, "  %s ", etiqueta)
	fmt.Fprintln(out, msg)
	logging.Write(nivelLog, "%s", msg)
}

// Ok, Warn, Err e Info sao atalhos de Status.
func Ok(format string, args ...any)   { Status(LevelOk, format, args...) }
func Warn(format string, args ...any) { Status(LevelWarn, format, args...) }
func Err(format string, args ...any)  { Status(LevelErr, format, args...) }
func Info(format string, args ...any) { Status(LevelInfo, format, args...) }

// ------------------------------------------------------------------ menu

// larguraRotulo e a coluna onde comeca a dica dos itens de menu.
const larguraRotulo = 34

// MenuItem imprime um item de menu numerado: [tecla] Rotulo (admin) dica.
// Usado quando a entrada nao e um terminal (pipe, automacao).
func MenuItem(tecla, rotulo, dica string, exigeAdmin bool) {
	tema := config.GetTheme()
	Write(tema.Accent, "   [%s] ", tecla)
	fmt.Fprint(out, padRight(rotulo, larguraRotulo, ' '))
	fimDoItem(dica, exigeAdmin)
}

// Opcao imprime um item de menu navegavel pelas setas; o selecionado fica
// destacado (video reverso) com um marcador a esquerda.
func Opcao(rotulo, dica string, exigeAdmin, selecionado bool) {
	tema := config.GetTheme()
	if selecionado {
		Write(tema.Accent, "  > ")
		if coresLigadas {
			fmt.Fprint(out, "\x1b[7m"+padRight(" "+rotulo, larguraRotulo, ' ')+"\x1b[0m ")
		} else {
			fmt.Fprint(out, padRight(" "+rotulo, larguraRotulo, ' ')+" ")
		}
	} else {
		fmt.Fprint(out, "    "+padRight(" "+rotulo, larguraRotulo, ' ')+" ")
	}
	fimDoItem(dica, exigeAdmin)
}

func fimDoItem(dica string, exigeAdmin bool) {
	tema := config.GetTheme()
	if exigeAdmin {
		Write(tema.Warn, "(admin) ")
	}
	if dica != "" {
		Line(tema.Muted, "%s", dica)
	} else {
		Blank()
	}
}

// ------------------------------------------------------------------ entrada

// ReadLine le uma linha do teclado.
//
// Quando a entrada acaba (EOF - console fechado, automacao alimentando o
// BigB por pipe), marca o fim da entrada: sem isso o loop do menu ficaria
// girando para sempre lendo string vazia. Ver EntradaEncerrada.
func ReadLine() string {
	linha, err := entrada.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			entradaFim = true
		}
		if linha == "" {
			return ""
		}
	}
	return strings.TrimRight(linha, "\r\n")
}

// EntradaEncerrada informa se o teclado/stdin chegou ao fim.
func EntradaEncerrada() bool { return entradaFim }

// Ask escreve a pergunta e le a resposta (uma linha).
func Ask(cor config.Color, format string, args ...any) string {
	Write(cor, format, args...)
	return ReadLine()
}

// Confirm pede confirmacao [s/N] para acoes destrutivas.
func Confirm(mensagem string) bool {
	tema := config.GetTheme()
	Blank()
	Write(tema.Warn, "  %s ", mensagem)
	Write(tema.Accent, "[s/N]: ")
	r := strings.ToLower(strings.TrimSpace(ReadLine()))
	return r == "s" || r == "y"
}

// Pause espera uma tecla antes de voltar ao menu.
func Pause() {
	tema := config.GetTheme()
	Blank()
	Line(tema.Muted, "  Pressione qualquer tecla para voltar ao menu...")
	if entradaFim {
		return
	}
	if Interativo() {
		LerTecla()
		return
	}
	ReadLine()
}

// Sleep e um atalho para as pequenas pausas de leitura das mensagens.
func Sleep(ms int) { time.Sleep(time.Duration(ms) * time.Millisecond) }
