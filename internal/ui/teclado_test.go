package ui

import "testing"

func TestInterpretarTecla(t *testing.T) {
	casos := []struct {
		entrada string
		tecla   Tecla
		r       rune
	}{
		{"\x1b[A", TeclaCima, 0},
		{"\x1b[B", TeclaBaixo, 0},
		{"\x1b[C", TeclaDireita, 0},
		{"\x1b[D", TeclaEsquerda, 0},
		{"\x1bOA", TeclaCima, 0}, // modo de aplicacao do terminal
		{"\x1b[H", TeclaInicio, 0},
		{"\x1b[F", TeclaFim, 0},
		{"\x1b[1~", TeclaInicio, 0},
		{"\x1b[4~", TeclaFim, 0},
		{"\x1b[5~", TeclaPaginaCima, 0},
		{"\x1b[6~", TeclaPaginaBaixo, 0},
		{"\r", TeclaEnter, 0},
		{"\n", TeclaEnter, 0},
		{"\x1b", TeclaEsc, 0},
		{"\x7f", TeclaApagar, 0},
		{"\x03", TeclaInterromper, 0},
		{"q", TeclaOutra, 'q'},
		{"é", TeclaOutra, 'é'},
		{"\x1b[Z", TeclaOutra, 0}, // sequencia desconhecida nao vira caractere
		{"", TeclaFimEntrada, 0},
	}
	for _, c := range casos {
		tecla, r := interpretarTecla([]byte(c.entrada))
		if tecla != c.tecla || r != c.r {
			t.Errorf("interpretarTecla(%q) = (%v, %q), esperado (%v, %q)", c.entrada, tecla, r, c.tecla, c.r)
		}
	}
}
