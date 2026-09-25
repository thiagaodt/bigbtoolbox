// Package scripts e o motor do BigB: descobre os scripts da pasta scripts/,
// le os metadados e monta a arvore de categorias que vira o menu.
//
// Regras da pasta:
//   - cada subpasta vira uma categoria (e subpastas dela, subcategorias);
//   - cada arquivo com extensao de script conhecida vira um item do menu;
//   - arquivos e pastas que comecam com "_" ou "." nao aparecem no menu,
//     mas sao extraidos junto com os scripts da mesma pasta (recursos:
//     instaladores, drivers, arquivos de apoio);
//   - um prefixo numerico ("10-rede") define a ordem e some na exibicao;
//   - extensoes que o sistema atual nao roda (bat no Linux, por exemplo) e
//     scripts marcados com @so de outro sistema nao aparecem.
//
// Camada: Core. Nao desenha nada na tela.
package scripts

import (
	"io/fs"
	"path"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Script e um item executavel do menu.
type Script struct {
	Nome      string
	Descricao string
	Admin     bool   // exige Administrador
	Confirmar bool   // pede confirmacao antes de rodar
	Formato   string // extensao sem ponto, em minusculo: "bat", "ps1"...
	Caminho   string // caminho dentro do FS de scripts (com barras "/")
	// Sistemas onde o script aparece ("windows", "linux"). Vazio = em todo
	// sistema que tenha o formato.
	Sistemas []string
}

// Categoria e uma pasta de scripts.
type Categoria struct {
	Nome          string
	Caminho       string
	Subcategorias []*Categoria
	Scripts       []*Script
}

// Total conta os scripts da categoria, incluindo as subcategorias.
func (c *Categoria) Total() int {
	n := len(c.Scripts)
	for _, sub := range c.Subcategorias {
		n += sub.Total()
	}
	return n
}

// Carregar percorre o FS de scripts e monta a arvore de categorias.
// Categorias sem nenhum script (nem nas subpastas) ficam de fora.
func Carregar(fsys fs.FS) (*Categoria, error) {
	raiz := &Categoria{Nome: "Inicio", Caminho: "."}

	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return nil
		}
		if oculto(d.Name()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		formato := FormatoDe(p)
		if formato == "" {
			return nil // extensao desconhecida: pode ser recurso de um script
		}
		conteudo, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		s := lerMetadados(conteudo)
		if len(s.Sistemas) > 0 && !slices.Contains(s.Sistemas, runtime.GOOS) {
			return nil // @so de outro sistema
		}
		s.Formato, s.Caminho = formato, p
		if s.Nome == "" {
			s.Nome = NomeExibicao(strings.TrimSuffix(path.Base(p), path.Ext(p)))
		}
		cat := categoria(raiz, path.Dir(p))
		cat.Scripts = append(cat.Scripts, &s)
		return nil
	})
	if err != nil {
		return nil, err
	}

	podar(raiz)
	return raiz, nil
}

// categoria devolve (criando se preciso) a categoria da pasta dir.
func categoria(raiz *Categoria, dir string) *Categoria {
	if dir == "." {
		return raiz
	}
	atual := raiz
	for _, parte := range strings.Split(dir, "/") {
		caminho := path.Join(atual.Caminho, parte)
		i := slices.IndexFunc(atual.Subcategorias, func(c *Categoria) bool { return c.Caminho == caminho })
		if i < 0 {
			atual.Subcategorias = append(atual.Subcategorias, &Categoria{Nome: NomeExibicao(parte), Caminho: caminho})
			i = len(atual.Subcategorias) - 1
		}
		atual = atual.Subcategorias[i]
	}
	return atual
}

// podar tira as categorias vazias e ordena tudo pelo nome em disco (o que
// respeita o prefixo numerico).
func podar(c *Categoria) {
	c.Subcategorias = slices.DeleteFunc(c.Subcategorias, func(sub *Categoria) bool {
		podar(sub)
		return sub.Total() == 0
	})
	slices.SortFunc(c.Subcategorias, func(a, b *Categoria) int {
		return strings.Compare(strings.ToLower(a.Caminho), strings.ToLower(b.Caminho))
	})
	slices.SortFunc(c.Scripts, func(a, b *Script) int {
		return strings.Compare(strings.ToLower(a.Caminho), strings.ToLower(b.Caminho))
	})
}

func oculto(nome string) bool {
	return strings.HasPrefix(nome, "_") || strings.HasPrefix(nome, ".")
}

var (
	reOrdem     = regexp.MustCompile(`^\d+[-_. ]+`)
	reSeparador = regexp.MustCompile(`[-_]+`)
)

// NomeExibicao transforma um nome em disco no rotulo do menu:
// "10-rede-e-internet" -> "Rede e internet".
func NomeExibicao(nome string) string {
	nome = reOrdem.ReplaceAllString(nome, "")
	nome = strings.TrimSpace(reSeparador.ReplaceAllString(nome, " "))
	if nome == "" {
		return "(sem nome)"
	}
	r, n := utf8.DecodeRuneInString(nome)
	return string(unicode.ToUpper(r)) + nome[n:]
}

// reMetadado casa uma linha "@chave valor" dentro de um comentario de
// qualquer formato suportado (::, REM, #, ').
var reMetadado = regexp.MustCompile(`(?i)^\s*(?:::|rem\s|#|')?\s*@(nome|descri[cç][aã]o|admin|confirmar|so)\b\s*[:=]?\s*(.*?)\s*$`)

// linhasDeMetadados limita a busca ao cabecalho do arquivo.
const linhasDeMetadados = 50

// lerMetadados extrai @nome, @descricao, @admin, @confirmar e @so do
// cabecalho.
func lerMetadados(conteudo []byte) Script {
	var s Script
	texto := strings.TrimPrefix(string(conteudo), bomUTF8)
	for i, linha := range strings.Split(texto, "\n") {
		if i >= linhasDeMetadados {
			break
		}
		m := reMetadado.FindStringSubmatch(linha)
		if m == nil {
			continue
		}
		chave, valor := strings.ToLower(m[1]), m[2]
		switch {
		case chave == "nome":
			s.Nome = valor
		case strings.HasPrefix(chave, "descri"):
			s.Descricao = valor
		case chave == "admin":
			s.Admin = verdadeiro(valor)
		case chave == "confirmar":
			s.Confirmar = verdadeiro(valor)
		case chave == "so":
			// "@so windows" ou "@so windows, linux"
			s.Sistemas = strings.FieldsFunc(strings.ToLower(valor), func(r rune) bool {
				return r == ',' || unicode.IsSpace(r)
			})
		}
	}
	return s
}

// verdadeiro aceita sim/s/true/yes/1 (e chave sem valor, como "@admin").
func verdadeiro(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "sim", "s", "true", "yes", "y", "1":
		return true
	}
	return false
}
