package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/byte-sat/llum-tools/tools"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/likexian/whois"
)

var addr = flag.String("addr", ":3333", "address to listen on")

//go:generate go run github.com/noonien/codoc/cmd/codoc@latest -out tools_codoc.go -pkg main .

// adds two numbers
// a: the first number
// b: the second number
func add(a int, b int) int {
	return a + b
}

type Foo struct {
	A int `json:"a"` // foo
	B int
}

// woops the foo
// f: foo
// y: yeets
// z: zzz
func woop(f Foo, y map[string]int, z []int) int {
	return f.A + f.B
}

// Get domain whois
// domain: domain name to check. e.g. example.com
func Whois(domain string) (string, error) {
	return whois.Whois(domain)
}

var toolz tools.Group

func main() {
	flag.Parse()

	toolz = tools.Group{
		Name:        "math",
		Description: "Mathematical tools",
		Tools: []tools.Tool{
			tools.CodocFunc(add),
			tools.CodocFunc(woop),
		},
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

	tr := &ToolRepo{}
	r.Get("/tool_schema", tr.GetToolSchema)
	r.Post("/tool", tr.InvokeTool)

	log.Println("listening on", *addr)
	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatal(err)
	}
}

type ToolRepo struct {
}

func (tr *ToolRepo) GetToolSchema(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(toolz)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (tr *ToolRepo) InvokeTool(w http.ResponseWriter, r *http.Request) {
	var call struct {
		Name string         `json:"name"`
		Args map[string]any `json:"arguments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	out, err := toolz.Invoke(r.Context(), call.Name, call.Args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(out)
}
