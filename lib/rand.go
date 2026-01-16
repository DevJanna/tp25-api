package lib

import (
	"math/rand"
	"time"
)

type baserandom struct {
	src rand.Source
}

func (r *baserandom) gen(chars string, n int) string {
	b := make([]byte, n)
	l := int64(len(chars))
	for i := range b {
		b[i] = chars[r.src.Int63()%l]
	}
	return string(b)
}

func (r *baserandom) Char(len int) string {
	return r.gen(
		"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		len,
	)
}

func (r *baserandom) Upper(len int) string {
	return r.gen(
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		len,
	)
}

func (r *baserandom) Num(len int) string {
	return r.gen(
		"0123456789",
		len,
	)
}

type random interface {
	Char(len int) string
	Upper(len int) string
	Num(len int) string
}

var Rand random = &baserandom{
	src: rand.NewSource(time.Now().Unix()),
}
