package scripts

import "golang.org/x/sys/windows"

// formatos sao as extensoes de script que o BigB reconhece no Windows.
//
// bat/cmd e ps1 rodam em qualquer Windows; vbs usa o cscript (a Microsoft
// esta descontinuando o VBScript); py e sh dependem de Python e bash
// (Git Bash ou WSL) instalados na maquina.
var formatos = map[string]formato{
	"bat": {Rotulo: "bat", programas: []string{"cmd.exe"}, args: []string{"/d", "/c"}, normalizar: normalizarBatch},
	"cmd": {Rotulo: "cmd", programas: []string{"cmd.exe"}, args: []string{"/d", "/c"}, normalizar: normalizarBatch},
	"ps1": {Rotulo: "ps1", programas: []string{"powershell.exe", "pwsh.exe"}, args: []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File"}, normalizar: normalizarPowerShell},
	"vbs": {Rotulo: "vbs", programas: []string{"cscript.exe"}, args: []string{"//nologo"}, normalizar: normalizarVBScript},
	"py":  {Rotulo: "py", programas: []string{"py.exe", "python.exe", "python3.exe"}},
	"sh":  {Rotulo: "sh", programas: []string{"bash.exe"}, normalizar: normalizarShell},
}

// preservarConsole guarda o code page do console e devolve a funcao que o
// restaura: scripts podem trocar o code page (chcp) e isso nao deve afetar
// o que vier depois.
func preservarConsole() func() {
	entrada, _ := windows.GetConsoleCP()
	saida, _ := windows.GetConsoleOutputCP()
	return func() {
		if entrada != 0 {
			windows.SetConsoleCP(entrada)
		}
		if saida != 0 {
			windows.SetConsoleOutputCP(saida)
		}
	}
}
