package scripts

// formatos sao as extensoes de script que o BigB reconhece no Linux.
//
// bat, cmd e vbs nao existem aqui e simplesmente nao aparecem no menu. ps1
// so roda com o PowerShell 7 (pwsh) instalado.
var formatos = map[string]formato{
	"sh":  {Rotulo: "sh", programas: []string{"bash", "sh"}, normalizar: normalizarShell},
	"py":  {Rotulo: "py", programas: []string{"python3", "python"}},
	"ps1": {Rotulo: "ps1", programas: []string{"pwsh"}, args: []string{"-NoProfile", "-File"}, normalizar: normalizarPowerShell},
}

// preservarConsole nao tem o que guardar no Linux (nao ha code page).
func preservarConsole() func() { return func() {} }
