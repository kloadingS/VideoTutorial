package dominio

import "time"

type TipoMovimiento string

const (
	Entrada TipoMovimiento = "ENTRADA"
	Salida  TipoMovimiento = "SALIDA"
	Ajuste  TipoMovimiento = "AJUSTE"
)

// Movimiento es un Value Object inmutable: un hecho ocurrido en el inventario.
// Solo se crea desde Inventario (campos no exportados) y nunca se modifica;
// la suma de sus deltas reconstruye las existencias.
type Movimiento struct {
	id     int
	sku    SKU
	tipo   TipoMovimiento
	delta  int // positivo suma, negativo resta
	motivo string
	fecha  time.Time
}

func (m Movimiento) ID() int              { return m.id }
func (m Movimiento) SKU() SKU             { return m.sku }
func (m Movimiento) Tipo() TipoMovimiento { return m.tipo }
func (m Movimiento) Delta() int           { return m.delta }
func (m Movimiento) Motivo() string       { return m.motivo }
func (m Movimiento) Fecha() time.Time     { return m.fecha }
