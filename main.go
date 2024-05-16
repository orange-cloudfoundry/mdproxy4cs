package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

// App -
type App struct {
	iname    string
	lastAddr string
}

var (
	gIName = kingpin.Flag("iname", "Use given interface to discover dchp server, can be set by $MDPROXY4CS_INAME").
		Envar("MDPROXY4CS_INAME").Default("eth0").String()
	gLogLevel = kingpin.Flag("log-level", "Set log level, can be set $MDPROXY4CS_LOG_LEVEL").
			Envar("MDPROXY4CS_LOG_LEVEL").Default("info").String()
	gLogFile = kingpin.Flag("log-file", "Set log output file, - for stdout, can be set $MDPROXY4CS_LOG_FILE").
			Envar("MDPROXY4CS_LOG_FILE").Default("-").String()
	gHTTPListen = kingpin.Flag("http-listen", "Set server listen address, can be set by $MDPROXY4CS_HTTP_LISTEN").
			Envar("MDPROXY4CS_HTTP_LISTEN").Default("169.254.169.254:39724").String()
)

// NewApp -
func NewApp() *App {
	if *gLogFile == "-" {
		log.SetFormatter(&log.TextFormatter{
			ForceColors: true,
		})
		log.SetOutput(os.Stdout)
	} else {
		file, err := os.OpenFile(*gLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("unable to create log file %s: %s", *gLogFile, err)
		}
		log.SetOutput(file)
	}

	lLevel, lErr := log.ParseLevel(*gLogLevel)
	if lErr != nil {
		lLevel = log.ErrorLevel
	}
	log.SetLevel(lLevel)

	return &App{
		iname:    *gIName,
		lastAddr: "0.0.0.0",
	}
}

// ServeHTTP
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("serving request %s...", r.URL.Path)

	var (
		attempt = 0
		res     *http.Response
		err     error
	)

	url := fmt.Sprintf("http://%s%s", a.getVrouterAddress(), r.URL.Path)
	for {
		res, err = http.Get(url)
		if err == nil {
			break
		}
		log.Errorf("could not request %s (attempt %d / 3): %s", url, attempt+1, err)
		if attempt == 2 {
			w.WriteHeader(500)
			return
		}
		attempt++
		time.Sleep(time.Second * 5)
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("cannot read reponse from %s: %s", url, err)
		w.WriteHeader(500)
		return
	}
	log.Debugf("serving response to %s: %s", r.URL.Path, string(content))
	w.WriteHeader(res.StatusCode)
	if _, err := w.Write(content); err != nil {
		log.Errorf("writing content failed: %s", err)
	}
}

func (a *App) getVrouterAddress() string {
	log.Info("discovering dhcp address...")
	ifname, err := net.InterfaceByName(a.iname)
	if err != nil {
		log.Errorf("got unknown interface %s: %s", a.iname, err)
		return a.lastAddr
	}

	client := Client{Iface: ifname}
	ip, err := client.DiscoverServer()
	if err != nil {
		log.Printf("error: unable to discover dhcp server: %s", err)
		return a.lastAddr
	}
	log.Infof("found dhcp address: %s", ip.String())
	return ip.String()
}

func main() {
	kingpin.Version(version.Print("mdproxy4cs"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	app := NewApp()
	log.Infof("listening on %s", *gHTTPListen)
	if err := http.ListenAndServe(*gHTTPListen, app); err != nil {
		panic(err)
	}
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
