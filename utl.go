package main

import "github.com/lib/pq"

func isNotNullViolation(err error) bool {
	errPq, ok := err.(*pq.Error)
	return ok && errPq.Code == "23502"
}

func isUniqueViolation(err error) bool {
	errPq, ok := err.(*pq.Error)
	return ok && errPq.Code == "23505"
}
