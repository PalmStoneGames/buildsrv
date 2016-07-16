package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/caddyserver/buildsrv/features"
	"github.com/caddyserver/buildsrv/server"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// Get GOPATH and path of caddy project
	cmd := exec.Command("go", "env", "GOPATH")
	result, err := cmd.Output()
	if err != nil {
		log.Fatal("Cannot locate GOPATH:", err)
	}
	server.CaddyPath = strings.TrimSpace(string(result)) + "/src/" + server.MainCaddyPackage

	// Log to a file
	outfile, err := os.OpenFile("builds.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Couldn't open log file:", err)
	}
	log.SetOutput(outfile)
}

func main() {
	go func() {
		// Delete existing builds on quit
		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt, os.Kill)
		<-interrupt

		err := os.RemoveAll(server.BuildPath)
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()
	http.HandleFunc("/download/build", server.BuildHandler)
	http.Handle("/download/builds/", http.StripPrefix("/download/builds/", http.FileServer(http.Dir(server.BuildPath))))
	http.HandleFunc("/features.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		var plugins features.Plugins
		for _, plugin := range features.Registry {
			if plugin.Name != "" {
				plugins = append(plugins, plugin)
			}
		}
		json.NewEncoder(w).Encode(plugins)
	})
	http.ListenAndServe(":5050", nil)
}
