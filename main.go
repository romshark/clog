package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"
)

type Config struct {
	DisplaySeparator string

	StyleLevelDebug      []color.Attribute
	StyleLevelInfo       []color.Attribute
	StyleLevelWarn       []color.Attribute
	StyleLevelFatal      []color.Attribute
	StyleLevelError      []color.Attribute
	StyleCaller          []color.Attribute
	StyleTime            []color.Attribute
	StyleMessage         []color.Attribute
	StyleStringQuotes    []color.Attribute
	StyleKey             []color.Attribute
	StyleValue           []color.Attribute
	StyleArrayHead       []color.Attribute
	StyleArrayListPrefix []color.Attribute
	StyleSeparator       []color.Attribute

	FormatTime            string
	FormatStringQuotes    string
	FormatArray           string
	FormatArrayListPrefix string

	LevelDebug string
	LevelInfo  string
	LevelWarn  string
	LevelError string
	LevelFatal string

	FieldLevel   string
	FieldTime    string
	FieldMessage string
	FieldCaller  string
}

func MustParseConfig() (c Config, ignoreInterruptSig bool) {
	var configPath string
	flag.StringVar(
		&configPath, "c", "clog.toml",
		"path to configuration TOML file",
	)
	flag.BoolVar(
		&ignoreInterruptSig, "i", false,
		"ignore OS interrupt signals",
	)
	flag.Parse()
	var m map[string]interface{}
	if _, err := toml.DecodeFile(configPath, &m); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Fatal("parsing config:", err)
		}
	}

	getValue := func(category, key string, defaultVal string) string {
		c, ok := m[category]
		if !ok {
			return defaultVal
		}

		ks, ok := c.([]map[string]interface{})
		if !ok {
			return defaultVal
		}

		if len(ks) < 1 {
			return defaultVal
		}

		val, ok := ks[0][key]
		if !ok {
			return defaultVal
		}

		return val.(string)
	}

	parseStyleAttr := func(s string) (color.Attribute, error) {
		switch s {
		// Base attributes
		case "bold":
			return color.Bold, nil
		case "faint":
			return color.Faint, nil
		case "italic":
			return color.Italic, nil
		case "underline":
			return color.Underline, nil
		case "blinkslow":
			return color.BlinkSlow, nil
		case "blinkrapid":
			return color.BlinkRapid, nil
		case "reversevideo":
			return color.ReverseVideo, nil
		case "concealed":
			return color.Concealed, nil
		case "crossedout":
			return color.CrossedOut, nil
		// Foreground text colors
		case "fg-black":
			return color.FgBlack, nil
		case "fg-red":
			return color.FgRed, nil
		case "fg-green":
			return color.FgGreen, nil
		case "fg-yellow":
			return color.FgYellow, nil
		case "fg-blue":
			return color.FgBlue, nil
		case "fg-magenta":
			return color.FgMagenta, nil
		case "fg-cyan":
			return color.FgCyan, nil
		case "fg-white":
			return color.FgWhite, nil
		// High-intensity foreground text colors
		case "fg-hiblack":
			return color.FgHiBlack, nil
		case "fg-hired":
			return color.FgHiRed, nil
		case "fg-higreen":
			return color.FgHiGreen, nil
		case "fg-hiyellow":
			return color.FgHiYellow, nil
		case "fg-hiblue":
			return color.FgHiBlue, nil
		case "fg-himagenta":
			return color.FgHiMagenta, nil
		case "fg-hicyan":
			return color.FgHiCyan, nil
		case "fg-hiwhite":
			return color.FgHiWhite, nil
		// Background text colors
		case "bg-black":
			return color.BgBlack, nil
		case "bg-red":
			return color.BgRed, nil
		case "bg-green":
			return color.BgGreen, nil
		case "bg-yellow":
			return color.BgYellow, nil
		case "bg-blue":
			return color.BgBlue, nil
		case "bg-magenta":
			return color.BgMagenta, nil
		case "bg-cyan":
			return color.BgCyan, nil
		case "bg-white":
			return color.BgWhite, nil
		// High-intensity background text colors
		case "bg-hiblack":
			return color.BgHiBlack, nil
		case "bg-hired":
			return color.BgHiRed, nil
		case "bg-higreen":
			return color.BgHiGreen, nil
		case "bg-hiyellow":
			return color.BgHiYellow, nil
		case "bg-hiblue":
			return color.BgHiBlue, nil
		case "bg-himagenta":
			return color.BgHiMagenta, nil
		case "bg-hicyan":
			return color.BgHiCyan, nil
		case "bg-hiwhite":
			return color.BgHiWhite, nil
		}
		return 0, fmt.Errorf("unknown style attribute: %q", s)
	}

	getStyle := func(
		category, key string,
		defaultVal ...color.Attribute,
	) (v []color.Attribute) {
		if s := getValue(category, key, ""); s != "" {
			f := strings.Fields(s)
			for _, f := range f {
				a, err := parseStyleAttr(f)
				if err != nil {
					log.Fatalf("parsing %s.%s %s", category, key, err)
				}
				v = append(v, a)
			}
			return v
		}
		return defaultVal
	}

	c.DisplaySeparator = getValue("display", "separator", "\n")

	c.StyleLevelDebug = getStyle(
		"style", "level-debug", color.FgHiGreen, color.Bold,
	)
	c.StyleLevelInfo = getStyle(
		"style", "level-info", color.FgGreen, color.Bold,
	)
	c.StyleLevelWarn = getStyle(
		"style", "level-warn", color.FgYellow, color.Bold,
	)
	c.StyleLevelError = getStyle(
		"style", "level-error", color.FgHiRed, color.Bold,
	)
	c.StyleLevelFatal = getStyle(
		"style", "level-fatal", color.FgRed, color.Bold,
	)
	c.StyleTime = getStyle("style", "time", color.FgHiBlack)
	c.StyleMessage = getStyle("style", "message")
	c.StyleCaller = getStyle("style", "caller", color.FgHiBlack)
	c.StyleStringQuotes = getStyle("style", "string-quotes", color.FgHiBlack)
	c.StyleKey = getStyle("style", "key", color.FgHiBlue)
	c.StyleValue = getStyle("style", "value")
	c.StyleArrayHead = getStyle("style", "array-head", color.FgHiBlack)
	c.StyleArrayListPrefix = getStyle(
		"style", "array-list-prefix", color.FgHiBlack,
	)
	c.StyleSeparator = getStyle(
		"style", "separator", color.FgHiBlack,
	)

	c.FormatTime = getValue("format", "time", time.RFC1123)

	c.FormatStringQuotes = getValue("format", "string-quotes", `"`)

	c.FormatArray = getValue("format", "array", "list")
	if c.FormatArray != "list" && c.FormatArray != "raw" {
		log.Fatalf("unsupported array format: %q", c.FormatArray)
	}

	c.FormatArrayListPrefix = getValue("format", "array-list-prefix", "- %d:")

	c.LevelDebug = getValue("level", "debug", "debug")
	c.LevelInfo = getValue("level", "info", "info")
	c.LevelWarn = getValue("level", "warn", "warn")
	c.LevelError = getValue("level", "error", "error")
	c.LevelFatal = getValue("level", "fatal", "fatal")

	c.FieldTime = getValue("field", "time", "time")
	c.FieldMessage = getValue("field", "message", "msg")
	c.FieldCaller = getValue("field", "caller", "caller")
	c.FieldLevel = getValue("field", "level", "level")

	return
}

func main() {
	conf, ignoreInterruptSig := MustParseConfig()

	var maxLevelLabelLen int
	for _, l := range []int{
		len(conf.LevelDebug),
		len(conf.LevelInfo),
		len(conf.LevelWarn),
		len(conf.LevelError),
		len(conf.LevelFatal),
	} {
		if l > maxLevelLabelLen {
			maxLevelLabelLen = l
		}
	}

	if ignoreInterruptSig {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			for {
				<-sig
			}
		}()
	}

	scanner := bufio.NewScanner(os.Stdin)
	i := 0
	for scanner.Scan() {
		if i > 0 && conf.DisplaySeparator != "" {
			color.New(conf.StyleSeparator...).Print(conf.DisplaySeparator)
		}
		i++
		txt := scanner.Text()
		if !gjson.Valid(txt) {
			color.New(conf.StyleLevelError...).Println("non-JSON output:")
			color.New(conf.StyleMessage...).Println(txt)
			continue
		}
		j := gjson.Parse(txt)

		tm, err := time.Parse(time.RFC3339, j.Get("time").Str)
		if err != nil {
			log.Fatal("parsing time:", tm)
		}

		// Print level
		level := j.Get("level").Str
		switch level {
		case conf.LevelDebug:
			color.New(conf.StyleLevelDebug...).Print(strings.ToUpper(level))
		case conf.LevelInfo:
			color.New(conf.StyleLevelInfo...).Print(strings.ToUpper(level))
		case conf.LevelWarn:
			color.New(conf.StyleLevelWarn...).Print(strings.ToUpper(level))
		case conf.LevelError:
			color.New(conf.StyleLevelError...).Print(strings.ToUpper(level))
		case conf.LevelFatal:
			color.New(conf.StyleLevelFatal...).Print(strings.ToUpper(level))
		default:
			log.Fatalf("unsupported log level: %q", level)
		}

		for i := maxLevelLabelLen - len(level); i >= 0; i-- {
			fmt.Print(" ")
		}

		// Print time
		color.New(conf.StyleTime...).Println(tm.Format(conf.FormatTime))

		// Print file:line
		if c := j.Get(conf.FieldCaller); c.Str != "" {
			color.New(conf.StyleCaller...).Println(j.Get(conf.FieldCaller))
		}

		// Print message
		color.New(conf.StyleMessage...).Println(j.Get(conf.FieldMessage))

		j.ForEach(func(key, v gjson.Result) bool {
			if isBasicField(conf, key.String()) {
				// Ignore basic fields
				return true
			}
			color.New(conf.StyleKey...).Print(key, ":")
			fmt.Print(" ")
			printValue(conf, v)
			return true
		})
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func printValue(conf Config, v gjson.Result) {
	if len(v.Raw) > 0 && v.Raw[0] == '"' {
		if q := conf.FormatStringQuotes; q != "" {
			color.New(conf.StyleStringQuotes...).
				Print(conf.FormatStringQuotes)
			color.New(conf.StyleValue...).
				Print(v)
			color.New(conf.StyleStringQuotes...).
				Println(conf.FormatStringQuotes)
		} else {
			color.New(conf.StyleValue...).Println(v)
		}
	} else if conf.FormatArray == "list" && len(v.Raw) > 0 && v.Raw[0] == '[' {
		a := v.Array()
		color.New(conf.StyleArrayHead...).Printf("%d item(s)", len(a))
		fmt.Println()
		if len(a) < 1 {
			return
		}
		for i, x := range a {
			if p := conf.FormatArrayListPrefix; p != "" {
				color.New(conf.StyleArrayListPrefix...).Printf(p, i)
			}
			fmt.Print(" ")
			printValue(conf, x)
		}
	} else {
		color.New(conf.StyleValue...).Println(v)
	}
}

func isBasicField(conf Config, s string) bool {
	return s == conf.FieldCaller ||
		s == conf.FieldMessage ||
		s == conf.FieldTime ||
		s == conf.FieldLevel
}
