package tui

import (
	"errors"
	"fmt"
	"strconv"

	"codeberg.org/malik/riscvemu/internal/core"
	"github.com/charmbracelet/huh"
)

// ErrCancelled is the sentinel error returned when the user
// aborts a Huh form (Ctrl+C, Esc). Callers use IsCancelled to
// distinguish user cancellation from real errors and avoid
// surfacing a noisy error message.
var ErrCancelled = errors.New("form cancelled")

// IsCancelled reports whether the error is the form-cancellation
// sentinel.
func IsCancelled(err error) bool {
	return errors.Is(err, ErrCancelled)
}

// ValidateLoadProgramPath is the headless-testable validator
// used by LoadProgramForm. It returns nil for non-empty paths
// and an error for empty ones. The actual file existence is
// not checked here; LoadProgramFromFile does that.
func ValidateLoadProgramPath(path string) error {
	if path == "" {
		return errors.New("path is required")
	}
	return nil
}

// ValidatePositiveInt is the headless-testable validator used
// by ConfigForm for every reservation-station count field. It
// returns nil for positive integers and an error for anything
// else.
func ValidatePositiveInt(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("must be an integer, got %q", s)
	}
	if n <= 0 {
		return fmt.Errorf("must be positive, got %d", n)
	}
	return nil
}

// LoadProgramForm runs a Huh form that asks for a file path
// and calls app.LoadProgramFromFile on submit. The form
// blocks until the user submits or aborts. The returned error
// is ErrCancelled on abort, the LoadProgramFromFile error on
// load failure, or a non-nil huh error on unexpected form
// failures.
func LoadProgramForm(app *core.App) error {
	if app == nil {
		return errors.New("LoadProgramForm: app is nil")
	}
	var path string
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Program file path").
				Description("Path to a RISC-V assembly file").
				Value(&path).
				Validate(ValidateLoadProgramPath),
		),
	).WithTheme(huh.ThemeCharm())
	if err := f.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrCancelled
		}
		return fmt.Errorf("LoadProgramForm: %w", err)
	}
	return app.LoadProgramFromFile(path)
}

// ConfigForm runs a Huh form that mutates the pipeline
// configuration. The form pre-fills the current RS counts;
// on submit it parses the strings, builds a new cpu.Config,
// and calls app.Rebuild. The returned error follows the same
// rules as LoadProgramForm.
func ConfigForm(app *core.App) error {
	if app == nil {
		return errors.New("ConfigForm: app is nil")
	}
	cfg := app.Cfg()
	aluStr := strconv.Itoa(cfg.ALURSCount)
	lsuStr := strconv.Itoa(cfg.LSURSCount)
	mulStr := strconv.Itoa(cfg.MulRSCount)
	divStr := strconv.Itoa(cfg.DivRSCount)

	f := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("ALU RS count").
				Value(&aluStr).
				Validate(ValidatePositiveInt),
			huh.NewInput().
				Title("LSU RS count").
				Value(&lsuStr).
				Validate(ValidatePositiveInt),
			huh.NewInput().
				Title("MUL RS count").
				Value(&mulStr).
				Validate(ValidatePositiveInt),
			huh.NewInput().
				Title("DIV RS count").
				Value(&divStr).
				Validate(ValidatePositiveInt),
		),
	).WithTheme(huh.ThemeCharm())
	if err := f.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrCancelled
		}
		return fmt.Errorf("ConfigForm: %w", err)
	}

	alu, _ := strconv.Atoi(aluStr)
	lsu, _ := strconv.Atoi(lsuStr)
	mul, _ := strconv.Atoi(mulStr)
	div, _ := strconv.Atoi(divStr)
	newCfg := cfg
	newCfg.ALURSCount = alu
	newCfg.LSURSCount = lsu
	newCfg.MulRSCount = mul
	newCfg.DivRSCount = div
	return app.Rebuild(newCfg)
}
