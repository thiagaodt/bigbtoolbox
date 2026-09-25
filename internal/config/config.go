// Package config concentra a configuracao do BigB Toolbox.
//
// Ponto UNICO de configuracao do aplicativo: tudo que depende do ambiente de
// instalacao (pasta de logs, tema de cores) mora aqui. Ajustar o comportamento em
// outra estacao = editar apenas este arquivo.
//
// O objeto e criado uma vez (sync.Once) e memorizado; use sempre Get().
//
// Camada: Core / Infraestrutura. Nao depende de nenhuma outra parte do toolbox.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Name, Version e Signature identificam a build mostrada no cabecalho.
const (
	Name      = "BigB Toolbox"
	Version   = "2.0.0"
	Signature = "by fakedev25"
)

// Color e uma sequencia ANSI de cor. Os nomes espelham o ConsoleColor usado na
// versao PowerShell do toolbox (Green = verde claro, DarkGreen = verde escuro).
type Color string

// Paleta basica (ANSI). Reset e aplicado pela camada de UI.
const (
	ColorNone      Color = ""
	ColorGreen     Color = "\x1b[92m"
	ColorDarkGreen Color = "\x1b[32m"
	ColorWhite     Color = "\x1b[97m"
	ColorDarkGray  Color = "\x1b[90m"
	ColorYellow    Color = "\x1b[93m"
	ColorRed       Color = "\x1b[91m"
	ColorCyan      Color = "\x1b[96m"
)

// Theme agrupa as cores da identidade visual (verde + assinatura fakedev25).
type Theme struct {
	Logo     Color
	LogoText Color
	Accent   Color
	Credit   Color
	Muted    Color
	Ok       Color
	Warn     Color
	Err      Color
}

// Config e o objeto de configuracao completo do app.
type Config struct {
	Version   string
	Signature string

	AppRoot string
	LogDir  string

	Theme Theme
}

var (
	once   sync.Once
	cached *Config
)

// appRoot descobre a pasta do app - onde fica logs\.
//
// Normalmente e a pasta do executavel. A excecao e o "go run", que compila
// num diretorio temporario: ali o binario nao tem logs\ ao lado,
// entao caimos para o diretorio de trabalho (a raiz do projeto), que e o que
// o desenvolvedor espera.
func appRoot() string {
	diretorioAtual := func() string {
		if wd, err := os.Getwd(); err == nil {
			return wd
		}
		return "."
	}

	exe, err := os.Executable()
	if err != nil {
		return diretorioAtual()
	}
	if resolvido, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolvido
	}
	dir := filepath.Dir(exe)

	if temp := os.TempDir(); temp != "" {
		if rel, err := filepath.Rel(temp, dir); err == nil && !strings.HasPrefix(rel, "..") {
			return diretorioAtual()
		}
	}
	return dir
}

// New constroi a configuracao a partir do layout de pastas do app.
func New() *Config {
	root := appRoot()
	return &Config{
		Version:   Version,
		Signature: Signature,

		AppRoot: root,
		LogDir:  filepath.Join(root, "logs"),

		Theme: Theme{
			Logo:     ColorCyan,
			LogoText: ColorWhite,
			Accent:   ColorCyan,
			Credit:   ColorCyan,
			Muted:    ColorDarkGray,
			Ok:       ColorCyan,
			Warn:     ColorYellow,
			Err:      ColorRed,
		},
	}
}

// Get retorna a configuracao do app, criando-a na primeira chamada.
func Get() *Config {
	once.Do(func() { cached = New() })
	return cached
}

// GetTheme e um atalho para o tema de cores.
func GetTheme() Theme { return Get().Theme }
