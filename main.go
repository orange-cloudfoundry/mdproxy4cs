package main

import (
	"fmt"
	"github.com/orange-cloudfoundry/mdproxy4cs/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
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
	gIName      = kingpin.Flag("iname", "Use given interface to discover dchp server").Default("eth0").String()
	gLogLevel   = kingpin.Flag("log-level", "Set log level").Default("info").String()
	gHttpListen = kingpin.Flag("http-listen", "Set server listen address").Default("169.254.169.254:80").String()
)

// NewApp -
func NewApp() *App {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stdout)
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

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("cannot read reponse from %s: %s", url, err)
		w.WriteHeader(500)
		return
	}
	log.Infof("serving response to %s: %s", r.URL.Path, string(content))
	w.Write(content)
	w.WriteHeader(res.StatusCode)
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
	kingpin.Version(version.PrintVersion("mdproxy4cs"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	app := NewApp()
	if err := http.ListenAndServe(*gHttpListen, app); err != nil {
		panic(err)
	}
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
