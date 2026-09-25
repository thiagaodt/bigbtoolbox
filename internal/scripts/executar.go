package scripts

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// bomUTF8 e a marca de ordem de bytes do UTF-8.
const bomUTF8 = "\xef\xbb\xbf"

// formato descreve como rodar uma extensao de script.
type formato struct {
	// Rotulo curto mostrado no menu.
	Rotulo string
	// Interpretadores aceitos, em ordem de preferencia (o primeiro achado
	// no PATH e usado).
	programas []string
	// Argumentos antes do nome do arquivo.
	args []string
	// Ajusta o conteudo antes de gravar (encoding, fim de linha).
	normalizar func([]byte) []byte
}

// FormatoDe devolve o formato ("bat", "ps1"...) de um arquivo, ou "" quando
// a extensao nao e de script.
func FormatoDe(nome string) string {
	ext := strings.ToLower(strings.TrimPrefix(path.Ext(nome), "."))
	if _, ok := formatos[ext]; ok {
		return ext
	}
	return ""
}

// FormatosSuportados lista as extensoes que este sistema sabe rodar
// (".bat", ".ps1"...), em ordem alfabetica.
func FormatosSuportados() []string {
	var lista []string
	for ext := range formatos {
		lista = append(lista, "."+ext)
	}
	slices.Sort(lista)
	return lista
}

// Rotulo devolve o rotulo curto do formato do script ("ps1", "bat"...).
func (s *Script) Rotulo() string { return formatos[s.Formato].Rotulo }

// Interpretador devolve o caminho do programa que roda o script, ou erro
// quando a maquina nao tem nenhum dos interpretadores aceitos.
func Interpretador(s *Script) (string, error) {
	f, ok := formatos[s.Formato]
	if !ok {
		return "", fmt.Errorf("formato desconhecido: %q", s.Formato)
	}
	for _, p := range f.programas {
		if caminho, err := exec.LookPath(p); err == nil {
			return caminho, nil
		}
	}
	return "", fmt.Errorf("nenhum interpretador para .%s nesta maquina (procurei %s)",
		s.Formato, strings.Join(f.programas, ", "))
}

// Executar roda um script e devolve o codigo de saida.
//
// O script e extraido para uma pasta temporaria privada junto com os
// arquivos da mesma pasta e as subpastas de recurso ("_"), e roda com o
// console herdado: ele pode escrever na tela e perguntar coisas ao tecnico.
// A pasta e apagada no fim.
func Executar(fsys fs.FS, s *Script) (int, error) {
	programa, err := Interpretador(s)
	if err != nil {
		return -1, err
	}

	pasta, err := pastaDeExecucao()
	if err != nil {
		return -1, err
	}
	defer os.RemoveAll(pasta)

	if err := extrair(fsys, s, pasta); err != nil {
		return -1, fmt.Errorf("extraindo o script: %w", err)
	}

	// Caminho relativo explicito: com NoDefaultCurrentDirectoryInExePath
	// ligado, o cmd nao procura "teste.bat" na pasta atual. O bash trata a
	// barra invertida como escape, entao para ele vai sempre "./".
	arquivo := "." + string(filepath.Separator) + path.Base(s.Caminho)
	if s.Formato == "sh" {
		arquivo = "./" + path.Base(s.Caminho)
	}
	f := formatos[s.Formato]
	args := append(append([]string{}, f.args...), arquivo)
	cmd := exec.Command(programa, args...)
	cmd.Dir = pasta
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	// Ctrl+C durante o script deve parar o script, nao o BigB: o console
	// entrega o sinal aos dois, entao o BigB o captura e ignora ate o fim.
	interrupcao := make(chan os.Signal, 1)
	signal.Notify(interrupcao, os.Interrupt)
	defer signal.Stop(interrupcao)

	defer preservarConsole()()

	err = cmd.Run()
	var saida *exec.ExitError
	if errors.As(err, &saida) {
		return saida.ExitCode(), nil
	}
	if err != nil {
		return -1, err
	}
	return 0, nil
}

// pastaDeExecucao cria uma pasta unica em <cache do usuario>/bigb/execucao
// (%LOCALAPPDATA% no Windows, ~/.cache no Linux).
//
// Fica fora da pasta temporaria de proposito: um script de limpeza de
// temporarios apagaria a si mesmo no meio da execucao.
func pastaDeExecucao() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	base = filepath.Join(base, "bigb", "execucao")
	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", err
	}
	return os.MkdirTemp(base, "*")
}

// extrair grava na pasta o script, os arquivos irmaos dele e as subpastas
// de recurso (as que comecam com "_" ou ".").
func extrair(fsys fs.FS, s *Script, destino string) error {
	dir := path.Dir(s.Caminho)
	itens, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return err
	}
	for _, item := range itens {
		origem := path.Join(dir, item.Name())
		switch {
		case item.IsDir() && oculto(item.Name()):
			if err := copiarPasta(fsys, origem, filepath.Join(destino, item.Name())); err != nil {
				return err
			}
		case !item.IsDir():
			conteudo, err := fs.ReadFile(fsys, origem)
			if err != nil {
				return err
			}
			if origem == s.Caminho {
				if n := formatos[s.Formato].normalizar; n != nil {
					conteudo = n(conteudo)
				}
			}
			if err := os.WriteFile(filepath.Join(destino, item.Name()), conteudo, 0o600); err != nil {
				return err
			}
		}
	}
	return nil
}

func copiarPasta(fsys fs.FS, origem, destino string) error {
	return fs.WalkDir(fsys, origem, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, origem), "/")
		alvo := filepath.Join(destino, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(alvo, 0o700)
		}
		conteudo, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		return os.WriteFile(alvo, conteudo, 0o600)
	})
}

// ------------------------------------------------------------ normalizacao
//
// Cada interpretador tem sua mania de encoding. Os scripts podem ser salvos
// em UTF-8 (o padrao dos editores atuais) e o BigB ajusta na extracao.

// temAcento informa se o conteudo e UTF-8 valido com caracteres nao ASCII.
func temAcento(b []byte) bool {
	if !utf8.Valid(b) {
		return false // provavelmente ANSI: o interpretador ja le assim
	}
	for _, c := range b {
		if c >= 0x80 {
			return true
		}
	}
	return false
}

// normalizarBatch garante CRLF (com LF puro o cmd erra goto/call :rotulo) e,
// se houver acento em UTF-8, liga o code page UTF-8 na primeira linha.
func normalizarBatch(b []byte) []byte {
	b = paraCRLF(b)
	if temAcento(b) && !bytes.Contains(bytes.ToLower(b), []byte("chcp")) {
		b = append([]byte("@chcp 65001 >nul\r\n"), b...)
	}
	return b
}

// normalizarPowerShell poe BOM no UTF-8 com acento: sem BOM o PowerShell
// 5.1 le o arquivo como ANSI e os acentos saem trocados.
func normalizarPowerShell(b []byte) []byte {
	if temAcento(b) && !bytes.HasPrefix(b, []byte(bomUTF8)) {
		b = append([]byte(bomUTF8), b...)
	}
	return b
}

// normalizarVBScript converte UTF-8 com acento para UTF-16LE com BOM, que
// o cscript entende (UTF-8 ele le como ANSI).
func normalizarVBScript(b []byte) []byte {
	b = bytes.TrimPrefix(b, []byte(bomUTF8))
	if !temAcento(b) {
		return b
	}
	unidades := utf16.Encode([]rune(string(paraCRLF(b))))
	out := make([]byte, 2, 2+2*len(unidades))
	out[0], out[1] = 0xFF, 0xFE
	for _, u := range unidades {
		out = append(out, byte(u), byte(u>>8))
	}
	return out
}

// normalizarShell troca CRLF por LF: o bash acusa "\r: command not found".
func normalizarShell(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}

func paraCRLF(b []byte) []byte {
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(b, []byte("\n"), []byte("\r\n"))
}
