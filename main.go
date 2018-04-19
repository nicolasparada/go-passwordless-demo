package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/matryer/way"
)

// ContextKey used on app middlewares.
type ContextKey int

var (
	db     *sql.DB
	config struct {
		port        int
		appURL      *url.URL
		databaseURL string
		jwtKey      []byte
		smtpAddr    string
		smtpAuth    smtp.Auth
	}
)

func init() {
	config.port, _ = strconv.Atoi(env("PORT", "80"))
	config.appURL, _ = url.Parse(env("APP_URL", "http://localhost:"+strconv.Itoa(config.port)+"/"))
	config.databaseURL = env(
		"DATABASE_URL",
		"postgresql://root@127.0.0.1:26257/passwordless_demo?sslmode=disable")
	config.jwtKey = []byte(env("JWT_KEY", "super-duper-secret-key"))
	smtpHost := env("SMTP_HOST", "smtp.mailtrap.io")
	config.smtpAddr = net.JoinHostPort(smtpHost, env("SMTP_PORT", "25"))
	smtpUsername, ok := os.LookupEnv("SMTP_USERNAME")
	if !ok {
		log.Fatalln("could not find SMTP_USERNAME on environment variables")
	}
	smtpPassword, ok := os.LookupEnv("SMTP_PASSWORD")
	if !ok {
		log.Fatalln("could not find SMTP_PASSWORD on environment variables")
	}
	config.smtpAuth = smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
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

	router := way.NewRouter()
	router.HandleFunc("POST", "/api/users", jsonRequired(createUser))
	router.HandleFunc("POST", "/api/passwordless/start", jsonRequired(passwordlessStart))
	router.HandleFunc("GET", "/api/passwordless/verify_redirect", passwordlessVerifyRedirect)
	router.Handle("GET", "/api/auth_user", authRequired(getAuthUser))
	router.HandleFunc("GET", "/", serveFile("views/index.html"))
	router.HandleFunc("GET", "/callback", serveFile("views/callback.html"))

	addr := fmt.Sprintf(":%d", config.port)
	log.Printf("starting server at %s 🚀\n", config.appURL)
	log.Fatalf("could not start server: %v\n", http.ListenAndServe(addr, withRecover(router)))
}

func env(key, fallbackValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	return v
}

func respondJSON(w http.ResponseWriter, payload interface{}, code int) {
	switch value := payload.(type) {
	case string:
		payload = map[string]string{"message": value}
	case int:
		payload = map[string]int{"value": value}
	case bool:
		payload = map[string]bool{"result": value}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		respondInternalError(w, fmt.Errorf("could not marshal response payload: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(b)
}

func respondInternalError(w http.ResponseWriter, err error) {
	log.Println(err)
	respondJSON(w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func jsonRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		isJSON := strings.HasPrefix(ct, "application/json")
		if !isJSON {
			respondJSON(w, "JSON body required", http.StatusUnsupportedMediaType)
			return
		}
		next(w, r)
	}
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic: %+v\n", r)
				debug.PrintStack()

				var err error
				switch x := r.(type) {
				case error:
					err = x
				case string:
					err = errors.New(x)
				default:
					err = errors.New("Unknown panic")
				}
				respondInternalError(w, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func serveFile(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	}
}

func sendMail(to mail.Address, subject, body string) error {
	from := mail.Address{
		Name:    "Passwordless Demo",
		Address: "noreply@" + config.appURL.Host,
	}
	headers := map[string]string{
		"From":         from.String(),
		"To":           to.String(),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": `text/html; charset="utf-8"`,
	}
	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n"
	msg += body

	return smtp.SendMail(
		config.smtpAddr,
		config.smtpAuth,
		from.Address,
		[]string{to.Address},
		[]byte(msg))
}
