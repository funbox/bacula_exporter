package daemon

// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"runtime"
	"strings"

	"github.com/avast/retry-go"

	"pkg.re/essentialkaos/ek.v12/fmtc"
	"pkg.re/essentialkaos/ek.v12/knf"
	"pkg.re/essentialkaos/ek.v12/log"
	"pkg.re/essentialkaos/ek.v12/options"
	"pkg.re/essentialkaos/ek.v12/signal"
	"pkg.re/essentialkaos/ek.v12/usage"

	rdbms "github.com/funbox/bacula_exporter/storage/rdbms"
	knfv "pkg.re/essentialkaos/ek.v12/knf/validators"
	knff "pkg.re/essentialkaos/ek.v12/knf/validators/fs"

	"github.com/prometheus/client_golang/prometheus"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic info
const (
	APP  = "bacula_exporter"
	VER  = "1.1.0"
	DESC = "Prometheus Exporter for Bacula"
)

// Options
const (
	OPT_CONFIG   = "c:config"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VERSION  = "v:version"
)

// Configuration file props
const (
	DB_USER       = "db:username"
	DB_PASSWORD   = "db:password"
	DB_HOST       = "db:host"
	DB_PORT       = "db:port"
	DB_NAME       = "db:name"
	DB_SSLMODE    = "db:sslmode"
	HTTP_IP       = "http:ip"
	HTTP_PORT     = "http:port"
	HTTP_ENDPOINT = "http:endpoint"
	LOG_OUTPUT    = "log:output"
	LOG_DIR       = "log:dir"
	LOG_FILE      = "log:file"
	LOG_PERMS     = "log:perms"
	LOG_LEVEL     = "log:level"
)

// Logger info
const (
	LOG_OUTPUT_FILE    = "file"
	LOG_OUTPUT_CONSOLE = "console"
)

// DB info
const (
	DB_FORMAT = "user=%s password=%s host=%s port=%s dbname=%s sslmode=%s"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Environment struct which contains models
type Env struct {
	DB rdbms.Datastore
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Options map
var optMap = options.Map{
	OPT_CONFIG:   {Value: "/etc/bacula_exporter.knf"},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL, Alias: "u:usage"},
	OPT_VERSION:  {Type: options.BOOL, Alias: "ver"},
}

var (
	env *Env
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Init() {
	runtime.GOMAXPROCS(8)

	_, errs := options.Parse(optMap)

	if len(errs) != 0 {
		for _, err := range errs {
			printError(err.Error())
		}

		os.Exit(1)
	}

	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}

	if options.GetB(OPT_VERSION) {
		showAbout()
		return
	}

	if options.GetB(OPT_HELP) {
		showUsage()
		return
	}

	loadConfig()

	registerSignalHandlers()
	setupLogger()

	log.Aux(strings.Repeat("-", 88))
	log.Aux("%s %s starting...", APP, VER)

	start()
}

// loadConfig read and parse configuration file
func loadConfig() {
	err := knf.Global(options.GetS(OPT_CONFIG))

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// validateConfig validate configuration file values
func validateConfig() {
	errs := knf.Validate([]*knf.Validator{
		{DB_USER, knfv.Empty, nil},
		{DB_PASSWORD, knfv.Empty, nil},
		{DB_HOST, knfv.Empty, nil},
		{DB_PORT, knfv.Less, 1},
		{DB_PORT, knfv.Greater, 65535},
		{DB_NAME, knfv.Empty, nil},
		{DB_SSLMODE, knfv.NotContains, []string{"disable", "verify-full", "verify-ca", "require"}},

		{HTTP_IP, knfv.Empty, nil},
		{HTTP_PORT, knfv.Empty, nil},
		{HTTP_PORT, knfv.Less, 1024},
		{HTTP_PORT, knfv.Greater, 65535},
		{HTTP_ENDPOINT, knfv.Empty, nil},

		{LOG_OUTPUT, knfv.NotContains, []string{LOG_OUTPUT_FILE, LOG_OUTPUT_CONSOLE}},
		{LOG_DIR, knfv.Empty, nil},
		{LOG_FILE, knfv.Empty, nil},
		{LOG_DIR, knff.Perms, "DW"},
		{LOG_DIR, knff.Perms, "DX"},
		{LOG_LEVEL, knfv.NotContains, []string{"debug", "info", "warn", "error", "crit"}},
	})

	if len(errs) != 0 {
		printError("Error while configuration file validation:")

		for _, err := range errs {
			printError("  %v", err)
		}

		os.Exit(1)
	}
}

// registerSignalHandlers register signal handlers
func registerSignalHandlers() {
	signal.Handlers{
		signal.TERM: termSignalHandler,
		signal.INT:  intSignalHandler,
		signal.HUP:  hupSignalHandler,
	}.TrackAsync()
}

// setupLogger setup logger
func setupLogger() {
	var err error

	if knf.GetS(LOG_OUTPUT) == LOG_OUTPUT_FILE {
		err := log.Set(knf.GetS(LOG_FILE), knf.GetM(LOG_PERMS, 644))

		if err != nil {
			printErrorAndExit(err.Error())
		}
	}

	err = log.MinLevel(knf.GetS(LOG_LEVEL))

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// buildConnectionString build DB connection string
func buildConnectionString() string {
	return fmtc.Sprintf(
		DB_FORMAT,
		knf.GetS(DB_USER),
		knf.GetS(DB_PASSWORD),
		knf.GetS(DB_HOST),
		knf.GetS(DB_PORT),
		knf.GetS(DB_NAME),
		knf.GetS(DB_SSLMODE),
	)
}

// start start service
func start() {
	_ = retry.Do(
		func() error {
			db, err := rdbms.NewDB(buildConnectionString())
			if err != nil {
				log.Crit(err.Error())
				return err
			}
			env = &Env{db}

			collector := baculaCollector()
			prometheus.MustRegister(collector)

			return nil
		},
	)

	err := startHTTPServer(
		knf.GetS(HTTP_IP),
		knf.GetS(HTTP_PORT),
		knf.GetS(HTTP_ENDPOINT),
	)

	if err != nil {
		log.Crit(err.Error())
		shutdown(1)
	}

	shutdown(0)
}

// INT signal handler
func intSignalHandler() {
	log.Aux("Received INT signal, shutdown...")
	shutdown(0)
}

// TERM signal handler
func termSignalHandler() {
	log.Aux("Received TERM signal, shutdown...")
	shutdown(0)
}

// HUP signal handler
func hupSignalHandler() {
	log.Info("Received HUP signal, log will be reopened...")
	log.Reopen()
	log.Info("Log reopened by HUP signal")
}

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
}

// printError prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{y}"+f+"{!}\n", a...)
}

// printErrorAndExit print error mesage and exit with exit code 1
func printErrorAndExit(f string, a ...interface{}) {
	printError(f, a...)
	os.Exit(1)
}

// shutdown stop deamon
func shutdown(code int) {
	os.Exit(code)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func showUsage() {
	info := usage.NewInfo()

	info.AddOption(OPT_CONFIG, "Path to configuraion file", "file")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VERSION, "Show version")

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2006,
		Owner:   "Gleb Goncharov",
		License: "MIT license",
	}

	about.Render()
}
