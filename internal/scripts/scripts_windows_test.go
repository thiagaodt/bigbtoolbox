package scripts

import (
	"testing"
	"testing/fstest"
)

func TestFormatoDe(t *testing.T) {
	casos := map[string]string{
		"a.bat": "bat", "a.CMD": "cmd", "x/y.Ps1": "ps1", "a.vbs": "vbs",
		"a.py": "py", "a.sh": "sh", "a.exe": "", "a.txt": "", "semext": "",
	}
	for nome, esperado := range casos {
		if got := FormatoDe(nome); got != esperado {
			t.Errorf("FormatoDe(%q) = %q, esperado %q", nome, got, esperado)
		}
	}
}

func TestExecutarBatch(t *testing.T) {
	// O script so sai com 7 se o recurso irmao e o da pasta oculta tiverem
	// sido extraidos junto com ele. LF puro de proposito (vira CRLF).
	fsys := fstest.MapFS{
		"cat/teste.bat":         {Data: []byte("@echo off\nif not exist apoio.txt exit /b 1\nif not exist _rec\\dado.txt exit /b 2\ncall :sair\n:sair\nexit /b 7\n")},
		"cat/apoio.txt":         {Data: []byte("x")},
		"cat/_rec/dado.txt":     {Data: []byte("y")},
		"outra/nao-extrair.txt": {Data: []byte("z")},
	}
	s := &Script{Formato: "bat", Caminho: "cat/teste.bat"}
	codigo, err := Executar(fsys, s)
	if err != nil {
		t.Fatal(err)
	}
	if codigo != 7 {
		t.Errorf("codigo = %d, esperado 7 (1 = faltou apoio.txt, 2 = faltou _rec)", codigo)
	}
}

func TestExecutarPowerShell(t *testing.T) {
	// Sem BOM e com acento: so passa se a normalizacao incluir o BOM, senao
	// o PowerShell 5.1 le "é" como dois caracteres ANSI.
	fsys := fstest.MapFS{
		"teste.ps1": {Data: []byte("if ('é'.Length -ne 1) { exit 3 }\nexit 5\n")},
	}
	s := &Script{Formato: "ps1", Caminho: "teste.ps1"}
	if _, err := Interpretador(s); err != nil {
		t.Skip(err)
	}
	codigo, err := Executar(fsys, s)
	if err != nil {
		t.Fatal(err)
	}
	if codigo != 5 {
		t.Errorf("codigo = %d, esperado 5 (3 = acento lido errado)", codigo)
	}
}
