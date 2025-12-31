package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/ttrubel/send-me-home/gen/game/v1/gamev1connect"
	"github.com/ttrubel/send-me-home/internal/api"
	"github.com/ttrubel/send-me-home/internal/config"
	"github.com/ttrubel/send-me-home/internal/services/firestore"
	"github.com/ttrubel/send-me-home/internal/services/gemini"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize services
	// Gemini client reads config from environment variables:
	// - GOOGLE_GENAI_USE_VERTEXAI=true for Vertex AI
	// - GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION for Vertex AI
	// - GOOGLE_API_KEY for AI Studio
	geminiClient := gemini.NewClient()

	firestoreClient, err := firestore.NewClient(cfg.FirestoreProjectID)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}

	// Initialize handler
	gameHandler := api.NewGameHandler(geminiClient, firestoreClient)

	// Create Connect-RPC service
	mux := http.NewServeMux()

	// Register game service
	path, handler := gamev1connect.NewGameServiceHandler(
		gameHandler,
		connect.WithInterceptors(loggingInterceptor()),
	)
	mux.Handle(path, handler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:5173",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Connect-Protocol-Version",
			"Connect-Timeout-Ms",
		},
		ExposedHeaders: []string{
			"Connect-Protocol-Version",
		},
	}).Handler(mux)

	// Use h2c for HTTP/2 without TLS (development)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: h2c.NewHandler(corsHandler, &http2.Server{}),
	}

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Game service available at http://localhost:%s%s", cfg.Port, path)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// loggingInterceptor logs all RPC calls
func loggingInterceptor() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			log.Printf("RPC: %s", req.Spec().Procedure)
			return next(ctx, req)
		})
	})
}
