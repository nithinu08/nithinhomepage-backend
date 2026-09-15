package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Links struct {
	Live string `json:"live,omitempty"`
	Code string `json:"code,omitempty"`
}

type Project struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	LongDescription string `json:"longDescription"`
	Highlights  []string `json:"highlights,omitempty"`
	TechStack   []string `json:"techStack"`
	Links       Links    `json:"links"`
}

var projects = []Project{
	{
		Title:       "Linkding Bookmark Application",
		Description: "A Bookmark application: Image taken from https://github.com/sissbruecker repo.",
		LongDescription: "The first application I hosted on the homelab, proof that Kubernetes concepts from tutorials could hold up against something I actually had to keep running.",
		Highlights: []string{
				"Wrote the Docker image into Kustomize manifests",
				"Deployed via a self managed Cloudflare tunnel",
				"Encrypted secrets with SOPS+age",
				"Automated image updates with Renovate",
		},
		TechStack:   []string{"Docker", "Kubernetes", "Renovate"},
		Links: Links{
			Live: "https://linkding.nithinhomepage.com",
			Code: "https://github.com/nithinu08/pi-cluster/tree/master/",
		},
	},
	{
		Title:       "AudioBookShelf Application",
		Description: "An application used to store Audiobooks, Podcasts: Image taken from https://github.com/advplyr/audiobookshelf",
		LongDescription: "A second homelab deployment, this time with a clear understanding of the full path from manifest to running pod.",
		Highlights: []string{
				"Reused the Linkding deployment pattern end-to-end",
				"Configured persistent storage for media libraries",
				"Accessible from any device, anywhere, via Cloudflare Tunnel",
		},
		TechStack:   []string{"Docker", "Kubernetes"},
		Links: Links{
			Live: "https://abs.nithinhomepage.com",
			Code: "https://github.com/nithinu08/pi-cluster/tree/master/",
		},
	},
	{
		Title:       "Grafana Dashboard",
		Description: "Internal monitoring setup for the homelab, accessible only within local network. Not publicly reachable.",
		LongDescription: "Internal-only monitoring for the Pi cluster, not publicly reachable, and not meant to be.",
		Highlights: []string{
				"Learned how Prometheus metrics are scraped and stored",
				"Built dashboards for cluster health and resource usage",
				"Reachable from any device on the local network",
		},
		TechStack:   []string{"Grafana"},
		Links:       Links{},
	},
}

func initTracer() *sdktrace.TracerProvider {
	endpoint := os.Getenv("OTEL_EXPORTER_ENDPOINT")
	if endpoint == "" {
		  endpoint = "localhost:4317"
	}

	exporter, err := otlptracegrpc.New(context.Background(),
      otlptracegrpc.WithEndpoint("signoz-otel-collector.signoz.svc.cluster.local:4317"),
      otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("nithinhomepage-backend"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(projects)
}

func main() {
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	handler := otelhttp.NewHandler(http.HandlerFunc(projectsHandler), "projects-handler")
	http.Handle("/api/projects", handler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
