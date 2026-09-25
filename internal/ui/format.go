package ui

import (
	"strings"
	"unicode/utf8"
)

// padRight completa o texto com o caractere de preenchimento ate a largura.
func padRight(s string, largura int, preenchimento rune) string {
	falta := largura - utf8.RuneCountInString(s)
	if falta <= 0 {
		return s
	}
	return s + strings.Repeat(string(preenchimento), falta)
}

// padLeft desloca o texto para a direita ate a largura pedida.
func padLeft(s string, largura int) string {
	falta := largura - utf8.RuneCountInString(s)
	if falta <= 0 {
		return s
	}
	return strings.Repeat(" ", falta) + s
}
