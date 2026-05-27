package cli

import (
	"fmt"
	"io"

	"github.com/mickamy/dbtop/internal/exit"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		PrintUsage(stderr)

		return exit.Usage
	}

	dsn := args[0]

	fmt.Fprintf(stderr, "dbtop: live monitor not implemented yet (dsn %q)\n", dsn)

	return exit.NotImplemented
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
