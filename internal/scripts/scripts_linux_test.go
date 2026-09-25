package scripts

import (
	"testing"
	"testing/fstest"
)

func TestFormatoDe(t *testing.T) {
	casos := map[string]string{
		"a.sh": "sh", "a.py": "py", "x/y.Ps1": "ps1",
		"a.bat": "", "a.cmd": "", "a.vbs": "", "a.exe": "", "semext": "",
	}
	for nome, esperado := range casos {
		if got := FormatoDe(nome); got != esperado {
			t.Errorf("FormatoDe(%q) = %q, esperado %q", nome, got, esperado)
		}
	}
}

func TestExecutarShell(t *testing.T) {
	// CRLF de proposito (vira LF); so sai com 7 se o recurso irmao e o da
	// pasta oculta tiverem sido extraidos junto.
	fsys := fstest.MapFS{
		"cat/teste.sh":      {Data: []byte("[ -f apoio.txt ] || exit 1\r\n[ -f _rec/dado.txt ] || exit 2\r\nexit 7\r\n")},
		"cat/apoio.txt":     {Data: []byte("x")},
		"cat/_rec/dado.txt": {Data: []byte("y")},
	}
	s := &Script{Formato: "sh", Caminho: "cat/teste.sh"}
	codigo, err := Executar(fsys, s)
	if err != nil {
		t.Fatal(err)
	}
	if codigo != 7 {
		t.Errorf("codigo = %d, esperado 7 (1 = faltou apoio.txt, 2 = faltou _rec)", codigo)
	}
}
