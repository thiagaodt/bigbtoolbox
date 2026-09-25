// Package app e a camada de orquestracao: monta o menu a partir da arvore
// de scripts, navega entre as categorias e dispara a execucao.
//
// O BigB nao tem funcionalidade propria: tudo o que aparece no menu vem da
// pasta scripts/.
//
// Navegacao: setas movem a selecao, Enter (ou seta para a direita) abre,
// Esc/Backspace/seta para a esquerda volta, q ou Ctrl+C sai. Quando a
// entrada nao e um terminal (pipe, automacao), o menu cai para o modo
// numerado: digita-se o numero do item e Enter; 0 volta.
//
// Camada: App. Depende de: ui, scripts, config, logging, winsys.
package app

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"bigb/internal/config"
	"bigb/internal/logging"
	"bigb/internal/scripts"
	"bigb/internal/ui"
	"bigb/internal/winsys"
)

// item e uma linha do menu: uma subcategoria, um script ou o Voltar/Sair.
type item struct {
	rotulo, dica string
	admin        bool
	categoria    *scripts.Categoria
	script       *scripts.Script
}

// sair e ligado por q/Ctrl+C: fecha todos os niveis de menu de uma vez.
var sair bool

// Run e o ponto de entrada do aplicativo: abre o menu na raiz da arvore.
func Run(fsys fs.FS, raiz *scripts.Categoria) {
	logging.Infof("=== BigB iniciado (%d scripts) ===", raiz.Total())
	ui.Init()

	if raiz.Total() == 0 {
		mostrarSemScripts()
		return
	}

	if ui.Interativo() {
		ui.CursorVisivel(false)
		navegar(fsys, raiz, nil)
		ui.CursorVisivel(true)
	} else {
		navegarNumerado(fsys, raiz, nil)
	}

	ui.Header()
	ui.Blank()
	ui.Ok("Ate a proxima.")
	ui.Blank()
	logging.Infof("=== BigB encerrado ===")
}

// itensDe lista as linhas de uma categoria: subcategorias, scripts e, por
// ultimo, Voltar (ou Sair, na raiz).
func itensDe(cat *scripts.Categoria, raiz bool) []item {
	var itens []item
	for _, sub := range cat.Subcategorias {
		itens = append(itens, item{rotulo: sub.Nome + "  >", dica: plural(sub.Total(), "script", "scripts"), categoria: sub})
	}
	for _, s := range cat.Scripts {
		dica := "[" + s.Rotulo() + "]"
		if s.Descricao != "" {
			dica += " " + s.Descricao
		}
		itens = append(itens, item{rotulo: s.Nome, dica: dica, admin: s.Admin, script: s})
	}
	if raiz {
		itens = append(itens, item{rotulo: "Sair"})
	} else {
		itens = append(itens, item{rotulo: "Voltar"})
	}
	return itens
}

// navegar e o menu por setas de uma categoria. Devolve quando o tecnico
// volta, sai ou a entrada acaba.
func navegar(fsys fs.FS, cat *scripts.Categoria, trilha []string) {
	trilha = append(trilha, cat.Nome)
	itens := itensDe(cat, len(trilha) == 1)
	sel := 0

	for !sair {
		ui.Quadro(func() { desenhar(itens, sel, trilha) })

		tecla, r := ui.LerTecla()
		switch tecla {
		case ui.TeclaCima:
			sel = (sel - 1 + len(itens)) % len(itens)
		case ui.TeclaBaixo:
			sel = (sel + 1) % len(itens)
		case ui.TeclaInicio, ui.TeclaPaginaCima:
			sel = 0
		case ui.TeclaFim, ui.TeclaPaginaBaixo:
			sel = len(itens) - 1
		case ui.TeclaEsc, ui.TeclaEsquerda, ui.TeclaApagar:
			return
		case ui.TeclaInterromper, ui.TeclaFimEntrada:
			sair = true
			return
		case ui.TeclaEnter, ui.TeclaDireita:
			it := itens[sel]
			switch {
			case it.categoria != nil:
				navegar(fsys, it.categoria, trilha)
			case it.script != nil:
				ui.CursorVisivel(true)
				executar(fsys, it.script)
				ui.CursorVisivel(false)
			default:
				return // Voltar / Sair
			}
		case ui.TeclaOutra:
			if r == 'q' || r == 'Q' {
				sair = true
				return
			}
		}
	}
}

// desenhar monta a tela de uma categoria com o item sel destacado. Quando a
// lista nao cabe na janela, mostra so uma faixa em volta da selecao.
func desenhar(itens []item, sel int, trilha []string) {
	tema := config.GetTheme()
	ui.Header()
	ui.Blank()
	ui.Rule(strings.ToUpper(strings.Join(trilha, " > ")))
	ui.Blank()

	// Cabecalho + titulo (3) + indicadores de rolagem (2) + rodape (4).
	cabem := max(ui.Height()-ui.LinhasCabecalho()-9, 3)
	inicio, fim := 0, len(itens)
	if len(itens) > cabem {
		inicio = min(max(sel-cabem/2, 0), len(itens)-cabem)
		fim = inicio + cabem
	}

	if inicio > 0 {
		ui.Line(tema.Muted, "      ^ mais %d", inicio)
	}
	for i := inicio; i < fim; i++ {
		it := itens[i]
		if it.categoria == nil && it.script == nil {
			ui.Blank() // separa o Voltar/Sair do resto
		}
		ui.Opcao(it.rotulo, it.dica, it.admin, i == sel)
	}
	if fim < len(itens) {
		ui.Line(tema.Muted, "      v mais %d", len(itens)-fim)
	}

	ui.Blank()
	ui.Rule("")
	ui.Line(tema.Muted, "  Setas: mover   Enter: abrir/executar   Esc: voltar   Q: sair")
}

// navegarNumerado e o menu para entrada que nao e terminal (pipe): o
// tecnico (ou um script) digita o numero do item e Enter.
func navegarNumerado(fsys fs.FS, cat *scripts.Categoria, trilha []string) {
	tema := config.GetTheme()
	trilha = append(trilha, cat.Nome)
	itens := itensDe(cat, len(trilha) == 1)

	for !sair {
		ui.Header()
		ui.Blank()
		ui.Rule(strings.ToUpper(strings.Join(trilha, " > ")))
		ui.Blank()
		for i, it := range itens[:len(itens)-1] {
			ui.MenuItem(strconv.Itoa(i+1), it.rotulo, it.dica, it.admin)
		}
		ui.Blank()
		ui.MenuItem("0", itens[len(itens)-1].rotulo, "", false)
		ui.Blank()

		escolha := strings.TrimSpace(ui.Ask(tema.Accent, "  Escolha uma opcao: "))
		if ui.EntradaEncerrada() && escolha == "" {
			sair = true
			return
		}
		switch escolha {
		case "0":
			return
		case "":
			continue
		}
		n, err := strconv.Atoi(escolha)
		if err != nil || n < 1 || n >= len(itens) {
			ui.Err("Opcao invalida: %q", escolha)
			continue
		}
		if it := itens[n-1]; it.categoria != nil {
			navegarNumerado(fsys, it.categoria, trilha)
		} else {
			executar(fsys, it.script)
		}
	}
}

// executar confere as exigencias do script, roda e mostra o resultado.
func executar(fsys fs.FS, s *scripts.Script) {
	tema := config.GetTheme()
	ui.Header()
	ui.Blank()
	ui.Rule(strings.ToUpper(s.Nome))
	if s.Descricao != "" {
		ui.Line(tema.Muted, "  %s", s.Descricao)
	}
	ui.Line(tema.Muted, "  %s", s.Caminho)

	if s.Admin && !winsys.IsAdmin() {
		ui.Blank()
		ui.Err("Este script exige administrador. %s", winsys.DicaElevacao())
		ui.Pause()
		return
	}
	if s.Confirmar && !ui.Confirm("Executar este script agora?") {
		return
	}

	ui.Blank()
	ui.Rule("")
	logging.Infof("script: %s", s.Caminho)
	codigo, err := scripts.Executar(fsys, s)
	ui.Rule("")
	ui.Blank()

	switch {
	case err != nil:
		ui.Err("%v", err)
		logging.Errorf("script %s: %v", s.Caminho, err)
	case codigo != 0:
		ui.Warn("O script terminou com codigo %d.", codigo)
		logging.Warnf("script %s: codigo %d", s.Caminho, codigo)
	default:
		ui.Ok("Concluido.")
		logging.Infof("script %s: OK", s.Caminho)
	}
	ui.Pause()
}

// mostrarSemScripts explica o que fazer quando nao ha scripts.
func mostrarSemScripts() {
	ui.Header()
	ui.Blank()
	ui.Warn("Nenhum script encontrado.")
	ui.Blank()
	ui.Plain("  Coloque scripts (%s) em scripts/<categoria>/", strings.Join(scripts.FormatosSuportados(), ", "))
	ui.Plain("  e recompile, ou aponte uma pasta externa com: bigb -scripts <pasta>")
	ui.Pause()
}

func plural(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}
