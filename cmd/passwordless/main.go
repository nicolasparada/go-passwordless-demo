package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/nicolasparada/go-passwordless-demo"
	smtpmailing "github.com/nicolasparada/go-passwordless-demo/mailing/smtp"
	"github.com/nicolasparada/go-passwordless-demo/repo/cockroach"
	httptransport "github.com/nicolasparada/go-passwordless-demo/transport/http"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := log.Default()
	err := run(ctx, logger, os.Args[1:])
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *log.Logger, args []string) error {
	var (
		port, _               = strconv.ParseUint(env("PORT", "3000"), 10, 64)
		cockroachURL          = env("COCKROACH_URL", "postgresql://root@127.0.0.1:26257/passwordless?sslmode=disable")
		disableCRDBRetries, _ = strconv.ParseBool(os.Getenv("DISABLE_CRDB_RETRIES"))
		smtpHost              = os.Getenv("SMTP_HOST")
		smtpPort, _           = strconv.ParseUint(os.Getenv("SMTP_PORT"), 10, 64)
		smtpUsername          = os.Getenv("SMTP_USERNAME")
		smtpPassword          = os.Getenv("SMTP_PASSWORD")
		originStr             = env("ORIGIN", fmt.Sprintf("http://localhost:%d", port))
		authTokenKey          = env("AUTH_TOKEN_KEY", "supersecretkeyyoushouldnotcommit")
	)

	fs := flag.NewFlagSet("passwordless", flag.ExitOnError)
	fs.Uint64Var(&port, "port", port, "HTTP port in which this very server listen")
	fs.StringVar(&cockroachURL, "cockroach", cockroachURL, "CockroachDB URL")
	fs.BoolVar(&disableCRDBRetries, "disable-crdb-retries", disableCRDBRetries, "Disables cockroach transaction retries")
	fs.StringVar(&originStr, "origin", originStr, "URL origin of this very server")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("could not parse flags: %w", err)
	}

	db, err := sql.Open("postgres", cockroachURL)
	if err != nil {
		return fmt.Errorf("could not open cockroach db: %w", err)
	}

	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("could not ping cockroach: %w", err)
	}

	origin, err := url.Parse(originStr)
	if err != nil {
		return fmt.Errorf("could not parse origin: %w", err)
	}

	if !origin.IsAbs() {
		return errors.New("origin must be absolute")
	}

	repo := &cockroach.Repository{DB: db, DisableCRDBRetries: disableCRDBRetries}
	mailSender := &smtpmailing.Sender{
		FromName:    "Passwordless",
		FromAddress: "noreply@" + origin.Hostname(),
		Host:        smtpHost,
		Port:        smtpPort,
		Username:    smtpUsername,
		Password:    smtpPassword,
	}
	svc := &passwordless.Service{
		Logger:       logger,
		Origin:       origin,
		Repository:   repo,
		MailSender:   mailSender,
		AuthTokenKey: authTokenKey,
	}
	h := httptransport.NewHandler(svc, logger)

	srv := &http.Server{
		Handler: h,
		Addr:    fmt.Sprintf(":%d", port),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()
		fmt.Println()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Printf("could not shutdown http server: %v\n", err)
			os.Exit(1)
		}
	}()

	logger.Printf("accepting http connections at %s\n", srv.Addr)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not http listen and serve: %w", err)
	}

	return nil
}

func env(key, fallback string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return v
}
