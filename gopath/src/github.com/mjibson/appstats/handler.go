/*
 * Copyright (c) 2013 Matt Jibson <matt.jibson@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package appstats

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"appengine"
	"appengine/memcache"
)

var templates *template.Template
var staticFiles map[string][]byte

func init() {
	templates = template.New("appstats").Funcs(funcs)
	templates.Parse(htmlBase)
	templates.Parse(htmlMain)
	templates.Parse(htmlDetails)
	templates.Parse(htmlFile)

	staticFiles = map[string][]byte{
		"app_engine_logo_sm.gif": app_engine_logo_sm_gif,
		"appstats_css.css":       appstats_css_css,
		"appstats_js.js":         appstats_js_js,
		"gantt.js":               gantt_js,
		"minus.gif":              minus_gif,
		"pix.gif":                pix_gif,
		"plus.gif":               plus_gif,
	}
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func AppstatsHandler(w http.ResponseWriter, r *http.Request) {
	if detailsURL == r.URL.Path {
		Details(w, r)
	} else if fileURL == r.URL.Path {
		File(w, r)
	} else if strings.HasPrefix(r.URL.Path, staticURL) {
		Static(w, r)
	} else {
		Index(w, r)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	keys := make([]string, modulus)
	for i := range keys {
		keys[i] = fmt.Sprintf(keyPart, i*distance)
	}

	c := context(r)
	items, err := memcache.GetMulti(c, keys)
	if err != nil {
		return
	}

	ars := AllRequestStats{}
	for _, v := range items {
		t := stats_part{}
		err := gob.NewDecoder(bytes.NewBuffer(v.Value)).Decode(&t)
		if err != nil {
			continue
		}
		r := RequestStats(t)
		ars = append(ars, &r)
	}
	sort.Sort(reverse{ars})

	requestById := make(map[int]*RequestStats, len(ars))
	idByRequest := make(map[*RequestStats]int, len(ars))
	requests := make(map[int]*StatByName)
	byRequest := make(map[int]map[string]cVal)
	for i, v := range ars {
		idx := i + 1
		requestById[idx] = v
		idByRequest[v] = idx
		requests[idx] = &StatByName{
			RequestStats: v,
		}
		byRequest[idx] = make(map[string]cVal)
	}

	requestByPath := make(map[string][]int)
	byCount := make(map[string]cVal)
	byRPC := make(map[SKey]cVal)
	for _, t := range ars {
		id := idByRequest[t]

		requestByPath[t.Path] = append(requestByPath[t.Path], id)

		for _, r := range t.RPCStats {
			rpc := r.Name()

			v := byRequest[id][rpc]
			v.count++
			v.cost += r.Cost
			byRequest[id][rpc] = v

			v = byCount[rpc]
			v.count++
			v.cost += r.Cost
			byCount[rpc] = v

			v = byRPC[SKey{rpc, t.Path}]
			v.count++
			v.cost += r.Cost
			byRPC[SKey{rpc, t.Path}] = v
		}
	}

	for k, v := range byRequest {
		stats := StatsByName{}
		for rpc, s := range v {
			stats = append(stats, &StatByName{
				Name:  rpc,
				Count: s.count,
				Cost:  s.cost,
			})
		}
		sort.Sort(reverse{stats})
		requests[k].SubStats = stats
	}

	statsByRPC := make(map[string]StatsByName)
	pathStats := make(map[string]StatsByName)
	for k, v := range byRPC {
		statsByRPC[k.a] = append(statsByRPC[k.a], &StatByName{
			Name:  k.b,
			Count: v.count,
			Cost:  v.cost,
		})
		pathStats[k.b] = append(pathStats[k.b], &StatByName{
			Name:  k.a,
			Count: v.count,
			Cost:  v.cost,
		})
	}
	for k, v := range statsByRPC {
		sort.Sort(reverse{v})
		statsByRPC[k] = v
	}

	pathStatsByCount := StatsByName{}
	for k, v := range pathStats {
		total := 0
		var cost int64
		for _, stat := range v {
			total += stat.Count
			cost += stat.Cost
		}
		sort.Sort(reverse{v})

		pathStatsByCount = append(pathStatsByCount, &StatByName{
			Name:       k,
			Count:      total,
			Cost:       cost,
			SubStats:   v,
			Requests:   len(requestByPath[k]),
			RecentReqs: requestByPath[k],
		})
	}
	sort.Sort(reverse{pathStatsByCount})

	allStatsByCount := StatsByName{}
	for k, v := range byCount {
		allStatsByCount = append(allStatsByCount, &StatByName{
			Name:     k,
			Count:    v.count,
			Cost:     v.cost,
			SubStats: statsByRPC[k],
		})
	}
	sort.Sort(reverse{allStatsByCount})

	v := struct {
		Env                 map[string]string
		Requests            map[int]*StatByName
		RequestStatsByCount map[int]*StatByName
		AllStatsByCount     StatsByName
		PathStatsByCount    StatsByName
	}{
		Env: map[string]string{
			"APPLICATION_ID": appengine.AppID(c),
		},
		Requests:         requests,
		AllStatsByCount:  allStatsByCount,
		PathStatsByCount: pathStatsByCount,
	}

	_ = templates.ExecuteTemplate(w, "main", v)
}

func Details(w http.ResponseWriter, r *http.Request) {
	qtime := r.URL.Query().Get("time")
	key := fmt.Sprintf(keyFull, qtime)

	c := context(r)

	v := struct {
		Env             map[string]string
		Record          *RequestStats
		Header          http.Header
		AllStatsByCount StatsByName
		Real            time.Duration
	}{
		Env: map[string]string{
			"APPLICATION_ID": appengine.AppID(c),
		},
	}

	item, err := memcache.Get(c, key)
	if err != nil {
		templates.ExecuteTemplate(w, "details", v)
		return
	}

	full := stats_full{}
	err = gob.NewDecoder(bytes.NewBuffer(item.Value)).Decode(&full)
	if err != nil {
		templates.ExecuteTemplate(w, "details", v)
		return
	}

	byCount := make(map[string]cVal)
	durationCount := make(map[string]time.Duration)
	var _real time.Duration
	for _, r := range full.Stats.RPCStats {
		rpc := r.Name()

		// byCount
		if _, present := byCount[rpc]; !present {
			durationCount[rpc] = 0
		}
		v := byCount[rpc]
		v.count++
		v.cost += r.Cost
		byCount[rpc] = v
		durationCount[rpc] += r.Duration
		_real += r.Duration
	}

	allStatsByCount := StatsByName{}
	for k, v := range byCount {
		allStatsByCount = append(allStatsByCount, &StatByName{
			Name:     k,
			Count:    v.count,
			Cost:     v.cost,
			Duration: durationCount[k],
		})
	}
	sort.Sort(allStatsByCount)

	v.Record = full.Stats
	v.Header = full.Header
	v.AllStatsByCount = allStatsByCount
	v.Real = _real

	_ = templates.ExecuteTemplate(w, "details", v)
}

func File(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Query().Get("f")
	n := r.URL.Query().Get("n")
	lineno, _ := strconv.Atoi(n)
	c := context(r)

	f, err := ioutil.ReadFile(fname)
	if err != nil {
		serveError(w, err)
		return
	}

	fp := make(map[int]string)
	for k, v := range strings.Split(string(f), "\n") {
		fp[k+1] = v
	}

	v := struct {
		Env      map[string]string
		Filename string
		Lineno   int
		Fp       map[int]string
	}{
		Env: map[string]string{
			"APPLICATION_ID": appengine.AppID(c),
		},
		Filename: fname,
		Lineno:   lineno,
		Fp:       fp,
	}

	_ = templates.ExecuteTemplate(w, "file", v)
}

func Static(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	if v, present := staticFiles[fname]; present {
		h := w.Header()

		if strings.HasSuffix(r.URL.Path, ".css") {
			h.Set("Content-type", "text/css")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			h.Set("Content-type", "text/javascript")
		}

		h.Set("Cache-Control", "public, max-age=expiry")
		expires := time.Now().Add(time.Hour)
		h.Set("Expires", expires.Format(time.RFC1123))

		w.Write(v)
	}
}
