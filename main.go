package main

import (
	"crypto/tls"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/wbergg/asterisk-extention-manager/internal/auth"
	"github.com/wbergg/asterisk-extention-manager/internal/config"
	"github.com/wbergg/asterisk-extention-manager/internal/database"
	"github.com/wbergg/asterisk-extention-manager/internal/handlers"
	"github.com/wbergg/asterisk-extention-manager/internal/middleware"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := database.SeedAdmin(db, cfg.AdminUser, cfg.AdminPass); err != nil {
		log.Fatalf("Failed to seed admin: %v", err)
	}

	authHandler := &handlers.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	userHandler := &handlers.UserHandler{DB: db}
	extHandler := &handlers.ExtensionHandler{DB: db, Config: cfg}
	cdrHandler := &handlers.CDRHandler{Config: cfg}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.CORS)

	// Public
	r.Post("/api/login", authHandler.Login)
	r.Get("/directory.xml", extHandler.Directory)

	// Authenticated
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(cfg.JWTSecret))

		r.Get("/api/me", authHandler.Me)
		r.Get("/api/directory", extHandler.DirectoryJSON)
		r.Put("/api/me/password", authHandler.ChangePassword)
		r.Get("/api/extensions", extHandler.List)
		r.Post("/api/extensions", extHandler.Create)
		r.Get("/api/extensions/{ext}", extHandler.Get)
		r.Put("/api/extensions/{ext}", extHandler.Update)
		r.Delete("/api/extensions/{ext}", extHandler.Delete)

		// Call Log (accessible to users with call_log_access)
		r.Group(func(r chi.Router) {
			r.Use(auth.CallLogAccess)

			r.Get("/api/cdr", cdrHandler.ListCDR)
			r.Get("/api/cdr/stats", cdrHandler.Stats)
		})

		// Admin
		r.Group(func(r chi.Router) {
			r.Use(auth.AdminOnly)

			r.Get("/api/admin/users", userHandler.List)
			r.Post("/api/admin/users", userHandler.Create)
			r.Put("/api/admin/users/{id}", userHandler.Update)
			r.Delete("/api/admin/users/{id}", userHandler.Delete)
			r.Get("/api/admin/extensions", extHandler.ListAll)
			r.Put("/api/admin/extensions/{ext}", extHandler.AdminUpdate)
			r.Delete("/api/admin/extensions/{ext}", extHandler.AdminDelete)
			r.Post("/api/admin/sync", extHandler.ForceSync)
			r.Post("/api/admin/impersonate/{id}", authHandler.Impersonate)
			r.Get("/api/admin/blocked", extHandler.ListBlocked)
			r.Post("/api/admin/blocked", extHandler.BlockExtension)
			r.Delete("/api/admin/blocked/{ext}", extHandler.UnblockExtension)
		})
	})

	// Serve embedded frontend
	frontendDist, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to get frontend sub-filesystem: %v", err)
	}
	// Serve logo.bmp with correct headers for Cisco phones
	r.Get("/logo.bmp", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(frontendDist, "logo.bmp")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/bmp")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Write(data)
	})

	// Known frontend routes that should get SPA fallback (serve index.html)
	spaRoutes := map[string]bool{
		"login":      true,
		"extensions": true,
		"directory":  true,
		"settings":   true,
		"call-log":   true,
		"admin":      true,
	}

	fileServer := http.FileServer(http.FS(frontendDist))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(frontendDist, path); err != nil {
			// Only serve SPA fallback for known frontend routes
			if spaRoutes[path] {
				r.URL.Path = "/"
			} else {
				http.NotFound(w, r)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	// HTTP handler for port 80: serve Cisco phone endpoints, redirect the rest
	httpHandler := func(fallback http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/directory.xml":
				extHandler.Directory(w, req)
			case "/logo.bmp":
				data, err := fs.ReadFile(frontendDist, "logo.bmp")
				if err != nil {
					http.NotFound(w, req)
					return
				}
				w.Header().Set("Content-Type", "image/bmp")
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
				w.Write(data)
			default:
				if fallback != nil {
					fallback.ServeHTTP(w, req)
					return
				}
				http.Redirect(w, req, "https://"+req.Host+req.RequestURI, http.StatusMovedPermanently)
			}
		})
	}

	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		// HTTPS with existing certificate
		go func() {
			log.Printf("Starting HTTP on :80 (directory.xml, logo.bmp + redirect)")
			srv := &http.Server{
				Addr:    ":80",
				Handler: httpHandler(nil),
			}
			if err := srv.ListenAndServe(); err != nil {
				log.Printf("HTTP server: %v", err)
			}
		}()

		log.Printf("Starting HTTPS server with certificate %s", cfg.TLSCertFile)
		srv := &http.Server{
			Addr:    ":443",
			Handler: r,
		}
		if err := srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	} else if cfg.TLSDomain != "" {
		// HTTPS with Let's Encrypt
		m := &autocert.Manager{
			Cache:      autocert.DirCache(cfg.TLSCertDir),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.TLSDomain),
		}

		go func() {
			log.Printf("Starting HTTP on :80 (directory.xml, logo.bmp + ACME + redirect)")
			srv := &http.Server{
				Addr:    ":80",
				Handler: httpHandler(m.HTTPHandler(nil)),
			}
			if err := srv.ListenAndServe(); err != nil {
				log.Printf("HTTP server: %v", err)
			}
		}()

		srv := &http.Server{
			Addr:    ":443",
			Handler: r,
			TLSConfig: &tls.Config{
				GetCertificate: m.GetCertificate,
			},
		}
		log.Printf("Starting HTTPS server for %s (Let's Encrypt)", cfg.TLSDomain)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	} else {
		// Plain HTTP
		log.Printf("Starting server on %s", cfg.ListenAddr)
		if err := http.ListenAndServe(cfg.ListenAddr, r); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}
