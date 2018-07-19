package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/knq/jwt"
	_ "github.com/lib/pq"
	"github.com/matryer/way"
)

var (
	origin    *url.URL
	db        *sql.DB
	sendMail  MailSender
	jwtSigner jwt.Signer
)

func main() {
	godotenv.Load()

	port := intEnv("PORT", 3000)
	originStr := env("ORIGIN", fmt.Sprintf("http://localhost:%d/", port))
	databaseURL := env("DATABASE_URL", "postgresql://root@127.0.0.1:26257/passwordless_demo?sslmode=disable")
	smtpHost := env("SMTP_HOST", "smtp.mailtrap.io")
	smtpPort := intEnv("SMTP_PORT", 25)
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	jwtKey := env("JWT_KEY", "super-duper-secret-key")

	var err error
	if origin, err = url.Parse(originStr); err != nil || !origin.IsAbs() {
		log.Fatalln("invalid origin")
		return
	}

	if i, err := strconv.Atoi(origin.Port()); err == nil {
		port = i
	}

	if smtpUsername == "" || smtpPassword == "" {
		log.Fatalln("remember to set $SMTP_USERNAME and $SMTP_PASSWORD")
		return
	}

	if db, err = sql.Open("postgres", databaseURL); err != nil {
		log.Fatalf("could not open database connection: %v\n", err)
		return
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatalf("could not ping to database: %v\n", err)
		return
	}

	sendMail = newMailSender(smtpHost, smtpPort, smtpUsername, smtpPassword)

	if jwtSigner, err = jwt.HS256.New([]byte(jwtKey)); err != nil {
		log.Fatalf("could not create JWT signer: %v\n", err)
		return
	}

	router := way.NewRouter()
	router.HandleFunc("POST", "/api/users", requireJSON(createUser))
	router.HandleFunc("POST", "/api/passwordless/start", requireJSON(passwordlessStart))
	router.HandleFunc("GET", "/api/passwordless/verify_redirect", passwordlessVerifyRedirect)
	router.HandleFunc("GET", "/api/auth_user", guard(getAuthUser))
	router.Handle("GET", "/...", http.FileServer(SPAFileSystem{http.Dir("static")}))

	log.Printf("accepting connections on port: %d\n", port)
	log.Printf("starting server at %s 🚀\n", origin.String())

	addr := fmt.Sprintf(":%d", port)
	handler := withRecoverer(router)
	if err = http.ListenAndServe(addr, handler); err != nil {
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
