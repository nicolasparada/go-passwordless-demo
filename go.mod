// +heroku goVersion go1.16
// +heroku install ./cmd/passwordless

module github.com/nicolasparada/go-passwordless-demo

go 1.16

require (
	github.com/cockroachdb/cockroach-go v2.0.1+incompatible
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/hako/branca v0.0.0-20200807062402-6052ac720505
	github.com/hako/durafmt v0.0.0-20210316092057-3a2c319c1acd
	github.com/joho/godotenv v1.3.0
	github.com/lib/pq v1.10.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)
