package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mickamy/dbtop/internal/driver/postgres"
	"github.com/mickamy/dbtop/internal/exit"
	"github.com/mickamy/dbtop/internal/tui"
)

// connectTimeout bounds the initial connect + version probe.
const connectTimeout = 10 * time.Second

// defaultInterval is the poll interval until --interval is wired up.
const defaultInterval = time.Second

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		PrintUsage(stderr)

		return exit.Usage
	}

	dsn := args[0]

	if strings.HasPrefix(dsn, "mysql") {
		fmt.Fprintln(stderr, "dbtop: the MySQL driver is not implemented yet")

		return exit.Error
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	d, err := postgres.Open(ctx, dsn)
	if err != nil {
		fmt.Fprintf(stderr, "dbtop: %v\n", err)

		return exit.Error
	}
	defer func() { _ = d.Close() }()

	caps, err := d.Capabilities(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "dbtop: %v\n", err)

		return exit.Error
	}

	program := tea.NewProgram(tui.New(d, caps, defaultInterval), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(stderr, "dbtop: %v\n", err)

		return exit.Error
	}

	return exit.OK
}

func PrintUsage(w io.Writer) {
	fmt.Fprintln(w, "dbtop — top for Postgres and MySQL.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "USAGE:")
	fmt.Fprintln(w, "  dbtop <dsn> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Opens a live monitor for active queries, locks, who's blocking whom,")
	fmt.Fprintln(w, "  and the costliest normalized queries over time.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "EXAMPLES:")
	fmt.Fprintln(w, "  dbtop postgres://postgres:pass@127.0.0.1:5432/dev")
	fmt.Fprintln(w, "  dbtop mysql://root:pass@127.0.0.1:3306/dev")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FLAGS:")
	fmt.Fprintln(w, "  --interval <dur>   Poll interval (default 1s)")
	fmt.Fprintln(w, "  --ascii            ASCII-only glyphs (no Unicode gutter markers)")
	fmt.Fprintln(w, "  -a, --all          Show idle connections and background workers")
	fmt.Fprintln(w, "  --version, -v      Print dbtop version")
	fmt.Fprintln(w, "  --help, -h         Show this help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "KEYS:")
	fmt.Fprintln(w, "  tab          Switch screen (Activity / Metrics / Statements)")
	fmt.Fprintln(w, "  j/k ↑↓       Move")
	fmt.Fprintln(w, "  enter        Detail pane")
	fmt.Fprintln(w, "  K            Kill / terminate a backend")
	fmt.Fprintln(w, "  r            Reset statement stats")
	fmt.Fprintln(w, "  / s i ␣ q    Filter / sort / toggle idle / pause / quit")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Supports PostgreSQL 14+ and MySQL 8.0+.")
	fmt.Fprintln(w, "More: https://github.com/mickamy/dbtop")
}
