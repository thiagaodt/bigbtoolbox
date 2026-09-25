package ui

import "unicode/utf8"

// Tecla e uma tecla de navegacao lida em modo cru (sem Enter).
type Tecla int

// Teclas reconhecidas. TeclaOutra vem acompanhada do caractere digitado.
const (
	TeclaOutra Tecla = iota
	TeclaCima
	TeclaBaixo
	TeclaEsquerda
	TeclaDireita
	TeclaEnter
	TeclaEsc
	TeclaApagar // Backspace
	TeclaInicio // Home
	TeclaFim    // End
	TeclaPaginaCima
	TeclaPaginaBaixo
	TeclaInterromper // Ctrl+C
	TeclaFimEntrada  // stdin fechado
)

// interpretarTecla traduz os bytes de uma tecla lida em modo cru num
// terminal VT (Linux) para uma Tecla. Setas chegam como sequencias de
// escape: ESC [ A (ou ESC O A, no modo de aplicacao).
func interpretarTecla(b []byte) (Tecla, rune) {
	if len(b) == 0 {
		return TeclaFimEntrada, 0
	}
	if len(b) == 1 {
		switch b[0] {
		case '\r', '\n':
			return TeclaEnter, 0
		case 0x1b:
			return TeclaEsc, 0
		case 0x7f, 0x08:
			return TeclaApagar, 0
		case 0x03:
			return TeclaInterromper, 0
		}
	}
	if b[0] == 0x1b && len(b) >= 3 && (b[1] == '[' || b[1] == 'O') {
		switch b[2] {
		case 'A':
			return TeclaCima, 0
		case 'B':
			return TeclaBaixo, 0
		case 'C':
			return TeclaDireita, 0
		case 'D':
			return TeclaEsquerda, 0
		case 'H':
			return TeclaInicio, 0
		case 'F':
			return TeclaFim, 0
		}
		// ESC [ n ~  (Home, End, PgUp, PgDn em varios terminais)
		if len(b) >= 4 && b[len(b)-1] == '~' {
			switch b[2] {
			case '1', '7':
				return TeclaInicio, 0
			case '4', '8':
				return TeclaFim, 0
			case '5':
				return TeclaPaginaCima, 0
			case '6':
				return TeclaPaginaBaixo, 0
			}
		}
		return TeclaOutra, 0
	}
	r, _ := utf8.DecodeRune(b)
	return TeclaOutra, r
}
