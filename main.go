package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/byte-sat/llum-tools/tools"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/likexian/whois"
)

var addr = flag.String("addr", ":3333", "address to listen on")

//go:generate go run github.com/noonien/codoc/cmd/codoc@latest -out tools_codoc.go -pkg main .

// Get the chat id
func GetCID(cid ChatID) string {
	return string(cid)
}

// Get domain whois
// domain: domain name to check. e.g. example.com
func Whois(domain string) (string, error) {
	return whois.Whois(domain)
}

func main() {
	flag.Parse()

	inj, err := tools.Inject(context.Background, ChatID(""))
	if err != nil {
		log.Fatal(err)
	}

	repo, err := tools.New(inj, GetCID, Whois)
	if err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}))

	tr := &ToolRepo{repo}
	r.Get("/tool_schema", tr.GetToolSchema)
	r.Post("/tool", tr.InvokeTool)

	log.Println("listening on", *addr)
	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatal(err)
	}
}

type ToolRepo struct {
	*tools.Repo
}

func (tr *ToolRepo) GetToolSchema(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(tr.Schema())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

type ChatID string

func (tr *ToolRepo) InvokeTool(w http.ResponseWriter, r *http.Request) {
	var call struct {
		ChatID ChatID         `json:"chat_id"`
		Name   string         `json:"name"`
		Args   map[string]any `json:"arguments"`
	}
	if err := json.NewDecoder(io.TeeReader(r.Body, os.Stdout)).Decode(&call); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	inj, _ := tools.Inject(
		func() context.Context { return ctx },
		call.ChatID,
	)
	out, err := tr.Invoke(inj, call.Name, call.Args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(out)
}
