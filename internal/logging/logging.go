// Package logging mantem a trilha em arquivo do BigB Toolbox.
//
// Registro simples e sem dependencias: cada chamada acrescenta uma linha
// datada em logs\toolbox_AAAAMMDD.log. Serve de trilha para diagnosticar o que
// o tecnico executou na estacao (util nas acoes destrutivas: limpeza,
// encerramento de processos, reset de rede).
//
// Niveis: DEBUG, INFO, WARN, ERROR. Nao ha rotacao alem do arquivo-por-dia; a
// limpeza de logs antigos e responsabilidade do operador.
//
// Camada: Core / Infraestrutura. Depende de: config (pasta de logs).
package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bigb/internal/config"
)

// Level e a severidade de uma entrada de log.
type Level string

// Niveis aceitos.
const (
	Debug Level = "DEBUG"
	Info  Level = "INFO"
	Warn  Level = "WARN"
	Error Level = "ERROR"
)

var mu sync.Mutex

// File devolve o caminho do arquivo de log do dia atual.
func File() string {
	return filepath.Join(config.Get().LogDir, fmt.Sprintf("toolbox_%s.log", time.Now().Format("20060102")))
}

// Write acrescenta uma entrada ao log do dia. Nunca retorna erro: log jamais
// deve derrubar o app, entao falhas sao ignoradas em silencio.
func Write(level Level, format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()

	dir := config.Get().LogDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(File(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	fmt.Fprintf(f, "%s [%-5s] %s\r\n", time.Now().Format("2006-01-02 15:04:05"), level, msg)
}

// Infof, Warnf, Errorf e Debugf sao atalhos por nivel.
func Infof(format string, args ...any)  { Write(Info, format, args...) }
func Warnf(format string, args ...any)  { Write(Warn, format, args...) }
func Errorf(format string, args ...any) { Write(Error, format, args...) }
func Debugf(format string, args ...any) { Write(Debug, format, args...) }
