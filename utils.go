package events

import "github.com/satori/go.uuid"

func RandomID() string {
	return uuid.NewV4().String()
}
