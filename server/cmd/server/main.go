package main

import (
    "net/http"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize storage (in-memory for now)
    repo := memory.NewSnapshotRepository()
    objectStore := memory.NewObjectStore()

    // Create services
    svc := service.New(repo, objectStore)

    // Setup HTTP server
    handler := api.New(svc)
    router := api.NewRouter(handler)

    log.Printf("Starting server on :%d", cfg.HTTPPort)
    if err := http.ListenAndServe(":"+strconv.Itoa(cfg.HTTPPort), router); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
