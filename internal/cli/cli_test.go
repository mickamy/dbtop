package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mickamy/dbtop/internal/cli"
	"github.com/mickamy/dbtop/internal/exit"
)

func TestRunNoArgsPrintsUsage(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	if code := cli.Run(nil, &out, &errOut); code != exit.Usage {
		t.Errorf("exit code = %d, want %d", code, exit.Usage)
	}

	if !strings.Contains(errOut.String(), "USAGE:") {
		t.Errorf("usage not printed:\n%s", errOut.String())
	}
}

func TestRunMySQLNotImplemented(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	if code := cli.Run([]string{"mysql://root:pass@localhost:3306/db"}, &out, &errOut); code != exit.Error {
		t.Errorf("exit code = %d, want %d", code, exit.Error)
	}

	if !strings.Contains(errOut.String(), "MySQL") {
		t.Errorf("expected a MySQL not-implemented message:\n%s", errOut.String())
	}
}
