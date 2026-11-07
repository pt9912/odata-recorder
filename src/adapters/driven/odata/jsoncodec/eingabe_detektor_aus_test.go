//go:build !race

package jsoncodec_test

// eUnterDetektor ist wahr, wenn der Test mit dem Datenrennen-Detektor gebaut ist (`make test-race`).
const eUnterDetektor = false
