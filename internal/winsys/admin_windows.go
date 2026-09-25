// Package winsys concentra o acesso ao Windows que o BigB precisa:
// elevacao (UAC). Nada aqui escreve na tela.
//
// Camada: Core / Infraestrutura.
package winsys

import (
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// IsAdmin informa se o processo atual esta elevado (Administrador).
func IsAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

// RelaunchElevated reabre o proprio executavel pedindo elevacao (UAC), com
// os argumentos informados.
//
// Devolve erro quando o usuario nega o UAC ou a chamada falha; nesse caso o
// chamador segue sem elevacao.
func RelaunchElevated(argumentos []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cwd, _ := os.Getwd()
	args := strings.Join(quoteArgs(argumentos), " ")

	verbo, _ := syscall.UTF16PtrFromString("runas")
	arquivo, _ := syscall.UTF16PtrFromString(exe)
	diretorio, _ := syscall.UTF16PtrFromString(cwd)
	var parametros *uint16
	if args != "" {
		parametros, _ = syscall.UTF16PtrFromString(args)
	}

	const swShowNormal = 1
	return windows.ShellExecute(0, verbo, arquivo, parametros, diretorio, swShowNormal)
}

func quoteArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if strings.ContainsAny(a, " \t\"") {
			a = `"` + strings.ReplaceAll(a, `"`, `\"`) + `"`
		}
		out = append(out, a)
	}
	return out
}

// DicaElevacao explica como abrir o BigB com privilegio de administrador.
func DicaElevacao() string {
	return "Abra o BigB elevado (sem -sem-elevacao) e aceite o UAC."
}
