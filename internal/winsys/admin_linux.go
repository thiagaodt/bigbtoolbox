// Package winsys concentra o acesso ao sistema operacional que o BigB
// precisa: privilegio de administrador. Nada aqui escreve na tela.
package winsys

import (
	"errors"
	"os"
)

// IsAdmin informa se o processo roda como root.
func IsAdmin() bool { return os.Geteuid() == 0 }

// RelaunchElevated nao existe no Linux: la a convencao e o proprio usuario
// chamar "sudo bigb" quando precisa de root.
func RelaunchElevated(argumentos []string) error { return errors.ErrUnsupported }

// DicaElevacao explica como abrir o BigB com privilegio de administrador.
func DicaElevacao() string {
	return "Rode o BigB com sudo: sudo ./bigb"
}
