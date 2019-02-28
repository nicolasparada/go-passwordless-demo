package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/hako/branca"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/matryer/way"
)

var (
	origin     *url.URL
	db         *sql.DB
	mailSender *MailSender
	codec      *branca.Branca
)

func main() {
	godotenv.Load()

	var (
		port         = intEnv("PORT", 3000)
		originStr    = env("ORIGIN", fmt.Sprintf("http://localhost:%d", port))
		dbURL        = env("DB_URL", "postgresql://root@127.0.0.1:26257/passwordless_demo?sslmode=disable")
		smtpHost     = env("SMTP_HOST", "smtp.mailtrap.io")
		smtpPort     = intEnv("SMTP_PORT", 25)
		smtpUsername = mustEnv("SMTP_USERNAME")
		smtpPassword = mustEnv("SMTP_PASSWORD")
		secretKey    = env("SECRET_KEY", "supersecretkeyyoushouldnotcommit")
	)

	flag.IntVar(&port, "p", port, "Port ($PORT)")
	flag.StringVar(&originStr, "origin", originStr, "Origin ($ORIGIN)")
	flag.StringVar(&dbURL, "db", dbURL, "Database URL ($DB_URL)")
	flag.StringVar(&smtpHost, "smtp.host", smtpHost, "SMTP Host ($SMTP_HOST)")
	flag.IntVar(&smtpPort, "smtp.port", smtpPort, "SMTP Port ($SMTP_PORT)")
	flag.Parse()

	var err error
	if origin, err = url.Parse(originStr); err != nil || !origin.IsAbs() {
		log.Fatalln("invalid origin")
		return
	}

	if i, err := strconv.Atoi(origin.Port()); err == nil {
		port = i
	}

	if db, err = sql.Open("postgres", dbURL); err != nil {
		log.Fatalf("could not open database connection: %v\n", err)
		return
	}

	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("could not ping to database: %v\n", err)
		return
	}

	mailSender = newMailSender(smtpHost, smtpPort, smtpUsername, smtpPassword)

	codec = branca.NewBranca(secretKey)
	codec.SetTTL(uint32(tokenLifespan.Seconds()))

	// Weird bug that maybe only happens on my machine.
	mime.AddExtensionType(".js", "application/javascript; charset=utf-8")

	r := way.NewRouter()
	r.HandleFunc("POST", "/api/users", requireJSON(createUser))
	r.HandleFunc("POST", "/api/passwordless/start", requireJSON(passwordlessStart))
	r.HandleFunc("GET", "/api/passwordless/verify_redirect", passwordlessVerifyRedirect)
	r.HandleFunc("GET", "/api/auth_user", guard(getAuthUser))
	r.Handle("GET", "/...", http.FileServer(SPAFileSystem{http.Dir("static")}))

	log.Printf("accepting connections on port: %d\n", port)
	log.Printf("starting server at %s 🚀\n", origin.String())

	addr := fmt.Sprintf(":%d", port)
	h := withRecoverer(r)
	if err = http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}

func env(key, fallbackValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	return v
}

func mustEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("%s required on environment variables", key))
	}
	return v
}

func intEnv(key string, fallbackValue int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallbackValue
	}
	return i
}
