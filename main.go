package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/matryer/way"
)

var (
	config struct {
		domain url.URL
		jwtKey []byte
	}
	db     *sql.DB
	sendMail MailSender
)

func main() {
	// Environment variables
	port, err := strconv.Atoi(env("PORT", "3000"))
	if err != nil {
		log.Fatalf("could not parse port: %v\n", err)
		return
	}
	domain := env("APP_URL", fmt.Sprintf("http://localhost:%d/", port))
	databaseURL := env("DATABASE_URL", "postgresql://root@127.0.0.1:26257/passwordless_demo?sslmode=disable")
	smtpHost := env("SMTP_HOST", "smtp.mailtrap.io")
	smtpPort, err := strconv.Atoi(env("SMTP_PORT", "25"))
	if err != nil {
		log.Printf("could not parse SMTP port: %v", err)
	}
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	// CLI flags
	flag.Usage = func() {
		fmt.Println("Don't forget to set $JWT_KEY. SMTP username and password are required.")
		flag.PrintDefaults()
	}
	flag.IntVar(&port, "port", port, "Port ($PORT)")
	flag.StringVar(&domain, "domain", domain, "Domain ($APP_URL)")
	flag.StringVar(&databaseURL, "db", databaseURL, "Database address ($DATABASE_URL)")
	flag.StringVar(&smtpHost, "smtphost", smtpHost, "SMTP host ($SMTP_HOST)")
	flag.IntVar(&smtpPort, "smtpport", smtpPort, "SMTP port ($SMTP_PORT)")
	flag.StringVar(&smtpUsername, "smtpuser", smtpUsername, "SMTP username ($SMTP_USERNAME)")
	flag.StringVar(&smtpPassword, "smtppwd", smtpPassword, "SMTP password ($SMTP_PASSWORD)")
	flag.Parse()

	// Config
	if u, err := url.Parse(domain); err != nil {
		log.Fatalf("could not parse domain URL: %v\n", err)
		return
	} else if !u.IsAbs() {
		log.Fatalln("could not use a non absolute domain URL")
		return
	} else {
		config.domain = *u
	}
	if smtpUsername == "" {
		log.Fatalln("could not find SMTP_USERNAME on environment variables")
		return
	}
	if smtpPassword == "" {
		log.Fatalln("could not find SMTP_PASSWORD on environment variables")
		return
	}
	config.jwtKey = []byte(env("JWT_KEY", "super-duper-secret-key"))

	// Database
	if db, err = sql.Open("postgres", databaseURL); err != nil {
		log.Fatalf("could not open database connection: %v\n", err)
		return
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatalf("could not ping to database: %v\n", err)
		return
	}

	// Mailing
	sendMail = newMailSender(smtpHost, smtpPort, smtpUsername, smtpPassword)

	// Routing
	router := way.NewRouter()
	router.HandleFunc("POST", "/api/users", requireJSON(createUser))
	router.HandleFunc("POST", "/api/passwordless/start", requireJSON(passwordlessStart))
	router.HandleFunc("GET", "/api/passwordless/verify_redirect", passwordlessVerifyRedirect)
	router.HandleFunc("GET", "/api/auth_user", guard(getAuthUser))
	router.Handle("GET", "/...", http.FileServer(SPAFileSystem{http.Dir("static")}))

	// Server
	log.Printf("starting server at %s 🚀\n", config.domain.String())
	log.Fatalf("could not start server: %v\n", http.ListenAndServe(fmt.Sprintf(":%d", port), withRecover(router)))
}

func env(key, fallbackValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	return v
}
