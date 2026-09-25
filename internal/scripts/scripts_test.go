package scripts

import (
	"bytes"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

func TestNomeExibicao(t *testing.T) {
	casos := map[string]string{
		"10-rede":            "Rede",
		"20_sistema":         "Sistema",
		"reset-de-rede":      "Reset de rede",
		"01. limpeza":        "Limpeza",
		"impressão":          "Impressão",
		"Rede e Internet":    "Rede e Internet",
		"123":                "123",
		"99-":                "(sem nome)",
		"flush__dns--rapido": "Flush dns rapido",
	}
	for entrada, esperado := range casos {
		if got := NomeExibicao(entrada); got != esperado {
			t.Errorf("NomeExibicao(%q) = %q, esperado %q", entrada, got, esperado)
		}
	}
}

func TestLerMetadados(t *testing.T) {
	casos := []struct {
		nome     string
		conteudo string
		esperado Script
	}{
		{"batch com ::", "@echo off\r\n:: @nome Limpar DNS\r\n:: @descricao ipconfig /flushdns\r\n:: @admin sim\r\n",
			Script{Nome: "Limpar DNS", Descricao: "ipconfig /flushdns", Admin: true}},
		{"batch com REM", "@echo off\r\nREM @nome: Teste\r\nrem @confirmar = s\r\n",
			Script{Nome: "Teste", Confirmar: true}},
		{"powershell com BOM e acento", bomUTF8 + "# @nome Informações\n# @descrição Memória e discos\n# @admin nao\n",
			Script{Nome: "Informações", Descricao: "Memória e discos"}},
		{"vbscript", "' @nome Mensagem\n' @confirmar\n", Script{Nome: "Mensagem", Confirmar: true}},
		{"so com lista", "# @so Windows, linux\n", Script{Sistemas: []string{"windows", "linux"}}},
		{"sem cabecalho", "@echo off\r\necho oi\r\n", Script{}},
		{"@echo nao e metadado", "@echo off\r\n@admin_nao_existe\r\n", Script{}},
	}
	for _, c := range casos {
		if got := lerMetadados([]byte(c.conteudo)); !reflect.DeepEqual(got, c.esperado) {
			t.Errorf("%s: lerMetadados = %+v, esperado %+v", c.nome, got, c.esperado)
		}
	}
}

func TestCarregar(t *testing.T) {
	// So formatos que existem nos dois sistemas (sh, py, ps1), para o teste
	// valer no Windows e no Linux.
	outroSO := "linux"
	if runtime.GOOS == "linux" {
		outroSO = "windows"
	}
	fsys := fstest.MapFS{
		"_LEIAME.md":                     {Data: []byte("oculto")},
		"raiz.ps1":                       {Data: []byte("Write-Host oi")},
		"20-sistema/info.ps1":            {Data: []byte("# @nome Info\n")},
		"20-sistema/so-do-outro.sh":      {Data: []byte("# @so " + outroSO + "\n")},
		"20-sistema/so-deste.sh":         {Data: []byte("# @so " + runtime.GOOS + "\n")},
		"10-rede/b-reset.sh":             {Data: []byte("# @admin sim\n")},
		"10-rede/a-flush.py":             {Data: []byte("")},
		"10-rede/instalador.exe":         {Data: []byte("MZ")}, // recurso: fora do menu
		"10-rede/_drivers/x.inf":         {Data: []byte("")},
		"10-rede/_drivers/instalar.sh":   {Data: []byte("")}, // dentro de pasta oculta
		"10-rede/wifi/perfis.ps1":        {Data: []byte("")},
		"30-vazia/leia.txt":              {Data: []byte("")}, // sem script: categoria some
		"40-so-subpasta/vazia/nada.docx": {Data: []byte("")},
		".git/config":                    {Data: []byte("")},
	}
	raiz, err := Carregar(fsys)
	if err != nil {
		t.Fatal(err)
	}

	if raiz.Total() != 6 {
		t.Errorf("Total = %d, esperado 6", raiz.Total())
	}
	if len(raiz.Scripts) != 1 || raiz.Scripts[0].Nome != "Raiz" {
		t.Errorf("scripts da raiz = %+v", raiz.Scripts)
	}

	var nomes []string
	for _, c := range raiz.Subcategorias {
		nomes = append(nomes, c.Nome)
	}
	if got := strings.Join(nomes, ","); got != "Rede,Sistema" {
		t.Fatalf("categorias = %s, esperado Rede,Sistema (ordenadas, sem as vazias)", got)
	}

	rede := raiz.Subcategorias[0]
	if len(rede.Scripts) != 2 || rede.Scripts[0].Nome != "A flush" || rede.Scripts[1].Nome != "B reset" {
		t.Fatalf("scripts de Rede fora de ordem: %+v", rede.Scripts)
	}
	if !rede.Scripts[1].Admin || rede.Scripts[1].Formato != "sh" || rede.Scripts[0].Formato != "py" {
		t.Errorf("metadados de Rede errados: %+v %+v", rede.Scripts[0], rede.Scripts[1])
	}
	if len(rede.Subcategorias) != 1 || rede.Subcategorias[0].Nome != "Wifi" ||
		rede.Subcategorias[0].Scripts[0].Caminho != "10-rede/wifi/perfis.ps1" {
		t.Errorf("subcategoria Wifi errada: %+v", rede.Subcategorias)
	}

	sistema := raiz.Subcategorias[1]
	if len(sistema.Scripts) != 2 {
		t.Errorf("Sistema deveria ter so info e so-deste: %+v", sistema.Scripts)
	}
	for _, s := range sistema.Scripts {
		if s.Nome == "So do outro" {
			t.Errorf("script com @so %s apareceu no %s", outroSO, runtime.GOOS)
		}
	}
}

func TestNormalizar(t *testing.T) {
	// Batch: LF vira CRLF; com acento em UTF-8 ganha chcp 65001.
	if got := string(normalizarBatch([]byte("@echo off\necho oi\n"))); got != "@echo off\r\necho oi\r\n" {
		t.Errorf("batch ASCII = %q", got)
	}
	if got := string(normalizarBatch([]byte("echo máquina\r\n"))); !strings.HasPrefix(got, "@chcp 65001 >nul\r\n") {
		t.Errorf("batch com acento sem chcp: %q", got)
	}
	if got := string(normalizarBatch([]byte("chcp 850\r\necho máquina\r\n"))); strings.Contains(got, "65001") {
		t.Errorf("batch que ja define chcp nao deve ganhar outro: %q", got)
	}
	if ansi := []byte("echo m\xe1quina\r\n"); !bytes.Equal(normalizarBatch(ansi), ansi) {
		t.Errorf("batch ANSI (UTF-8 invalido) deve ficar intacto")
	}

	// PowerShell: BOM so quando ha acento e ainda nao ha BOM.
	if got := normalizarPowerShell([]byte("Write-Host oi")); bytes.HasPrefix(got, []byte(bomUTF8)) {
		t.Errorf("ps1 ASCII nao precisa de BOM")
	}
	if got := normalizarPowerShell([]byte("Write-Host memória")); !bytes.HasPrefix(got, []byte(bomUTF8)) {
		t.Errorf("ps1 com acento sem BOM")
	}
	comBOM := []byte(bomUTF8 + "Write-Host memória")
	if got := normalizarPowerShell(comBOM); !bytes.Equal(got, comBOM) {
		t.Errorf("ps1 que ja tem BOM ganhou outro")
	}

	// VBScript: UTF-8 com acento vira UTF-16LE com BOM.
	if got := normalizarVBScript([]byte("MsgBox \"é\"\n")); !bytes.HasPrefix(got, []byte{0xFF, 0xFE, 'M', 0}) {
		t.Errorf("vbs com acento nao virou UTF-16LE: % x", got[:6])
	}

	// Shell: CRLF vira LF.
	if got := string(normalizarShell([]byte("echo oi\r\nexit 0\r\n"))); got != "echo oi\nexit 0\n" {
		t.Errorf("sh = %q", got)
	}
}
