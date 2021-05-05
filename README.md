# go-passwordless-demo

[Demo](https://go-passwordless-demo.herokuapp.com/)

## Build instructions

Make sure you have [CockroachDB](https://www.cockroachlabs.com/) installed, then:

```bash
cockroach start-single-node --insecure -listen-addr 127.0.0.1
```

Then make sure you have [Golang](https://golang.org/) installed too and build the code:

```bash
go build ./cmd/passwordless
```

Then run the server:<br>
_(Add `-migrate` the first time)_
```
./passwordless -migrate
```
