// Comando bigb: ponto de entrada do BigB Toolbox.
//
// O BigB e um lancador de scripts de suporte: o menu e montado a partir da
// pasta scripts/ (embutida no executavel), onde cada subpasta vira uma
// categoria e cada script vira uma opcao.
//
// No Windows o executavel sobe elevado sozinho (UAC); use -sem-elevacao para
// rodar como usuario padrao. No Linux, rode com sudo quando precisar de root.
// Sem privilegio, scripts marcados com @admin ficam bloqueados.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"bigb"
	"bigb/internal/app"
	"bigb/internal/config"
	"bigb/internal/scripts"
	"bigb/internal/winsys"
)

func main() {
	semElevacao := flag.Bool("sem-elevacao", false, "nao pedir elevacao (UAC); roda com o token atual")
	versao := flag.Bool("versao", false, "mostra a versao e sai")
	pasta := flag.String("scripts", "", "usa esta pasta de scripts no lugar da embutida (para testar sem recompilar)")
	flag.Parse()

	if *versao {
		fmt.Printf("%s v%s %s\n", config.Name, config.Version, config.Signature)
		return
	}

	fsys, err := abrirScripts(*pasta)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}

	// No Windows pedimos o UAC aqui e saimos: quem continua e o processo
	// elevado. No Linux a convencao e o proprio usuario rodar com sudo.
	if runtime.GOOS == "windows" && !*semElevacao && !winsys.IsAdmin() {
		var args []string
		if *pasta != "" {
			// O processo elevado pode nascer em outra pasta: vai caminho absoluto.
			abs, _ := filepath.Abs(*pasta)
			args = []string{"-scripts", abs}
		}
		if err := winsys.RelaunchElevated(args); err == nil {
			return
		}
		fmt.Println("Elevacao negada: seguindo como usuario padrao (scripts @admin ficam bloqueados).")
	}

	raiz, err := scripts.Carregar(fsys)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro lendo os scripts:", err)
		os.Exit(1)
	}

	app.Run(fsys, raiz)
	os.Exit(0)
}

// abrirScripts devolve a pasta de scripts: a externa, quando informada, ou
// a embutida no executavel.
func abrirScripts(pasta string) (fs.FS, error) {
	if pasta == "" {
		return fs.Sub(bigb.Scripts, "scripts")
	}
	info, err := os.Stat(pasta)
	if err != nil {
		return nil, fmt.Errorf("pasta de scripts: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("pasta de scripts: %s nao e uma pasta", pasta)
	}
	return os.DirFS(pasta), nil
}
