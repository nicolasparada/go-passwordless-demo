Blog posts:
- [Server](https://nicolasparada.netlify.com/posts/passwordless-auth-server/)
- [Client](https://nicolasparada.netlify.com/posts/passwordless-auth-client/)

---

Install [Go](https://golang.org/) and [CockroachDB](https://www.cockroachlabs.com/).
Go to [mailtrap.io](https://mailtrap.io/) and create an account. Save your SMTP server credentials into the `.env` file.

Get the code:
```
go get -u github.com/nicolasparada/go-passwordless-demo
```

Start a Cockroach instance:
```bash
cockroach start --insecure --host 127.0.0.1
```

Create schema and run:
```bash
cat schema.sql | cockroach sql --insecure
go-passwordless-demo
```
