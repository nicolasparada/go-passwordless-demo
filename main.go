package main

import (
	"database/sql"
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
	db     *sql.DB
	config struct {
		domain       url.URL
		port         int
		databaseURL  string
		jwtKey       []byte
		smtpHost     string
		smtpPort     int
		smtpUsername string
		smtpPassword string
	}
)

func init() {
	config.port, _ = strconv.Atoi(env("PORT", "3000"))
	u, _ := url.Parse(env("APP_URL", "http://localhost:"+strconv.Itoa(config.port)+"/"))
	config.domain = *u
	config.databaseURL = env("DATABASE_URL", "postgresql://root@127.0.0.1:26257/passwordless_demo?sslmode=disable")
	config.jwtKey = []byte(env("JWT_KEY", "super-duper-secret-key"))
	config.smtpHost = env("SMTP_HOST", "smtp.mailtrap.io")
	config.smtpPort, _ = strconv.Atoi(env("SMTP_PORT", "25"))
	var ok bool
	config.smtpUsername, ok = os.LookupEnv("SMTP_USERNAME")
	if !ok {
		log.Fatalln("could not find SMTP_USERNAME on environment variables")
	}
	config.smtpPassword, ok = os.LookupEnv("SMTP_PASSWORD")
	if !ok {
		log.Fatalln("could not find SMTP_PASSWORD on environment variables")
	}
}

func main() {
	var err error
	if db, err = sql.Open("postgres", config.databaseURL); err != nil {
		log.Fatalf("could not open database connection: %v\n", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatalf("could not ping to database: %v\n", err)
	}

	initMailing()

	router := way.NewRouter()
	router.HandleFunc("POST", "/api/users", requireJSON(createUser))
	router.HandleFunc("POST", "/api/passwordless/start", requireJSON(passwordlessStart))
	router.HandleFunc("GET", "/api/passwordless/verify_redirect", passwordlessVerifyRedirect)
	router.HandleFunc("GET", "/api/auth_user", guard(getAuthUser, nil))
	router.Handle("GET", "/...", http.FileServer(SPAFileSystem{http.Dir("static")}))

	addr := fmt.Sprintf(":%d", config.port)
	log.Printf("starting server at %s 🚀\n", config.domain.String())
	log.Fatalf("could not start server: %v\n", http.ListenAndServe(addr, withRecover(router)))
}

func env(key, fallbackValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	return v
}
