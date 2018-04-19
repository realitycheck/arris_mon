// Arris Monitor Server 2.0
// 2018
package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	htmlquery "github.com/antchfx/xquery/html"
	"golang.org/x/net/html"

	statik "github.com/rakyll/statik/fs"
	_ "github.com/realitycheck/arris_mon/statik" // register res files as embedded static.
)

var (
	addr                          = "127.0.0.1:8081"
	sourceURL                     = "http://192.168.100.1/cgi-bin/status_cgi"
	downstreamTable               = "//table[2]/tbody"
	upstreamTable                 = "//table[4]/tbody"
	delay           time.Duration = 1
	verbosity                     = false

	downstreamFreq           = newGauge("arris_downstream_freq", "Downstream Frequency")
	downstreamPower          = newGauge("arris_downstream_power", "Downstream Power")
	downstreamSnr            = newGauge("arris_downstream_snr", "Downstream SNR")
	downstreamOctets         = newGauge("arris_downstream_octets", "Downstream Octets")
	downstreamCorrecteds     = newGauge("arris_downstream_correcteds", "Downstream Correcteds")
	downstreamUncorrectables = newGauge("arris_downstream_uncorrectables", "Downstream Uncorrectables")

	upstreamFreq  = newGauge("arris_upstream_freq", "Upstream Frequency")
	upstreamPower = newGauge("arris_upstream_power", "Upstream Power")
)

func logVerbose(format string, v ...interface{}) {
	if verbosity {
		log.Printf(format, v)
	}
}

func newGauge(name, help string) *prom.GaugeVec {
	return prom.NewGaugeVec(prom.GaugeOpts{
		Name: name,
		Help: help,
	}, []string{
		"id", "name",
	})
}

func init() {
	flag.StringVar(&addr, "addr", addr, "Server address")
	flag.StringVar(&sourceURL, "src", sourceURL, "Source URL")
	flag.BoolVar(&verbosity, "v", verbosity, "Verbose print")

	prom.MustRegister(downstreamFreq)
	prom.MustRegister(downstreamPower)
	prom.MustRegister(downstreamSnr)
	prom.MustRegister(downstreamOctets)
	prom.MustRegister(downstreamCorrecteds)
	prom.MustRegister(downstreamUncorrectables)

	prom.MustRegister(upstreamFreq)
	prom.MustRegister(upstreamPower)
}

func main() {
	flag.Parse()

	// Background worker
	go func() {
		for i := 1; ; i++ {
			select {
			case <-time.After(delay * time.Second):
				log.Printf("%v: #%d\n", time.Now(), i)
				if err := pull(sourceURL, downstreamTable, upstreamTable); err != nil {
					log.Printf("%v: %v\n", time.Now(), err)
				}
			}
		}
	}()

	// Static files
	fs, err := statik.New()
	if err != nil {
		log.Panicf("arris_mon: %s", err)
	}

	// App template
	file, err := fs.Open("/templates/arris_mon.html")
	if err != nil {
		log.Panicf("arris_mon: %s", err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panicf("arris_mon: %s", err)
	}

	tmpl, err := template.New("index.html").Parse(string(bytes))
	if err != nil {
		log.Panicf("arris_mon: %s", err)
	}

	// App handler
	app := func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
			return
		}
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tmpl.Execute(w, nil)
	}

	http.Handle("/", http.HandlerFunc(app))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/static/", http.FileServer(fs))

	log.Fatal(http.ListenAndServe(addr, nil))
}

func pull(url, downstreamTable, upstreamTable string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	downstream := parseTable(doc, downstreamTable)
	logVerbose("Downstream: %v\n", downstream)

	upstream := parseTable(doc, upstreamTable)
	logVerbose("Upstream: %v\n", upstream)

	setGauge := func(g *prom.GaugeVec, id string, name string, value string, format string) {
		var val float64
		fmt.Sscanf(value, format, &val)
		g.WithLabelValues(id, name).Set(val)
	}

	it := downstream.iterator()
	for dc := it(); dc != nil; dc = it() {
		id, name := dc["DCID"], dc[""]
		setGauge(downstreamFreq, id, name, dc["Freq"], "%f MHz")
		setGauge(downstreamPower, id, name, dc["Power"], "%f dBmV")
		setGauge(downstreamSnr, id, name, dc["SNR"], "%f dB")
		setGauge(downstreamOctets, id, name, dc["Octets"], "%f")
		setGauge(downstreamCorrecteds, id, name, dc["Correcteds"], "%f")
		setGauge(downstreamUncorrectables, id, name, dc["Uncorrectables"], "%f")
	}

	it = upstream.iterator()
	for uc := it(); uc != nil; uc = it() {
		id, name := uc["UCID"], uc[""]
		setGauge(upstreamFreq, id, name, uc["Freq"], "%f MHz")
		setGauge(upstreamPower, id, name, uc["Power"], "%f dBmV")
	}

	return nil
}

type table [][]string
type channel map[string]string

func parseTable(doc *html.Node, expr string) table {
	tr := htmlquery.Find(htmlquery.FindOne(doc, expr), "/tr[td]")
	t := make(table, len(tr))
	for i, r := range tr {
		td := htmlquery.Find(r, "/td")
		t[i] = make([]string, len(td))
		for j, d := range td {
			t[i][j] = htmlquery.InnerText(d)
		}
	}
	return t
}

func (t table) iterator() func() channel {
	i, n := 0, len(t)
	return func() channel {
		if i++; i >= n {
			return nil
		}

		keys, vals := t[0], t[i]
		if len(keys) != len(vals) {
			log.Panic("arris_mon: lengthes of keys and values are mismatched")
		}

		res := make(channel, len(keys))
		for j, val := range vals {
			res[keys[j]] = val
		}
		return res
	}
}
