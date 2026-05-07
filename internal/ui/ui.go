package ui

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ─── ANSI COLOR PALETTE ─────────────────────────────────────────────────────
// Inspired by the FileOnix logo: dark navy bg, chrome white, electric blue gradient

const (
	Reset = "\033[0m"

	// Blues (logo gradient: light to electric)
	BlueDeep    = "\033[38;2;30;100;200m"
	BlueElec    = "\033[38;2;0;160;255m"
	BlueBright  = "\033[38;2;100;210;255m"
	BlueAccent  = "\033[38;2;0;200;255m"
	CyanGlow    = "\033[38;2;80;240;255m"

	// Chrome / Silver (logo "File" text)
	Chrome      = "\033[38;2;220;230;245m"
	Silver      = "\033[38;2;180;195;215m"
	SteelGray   = "\033[38;2;130;150;175m"

	// Whites
	White       = "\033[38;2;255;255;255m"
	WhiteDim    = "\033[38;2;200;210;225m"

	// Status colors
	Green       = "\033[38;2;80;230;140m"
	GreenSoft   = "\033[38;2;60;200;120m"
	Yellow      = "\033[38;2;255;210;60m"
	Orange      = "\033[38;2;255;155;50m"
	Red         = "\033[38;2;255;80;80m"
	RedSoft     = "\033[38;2;220;60;60m"
	Pink        = "\033[38;2;255;100;180m"

	// Background
	BgNavy      = "\033[48;2;8;15;30m"
	BgDark      = "\033[48;2;5;10;20m"
	BgBlue      = "\033[48;2;0;50;120m"
	BgRed       = "\033[48;2;180;20;20m"
	BgOrange    = "\033[48;2;180;80;0m"
	BgYellow    = "\033[48;2;160;130;0m"

	// Styles
	Bold        = "\033[1m"
	Dim         = "\033[2m"
	Italic      = "\033[3m"
	Underline   = "\033[4m"
)

// ─── SYMBOLS ────────────────────────────────────────────────────────────────

const (
	SymDiamond  = "◆"
	SymArrow    = "›"
	SymArrowR   = "»"
	SymDot      = "•"
	SymCircle   = "○"
	SymFilled   = "●"
	SymStar     = "★"
	SymTick     = "✓"
	SymCross    = "✗"
	SymWarn     = "⚠"
	SymInfo     = "◉"
	SymPulse    = "◈"
	SymSlash    = "/"
	SymBar      = "│"
	SymCorner   = "╭"
	SymEnd      = "╰"
	SymLine     = "─"
	SymBullet   = "⬡"
	SymSpark    = "⚡"
	SymEye      = "◎"
	SymChain    = "⬢"
	SymFlame    = "◈"
)

// ─── BANNER ─────────────────────────────────────────────────────────────────

func PrintBanner(version string) {
	width := 62

	fmt.Println()

	// Top border with glow effect
	printGlowLine(width)
	fmt.Println()

	// Logo lines - ASCII art styled like the logo
	logo := []struct{ left, right string }{
		{Chrome + Bold + "  ███████╗██╗██╗     ███████╗", BlueElec + Bold + "  ██████╗ ███╗   ██╗██╗██╗  ██╗"},
		{Chrome + Bold + "  ██╔════╝██║██║     ██╔════╝", BlueElec + Bold + " ██╔═══██╗████╗  ██║██║╚██╗██╔╝"},
		{Chrome + Bold + "  █████╗  ██║██║     █████╗  ", BlueElec + Bold + " ██║   ██║██╔██╗ ██║██║ ╚███╔╝ "},
		{Chrome + Bold + "  ██╔══╝  ██║██║     ██╔══╝  ", BlueElec + Bold + " ██║   ██║██║╚██╗██║██║ ██╔██╗ "},
		{Chrome + Bold + "  ██║     ██║███████╗███████╗", BlueElec + Bold + " ╚██████╔╝██║ ╚████║██║██╔╝ ██╗"},
		{Chrome + Bold + "  ╚═╝     ╚═╝╚══════╝╚══════╝", BlueElec + Bold + "  ╚═════╝ ╚═╝  ╚═══╝╚═╝╚═╝  ╚═╝"},
	}

	for i, line := range logo {
		delay := time.Duration(i) * 18 * time.Millisecond
		time.Sleep(delay)
		fmt.Println(line.left + line.right + Reset)
	}

	fmt.Println()

	// Swoosh line (like the logo's blue underline)
	printSwoosh(width)

	fmt.Println()

	// Tagline row
	tagLeft  := BlueDeep + Bold + "  High-Performance Native Watcher System" + Reset
	tagRight := Dim + Silver + "  v" + version + Reset
	fmt.Println(tagLeft + tagRight)

	// Sub-tagline
	fmt.Println(Dim + SteelGray + "  " + SymDiamond + " Engineered by " + Reset + BlueElec + "Nehonix Team" + Reset + Dim + SteelGray + "  —  https://github.com/Nehonix-Team" + Reset)

	fmt.Println()
	printGlowLine(width)
	fmt.Println()

	time.Sleep(40 * time.Millisecond)
}

func printGlowLine(width int) {
	// Gradient line: deep blue → electric blue → cyan → electric blue → deep blue
	segments := []struct {
		color string
		count int
	}{
		{BlueDeep, 8},
		{BlueElec, 10},
		{CyanGlow, 10},
		{BlueAccent + Bold, 6},
		{CyanGlow, 10},
		{BlueElec, 10},
		{BlueDeep, 8},
	}

	fmt.Print("  ")
	for _, s := range segments {
		fmt.Print(s.color + strings.Repeat(SymLine, s.count))
	}
	fmt.Print(Reset)
}

func printSwoosh(width int) {
	_ = width
	// Mimics the logo's sweeping blue underline
	fmt.Print("  ")
	fmt.Print(BlueDeep)
	fmt.Print(strings.Repeat("─", 12))
	fmt.Print(BlueElec + Bold)
	fmt.Print(strings.Repeat("─", 18))
	fmt.Print(CyanGlow + Bold)
	fmt.Print("━━━")
	fmt.Print(BlueAccent + Bold + " ✦ " + Reset)
	fmt.Print(BlueElec + Bold)
	fmt.Print(strings.Repeat("─", 18))
	fmt.Print(BlueDeep)
	fmt.Print(strings.Repeat("─", 7))
	fmt.Print(Reset)
	fmt.Println()
}

// ─── SECTION HEADERS ────────────────────────────────────────────────────────

func PrintSection(title, subtitle string) {
	fmt.Printf(
		"  %s%s %s%s %s%s%s\n",
		BlueElec+Bold, SymDiamond, Reset,
		Chrome+Bold, title, Reset,
		Dim+SteelGray+" "+SymArrow+" "+subtitle+Reset,
	)
}

func PrintSubSection(label string) {
	fmt.Printf("  %s%s%s %s%s%s\n",
		BlueDeep, SymBar, Reset,
		Silver, label, Reset,
	)
}

// ─── CONFIG DISPLAY ─────────────────────────────────────────────────────────

func PrintConfig(cfg interface{}) {
	// Use reflection-free display via the Stringer interface
	type ConfigDisplay interface {
		Display()
	}
	if cd, ok := cfg.(ConfigDisplay); ok {
		cd.Display()
		return
	}
	PrintSection("CONFIGURATION", "loaded from fileonix.config.json")
}

// PrintConfigFields displays config key/value pairs beautifully
func PrintConfigFields(fields []ConfigField) {
	fmt.Println()
	PrintSection("CONFIGURATION", "")
	fmt.Println()

	maxKey := 0
	for _, f := range fields {
		if len(f.Key) > maxKey {
			maxKey = len(f.Key)
		}
	}

	for _, f := range fields {
		padding := strings.Repeat(" ", maxKey-len(f.Key)+1)
		keyColor := SteelGray
		valColor := Chrome

		if f.Highlight {
			keyColor = BlueElec
			valColor = White + Bold
		}
		if f.Dim {
			keyColor = Dim + SteelGray
			valColor = Dim + Silver
		}

		icon := Dim + SteelGray + SymDot + Reset
		if f.Highlight {
			icon = BlueElec + SymDiamond + Reset
		}

		fmt.Printf("  %s  %s%s%s%s  %s%s%s\n",
			icon,
			keyColor, f.Key, Reset,
			padding,
			valColor, f.Value, Reset,
		)
	}
	fmt.Println()
}

type ConfigField struct {
	Key       string
	Value     string
	Highlight bool
	Dim       bool
}

// ─── EVENT LOGGING ──────────────────────────────────────────────────────────

type EventType int

const (
	EventChange EventType = iota
	EventCreate
	EventDelete
	EventRestart
	EventError
	EventSuccess
	EventInfo
	EventBoot
)

func Log(event EventType, message string, detail ...string) {
	ts := time.Now().Format("15:04:05")
	tsStr := Dim + SteelGray + ts + Reset

	var icon, color string
	switch event {
	case EventChange:
		icon = BlueElec + SymPulse
		color = BlueElec
	case EventCreate:
		icon = Green + SymTick
		color = Green
	case EventDelete:
		icon = Orange + SymCross
		color = Orange
	case EventRestart:
		icon = CyanGlow + Bold + SymSpark
		color = CyanGlow
	case EventError:
		icon = BgRed + White + Bold + " " + SymCross + " " + Reset
		color = Red + Bold
	case EventSuccess:
		icon = Green + Bold + SymTick
		color = Green
	case EventInfo:
		icon = Silver + SymInfo
		color = Silver
	case EventBoot:
		icon = BlueAccent + Bold + SymFlame
		color = BlueAccent
	}

	line := fmt.Sprintf("  %s%s%s  %s  %s%s%s",
		icon, Reset,
		tsStr,
		color+message+Reset,
		Dim+SteelGray,
		strings.Join(detail, " "+SymArrow+" "),
		Reset,
	)
	fmt.Println(line)
}

// ─── PROCESS STATUS ─────────────────────────────────────────────────────────

func PrintProcessStart(runner, script string) {
	fmt.Println()
	fmt.Printf("  %s%s%s  %sStarting process%s  %s%s %s%s %s%s%s\n",
		CyanGlow+Bold, SymSpark, Reset,
		Chrome+Bold, Reset,
		Dim+SteelGray, runner, SymArrow, Reset,
		BlueElec+Bold, script, Reset,
	)
}

func PrintProcessStop(exitCode int, duration time.Duration) {
	durStr := formatDuration(duration)
	if exitCode == 0 {
		fmt.Printf("  %s%s%s  %sProcess exited cleanly%s  %s%s%s  %s%s%s\n",
			Green+Bold, SymTick, Reset,
			GreenSoft, Reset,
			Dim+SteelGray, durStr, Reset,
			Dim+SteelGray, "code 0", Reset,
		)
	} else {
		fmt.Printf("  %s %s CRASH %s  %s%s%s  %s%s%s  %sCode %d%s\n",
			BgRed+White+Bold, SymCross, Reset,
			Red+Bold, "Process failed", Reset,
			Dim+SteelGray, durStr, Reset,
			Orange, exitCode, Reset,
		)
	}
}

func PrintFileChange(path string, changeType string) {
	// Shorten path for display
	display := shortenPath(path, 48)

	var icon, color string
	switch changeType {
	case "modified":
		icon = BlueElec + SymPulse
		color = BlueElec
	case "created":
		icon = Green + SymTick
		color = Green
	case "deleted":
		icon = Orange + SymCross
		color = Orange
	default:
		icon = Silver + SymDot
		color = Silver
	}

	fmt.Printf("  %s%s%s  %s%s%s\n",
		icon, Reset,
		Dim+SteelGray+time.Now().Format("15:04:05")+Reset,
		color+changeType+Reset,
		Dim+SteelGray+" "+SymArrow+" "+Reset,
		Chrome+display+Reset,
	)
}

// ─── DIVIDER ────────────────────────────────────────────────────────────────

func PrintRestartDivider(count int) {
	fmt.Println()
	fmt.Printf("  %s%s%s  %s%s RESTART #%d %s%s  %s\n",
		CyanGlow+Bold, strings.Repeat(SymLine, 6), Reset,
		CyanGlow+Bold, SymSpark, count, SymSpark, Reset,
		Dim+SteelGray+strings.Repeat(SymLine, 30)+Reset,
	)
	fmt.Println()
}

// ─── FATAL / ERROR ──────────────────────────────────────────────────────────

func Fatal(title, message string) {
	fmt.Println()
	fmt.Printf("  %s %s FATAL ERROR %s\n", BgRed+White+Bold, SymCross, Reset)
	fmt.Printf("  %s%s%s\n", Red+Bold, title, Reset)
	fmt.Printf("  %s%s%s\n", WhiteDim, message, Reset)
	fmt.Println()
	os.Exit(1)
}

func Warn(message string) {
	fmt.Printf("  %s %s WARN %s  %s%s%s\n",
		BgYellow+White+Bold, SymWarn, Reset,
		Yellow, message, Reset,
	)
}

func Info(message string) {
	fmt.Printf("  %s%s%s  %s%s%s\n",
		SteelGray, SymInfo, Reset,
		Silver, message, Reset,
	)
}

func Success(message string) {
	fmt.Printf("  %s%s%s  %s%s%s\n",
		Green+Bold, SymTick, Reset,
		Green, message, Reset,
	)
}

// ─── HELP ────────────────────────────────────────────────────────────────────

func ShowHelp(version string) {
	PrintBanner(version)

	fmt.Printf("  %s%s USAGE%s\n\n", Chrome+Bold, SymDiamond+" ", Reset)
	fmt.Printf("    %sfileonix%s %s[options]%s\n\n", BlueElec+Bold, Reset, Silver, Reset)

	printHelpSection("OPTIONS")
	printHelpRow("-script <file>",   "Entry point to watch and execute (optional)", true)
	printHelpRow("-watch <dirs>",    "Comma-separated directories to watch", false)
	printHelpRow("-ext <exts>",      "Extensions to watch (default: .ts,.js)", false)
	printHelpRow("-runner <name>",   "Runtime: bun | tsx | ts-node | node | bash | etc.", false)
	printHelpRow("-ignore <dirs>",   "Comma-separated directories to ignore", false)
	printHelpRow("-delay <ms>",      "Debounce delay in milliseconds (default: 100)", false)
	printHelpRow("-batch",           "Enable batch mode (group rapid changes)", false)
	printHelpRow("-clear",           "Clear screen on each restart", false)
	printHelpRow("--no-hash",        "Disable file hash change detection", false)
	printHelpRow("--migrate",        "Migrate deprecated 'typescriptRunner' to 'runner'", false)
	printHelpRow("--version, -v",    "Show version information", false)
	printHelpRow("--help, -h",       "Show this help message", false)

	fmt.Println()
	printHelpSection("COMMANDS")
	printHelpRow("init",             "Generate a fileonix.config.json in current directory", true)

	fmt.Println()
	printHelpSection("EXAMPLES")
	fmt.Printf("    %sfileonix%s %s-script src/index.ts%s\n", BlueElec+Bold, Reset, Chrome, Reset)
	fmt.Printf("    %sfileonix%s %s-watch src,internal -ignore dist%s\n", BlueElec+Bold, Reset, Chrome, Reset)
	fmt.Printf("    %sfileonix%s %s-script server.ts -watch src,config -runner bun%s\n", BlueElec+Bold, Reset, Chrome, Reset)
	fmt.Printf("    %sfileonix%s %s-script app.ts -delay 200 -clear -batch%s\n\n", BlueElec+Bold, Reset, Chrome, Reset)

	printHelpSection("CONFIG FILE")
	fmt.Printf("    %sfileonix.config.json%s %sor%s %s.fileonixrc.json%s\n\n", Chrome+Bold, Reset, Dim+SteelGray, Reset, Chrome+Bold, Reset)

	configExample := `    {
      "script":           "src/index.ts",
      "watch":            ["src", "internal"],
      "ignore":           ["node_modules", "dist"],
      "runner":           "bun",
      "clearScreen":      true,
      "debounceMs":       100,
      "batchMode":        false
    }`

	lines := strings.Split(configExample, "\n")
	for _, l := range lines {
		fmt.Printf("%s%s%s\n", Dim+SteelGray, l, Reset)
	}
	fmt.Println()
}

func printHelpSection(name string) {
	fmt.Printf("  %s%s %s%s\n\n", BlueElec+Bold, SymDiamond, name, Reset)
}

func printHelpRow(flag, desc string, highlight bool) {
	flagColor := Silver
	descColor := Dim + SteelGray
	if highlight {
		flagColor = Chrome + Bold
		descColor = Silver
	}

	padding := strings.Repeat(" ", 26-len(flag))
	fmt.Printf("    %s%s%s%s%s%s%s\n",
		flagColor, flag, Reset,
		padding,
		descColor, desc, Reset,
	)
}

// ─── SPINNER ────────────────────────────────────────────────────────────────

type Spinner struct {
	frames  []string
	message string
	done    chan struct{}
	color   string
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:  []string{"◐", "◓", "◑", "◒"},
		message: message,
		done:    make(chan struct{}),
		color:   BlueElec,
	}
}

func (s *Spinner) Start() {
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				fmt.Printf("\r  %s%s%s  %s%s\n", Green+Bold, SymTick, Reset, Silver, s.message)
				return
			default:
				fmt.Printf("\r  %s%s%s  %s%s%s", s.color+Bold, s.frames[i%len(s.frames)], Reset, Silver, s.message, Reset)
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.done <- struct{}{}
	time.Sleep(50 * time.Millisecond)
}

// ─── HELPERS ────────────────────────────────────────────────────────────────

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func shortenPath(path string, max int) string {
	if len(path) <= max {
		return path
	}
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return "..." + path[len(path)-max:]
	}
	// Keep last 2-3 parts
	short := strings.Join(parts[len(parts)-3:], "/")
	if len(short) > max {
		short = ".../" + parts[len(parts)-1]
	} else {
		short = ".../" + short
	}
	return short
}
