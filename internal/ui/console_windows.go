// Console: detalhes de baixo nivel do terminal do Windows (VT, tamanho da
// janela, titulo e leitura de tecla). Isolado aqui para o resto da UI ser
// apenas "desenho".
package ui

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                      = windows.NewLazySystemDLL("kernel32.dll")
	procSetConsoleTitleW          = kernel32.NewProc("SetConsoleTitleW")
	procSetConsoleScreenBufferSiz = kernel32.NewProc("SetConsoleScreenBufferSize")
	procSetConsoleWindowInfo      = kernel32.NewProc("SetConsoleWindowInfo")
	procGetLargestConsoleWindowSz = kernel32.NewProc("GetLargestConsoleWindowSize")
	procReadConsoleInputW         = kernel32.NewProc("ReadConsoleInputW")
)

// EnableVirtualTerminal liga o processamento de sequencias ANSI no console,
// que e como este app pinta as cores. Em terminais que ja suportam (Windows
// Terminal) e inofensivo; em consoles antigos sem suporte, as cores saem
// desligadas em vez de sujar a tela.
func EnableVirtualTerminal() bool {
	h := windows.Handle(os.Stdout.Fd())
	var modo uint32
	if err := windows.GetConsoleMode(h, &modo); err != nil {
		return false // saida redirecionada (pipe/arquivo)
	}
	if modo&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}
	if err := windows.SetConsoleMode(h, modo|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return false
	}
	return true
}

// SetTitle define o titulo da janela do console.
func SetTitle(titulo string) {
	p, err := syscall.UTF16PtrFromString(titulo)
	if err != nil {
		return
	}
	procSetConsoleTitleW.Call(uintptr(unsafe.Pointer(p)))
}

// Height devolve a altura visivel do console em linhas (fallback 30).
func Height() int {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stdout.Fd()), &info); err != nil {
		return 30
	}
	if h := int(info.Window.Bottom-info.Window.Top) + 1; h > 0 {
		return h
	}
	return 30
}

// Width devolve a largura util do console em colunas (fallback 120).
func Width() int {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stdout.Fd()), &info); err != nil {
		return 120
	}
	w := int(info.Window.Right-info.Window.Left) + 1
	if w <= 0 {
		return 120
	}
	return w
}

// Resize amplia a janela e o buffer do console ate o minimo pedido, quando o
// console permite. Em Windows Terminal/ISE a chamada e ignorada sem erro.
func Resize(colunas, linhas, bufferLinhas int) {
	h := windows.Handle(os.Stdout.Fd())
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(h, &info); err != nil {
		return
	}
	if int(info.Size.X) >= colunas {
		return
	}

	// O buffer precisa crescer antes da janela; a janela nunca pode exceder o
	// maior tamanho fisico suportado pela fonte/monitor atuais.
	novoBuf := windows.Coord{X: int16(colunas), Y: int16(bufferLinhas)}
	procSetConsoleScreenBufferSiz.Call(uintptr(h), uintptr(*(*uint32)(unsafe.Pointer(&novoBuf))))

	maxRet, _, _ := procGetLargestConsoleWindowSz.Call(uintptr(h))
	maxCoord := *(*windows.Coord)(unsafe.Pointer(&maxRet))
	if maxCoord.X > 0 && int(maxCoord.X) < colunas {
		colunas = int(maxCoord.X)
	}
	if maxCoord.Y > 0 && int(maxCoord.Y) < linhas {
		linhas = int(maxCoord.Y)
	}
	janela := windows.SmallRect{Left: 0, Top: 0, Right: int16(colunas - 1), Bottom: int16(linhas - 1)}
	procSetConsoleWindowInfo.Call(uintptr(h), 1, uintptr(unsafe.Pointer(&janela)))
}

// Interativo informa se entrada e saida sao o console (e nao pipe/arquivo):
// so assim da para navegar com as setas.
func Interativo() bool {
	var modo uint32
	return windows.GetConsoleMode(windows.Handle(os.Stdin.Fd()), &modo) == nil &&
		windows.GetConsoleMode(windows.Handle(os.Stdout.Fd()), &modo) == nil
}

// Registros de entrada do console (INPUT_RECORD / KEY_EVENT_RECORD).
const eventoTecla = 0x0001

type registroEntrada struct {
	tipo  uint16
	_     uint16
	tecla registroTecla
}

type registroTecla struct {
	pressionada   int32
	repeticoes    uint16
	codigoVirtual uint16
	codigoScan    uint16
	caractere     uint16
	estadoCtrl    uint32
}

// Codigos virtuais das teclas de navegacao.
var teclasVirtuais = map[uint16]Tecla{
	0x26: TeclaCima, 0x28: TeclaBaixo, 0x25: TeclaEsquerda, 0x27: TeclaDireita,
	0x0D: TeclaEnter, 0x1B: TeclaEsc, 0x08: TeclaApagar,
	0x24: TeclaInicio, 0x23: TeclaFim, 0x21: TeclaPaginaCima, 0x22: TeclaPaginaBaixo,
}

// LerTecla espera uma tecla e devolve qual foi. Le os eventos crus do
// console (ReadConsoleInputW), o que funciona igual no conhost e no Windows
// Terminal. Ctrl+C vira TeclaInterromper em vez de derrubar o processo.
func LerTecla() (Tecla, rune) {
	h := windows.Handle(os.Stdin.Fd())
	var modo uint32
	if err := windows.GetConsoleMode(h, &modo); err != nil {
		if ReadLine(); entradaFim {
			return TeclaFimEntrada, 0
		}
		return TeclaEnter, 0
	}
	windows.SetConsoleMode(h, modo&^windows.ENABLE_PROCESSED_INPUT)
	defer windows.SetConsoleMode(h, modo)

	for {
		var reg registroEntrada
		var lidos uint32
		r, _, _ := procReadConsoleInputW.Call(uintptr(h), uintptr(unsafe.Pointer(&reg)), 1, uintptr(unsafe.Pointer(&lidos)))
		if r == 0 {
			return TeclaFimEntrada, 0
		}
		if lidos == 0 || reg.tipo != eventoTecla || reg.tecla.pressionada == 0 {
			continue // soltar de tecla, mouse, redimensionamento...
		}
		if t, ok := teclasVirtuais[reg.tecla.codigoVirtual]; ok {
			return t, 0
		}
		switch c := rune(reg.tecla.caractere); {
		case c == 0x03:
			return TeclaInterromper, 0
		case c != 0:
			return TeclaOutra, c
		}
		// Shift, Ctrl, Alt sozinhos: sem caractere, espera a proxima.
	}
}
