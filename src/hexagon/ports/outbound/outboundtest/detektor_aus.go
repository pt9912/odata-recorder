//go:build !race

package outboundtest

// unterDetektor ist wahr, wenn der Test mit dem Datenrennen-Detektor gebaut ist (`make test-race`).
const unterDetektor = false
