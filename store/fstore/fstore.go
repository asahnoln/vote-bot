package fstore

import (
	"cloud.google.com/go/firestore"
)

type Firestore struct {
	c *firestore.Client
}

const EqualOp = "=="

func New(c *firestore.Client) *Firestore {
	return &Firestore{c}
}
