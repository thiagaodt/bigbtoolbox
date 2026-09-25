// Console: detalhes de baixo nivel do terminal no Linux (cores, largura,
// titulo e leitura de tecla), espelhando console_windows.go.
package ui

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// terminal informa se o descritor e um terminal (e nao pipe/arquivo).
func terminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), unix.TCGETS)
	return err == nil
}

// EnableVirtualTerminal informa se da para usar cores: terminais Linux
// entendem ANSI nativamente, so a saida redirecionada fica sem cor.
func EnableVirtualTerminal() bool { return terminal(os.Stdout.Fd()) }

// SetTitle define o titulo da janela do terminal (sequencia OSC 0).
func SetTitle(titulo string) {
	if terminal(os.Stdout.Fd()) {
		fmt.Printf("\x1b]0;%s\x07", titulo)
	}
}

// Width devolve a largura do terminal em colunas (fallback 120).
func Width() int {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 {
		return 120
	}
	return int(ws.Col)
}

// Resize nao faz nada no Linux: redimensionar a janela do terminal de fora
// e intrusivo e poucos emuladores aceitam.
func Resize(colunas, linhas, bufferLinhas int) {}

// Height devolve a altura do terminal em linhas (fallback 30).
func Height() int {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Row == 0 {
		return 30
	}
	return int(ws.Row)
}

// Interativo informa se entrada e saida sao o terminal (e nao pipe/arquivo):
// so assim da para navegar com as setas.
func Interativo() bool { return terminal(os.Stdin.Fd()) && terminal(os.Stdout.Fd()) }

// LerTecla espera uma tecla e devolve qual foi. Poe o terminal em modo cru
// (sem eco, sem esperar Enter, sem gerar sinal no Ctrl+C) so durante a
// leitura.
func LerTecla() (Tecla, rune) {
	fd := int(os.Stdin.Fd())
	original, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		if ReadLine(); entradaFim {
			return TeclaFimEntrada, 0
		}
		return TeclaEnter, 0
	}
	cru := *original
	cru.Lflag &^= unix.ICANON | unix.ECHO | unix.ISIG
	cru.Cc[unix.VMIN], cru.Cc[unix.VTIME] = 1, 0
	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &cru); err != nil {
		return TeclaFimEntrada, 0
	}
	defer unix.IoctlSetTermios(fd, unix.TCSETS, original)

	buf := make([]byte, 32)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return TeclaFimEntrada, 0
	}
	return interpretarTecla(buf[:n])
}
