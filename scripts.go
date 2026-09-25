// Package bigb guarda a pasta scripts/, embutida no executavel.
//
// A diretiva go:embed so enxerga arquivos da pasta do proprio pacote para
// baixo; por isso ela mora aqui na raiz, ao lado de scripts/.
package bigb

import "embed"

// Scripts e a pasta scripts/ inteira, embutida em tempo de compilacao. O
// prefixo all: inclui tambem os arquivos que comecam com "_" (recursos).
//
//go:embed all:scripts
var Scripts embed.FS
