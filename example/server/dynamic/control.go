package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/zitadel/oidc/v3/example/server/storage"
)

type control struct {
	router chi.Router
}

func NewControl() *control {
	c := &control{}
	c.createRouter()
	return c
}

func (c *control) createRouter() {
	c.router = chi.NewRouter()
	c.router.Route("/client/{clientId}", func(r chi.Router) {
		r.Use(c.clientCtx)                                          // Load the *storage.Client on the request context
		r.Patch("/idTokenLifetime", c.updateIdTokenLifetimeHandler) // PATCH /client/123/idTokenLifetime
	})
}

func (c *control) clientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientId := chi.URLParam(r, "clientId")
		client, ok := storage.Clients[clientId]
		if !ok {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "client", client)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *control) updateIdTokenLifetimeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client, ok := ctx.Value("client").(*storage.Client)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	// Read new lifetime from body of request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	// Parse duration
	newLifeTime, err := time.ParseDuration(string(body))
	if err != nil {
		log.Printf("Error parsing body (%s) as time.Duration: %v", string(body), err)
		http.Error(w, "can't parse body as time.Duration", http.StatusBadRequest)
		return
	}

	// Update this client's id token lifetime
	client.SetIDTokenLifetime(newLifeTime)

	w.Write([]byte(fmt.Sprintf("%v", newLifeTime)))
}
