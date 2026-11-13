package dominio

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Sucursal es una Entity. Actúa también como Factory de su Inventario.
type Sucursal struct {
	id     string
	nombre string
}

func NuevaSucursal(id, nombre string) (Sucursal, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(nombre) == "" {
		return Sucursal{}, fmt.Errorf("%w: id y nombre son obligatorios", ErrSucursalInvalida)
	}
	return Sucursal{id, nombre}, nil
}

func (s Sucursal) ID() string     { return s.id }
func (s Sucursal) Nombre() string { return s.nombre }

// NuevoInventario crea el inventario (vacío) de la sucursal. El reloj se
// inyecta para poder probar con fechas controladas.
func (s Sucursal) NuevoInventario(reloj func() time.Time) *Inventario {
	if reloj == nil {
		reloj = time.Now
	}
	return &Inventario{sucursalID: s.id, existencias: map[SKU]*existencia{}, reloj: reloj}
}

// existencia es una Entity interna del agregado Inventario: solo se alcanza
// a través de la raíz.
type existencia struct {
	cantidad Cantidad
	minimo   int
	objetivo int
}

// Faltante describe un producto que cayó al nivel mínimo o por debajo.
type Faltante struct {
	SKU      SKU
	Actual   int
	Minimo   int
	Objetivo int
}

// Inventario es la RAÍZ DEL AGREGADO de una sucursal. Todas las
// modificaciones de existencias pasan por aquí, de modo que el invariante
// "existencias >= 0 y = suma de movimientos" nunca se rompe.
type Inventario struct {
	sucursalID  string
	existencias map[SKU]*existencia
	movimientos []Movimiento
	reloj       func() time.Time
}

func (i *Inventario) SucursalID() string { return i.sucursalID }

// RegistrarProducto da de alta un SKU con su política de reposición
// (mínimo = disparador, objetivo = hasta dónde reponer). Inicia en 0.
func (i *Inventario) RegistrarProducto(sku SKU, minimo, objetivo int) error {
	if _, ok := i.existencias[sku]; ok {
		return fmt.Errorf("%w: %s en %s", ErrProductoDuplicado, sku, i.sucursalID)
	}
	if err := validarPolitica(minimo, objetivo); err != nil {
		return err
	}
	i.existencias[sku] = &existencia{minimo: minimo, objetivo: objetivo}
	return nil
}

func validarPolitica(minimo, objetivo int) error {
	if minimo < 0 || objetivo <= minimo {
		return fmt.Errorf("%w: mínimo %d, objetivo %d (el objetivo debe superar al mínimo)", ErrPoliticaInvalida, minimo, objetivo)
	}
	return nil
}

func (i *Inventario) ConfigurarPolitica(sku SKU, minimo, objetivo int) error {
	e, err := i.buscar(sku)
	if err != nil {
		return err
	}
	if err := validarPolitica(minimo, objetivo); err != nil {
		return err
	}
	e.minimo, e.objetivo = minimo, objetivo
	return nil
}

func (i *Inventario) buscar(sku SKU) (*existencia, error) {
	e, ok := i.existencias[sku]
	if !ok {
		return nil, fmt.Errorf("%w: %s en %s", ErrProductoNoRegistrado, sku, i.sucursalID)
	}
	return e, nil
}

func (i *Inventario) Existencias(sku SKU) (Cantidad, error) {
	e, err := i.buscar(sku)
	if err != nil {
		return Cantidad{}, err
	}
	return e.cantidad, nil
}

func (i *Inventario) RegistrarEntrada(sku SKU, unidades int, motivo string) error {
	e, err := i.buscar(sku)
	if err != nil {
		return err
	}
	if unidades <= 0 {
		return fmt.Errorf("%w: %d", ErrCantidadNoPositiva, unidades)
	}
	c, _ := NuevaCantidad(unidades)
	e.cantidad = e.cantidad.Sumar(c)
	i.anotar(sku, Entrada, unidades, motivo)
	return nil
}

// RegistrarSalida rechaza cualquier salida que dejaría existencias negativas.
func (i *Inventario) RegistrarSalida(sku SKU, unidades int, motivo string) error {
	e, err := i.buscar(sku)
	if err != nil {
		return err
	}
	if unidades <= 0 {
		return fmt.Errorf("%w: %d", ErrCantidadNoPositiva, unidades)
	}
	c, _ := NuevaCantidad(unidades)
	nueva, err := e.cantidad.Restar(c)
	if err != nil {
		return fmt.Errorf("salida de %s en %s: %w", sku, i.sucursalID, err)
	}
	e.cantidad = nueva
	i.anotar(sku, Salida, -unidades, motivo)
	return nil
}

// Ajustar fija las existencias a un valor contado (conteo físico). Exige motivo.
func (i *Inventario) Ajustar(sku SKU, nuevaCantidad int, motivo string) error {
	e, err := i.buscar(sku)
	if err != nil {
		return err
	}
	if strings.TrimSpace(motivo) == "" {
		return ErrMotivoRequerido
	}
	c, err := NuevaCantidad(nuevaCantidad)
	if err != nil {
		return err
	}
	delta := c.Valor() - e.cantidad.Valor()
	if delta == 0 {
		return nil
	}
	e.cantidad = c
	i.anotar(sku, Ajuste, delta, motivo)
	return nil
}

func (i *Inventario) anotar(sku SKU, tipo TipoMovimiento, delta int, motivo string) {
	i.movimientos = append(i.movimientos, Movimiento{
		id: len(i.movimientos) + 1, sku: sku, tipo: tipo,
		delta: delta, motivo: motivo, fecha: i.reloj(),
	})
}

// Movimientos devuelve una copia: el historial no se puede alterar desde fuera.
func (i *Inventario) Movimientos() []Movimiento {
	return append([]Movimiento(nil), i.movimientos...)
}

func (i *Inventario) HistorialDe(sku SKU) []Movimiento {
	var r []Movimiento
	for _, m := range i.movimientos {
		if m.sku == sku {
			r = append(r, m)
		}
	}
	return r
}

// ExistenciasEn reconstruye cuántas unidades había en una fecha pasada
// sumando los movimientos hasta ese momento.
func (i *Inventario) ExistenciasEn(sku SKU, fecha time.Time) int {
	total := 0
	for _, m := range i.movimientos {
		if m.sku == sku && !m.fecha.After(fecha) {
			total += m.delta
		}
	}
	return total
}

// VerificarConsistencia comprueba que el historial explica las existencias actuales.
func (i *Inventario) VerificarConsistencia() error {
	for sku, e := range i.existencias {
		suma := 0
		for _, m := range i.HistorialDe(sku) {
			suma += m.delta
		}
		if suma != e.cantidad.Valor() {
			return fmt.Errorf("%w: %s historial=%d existencias=%d", ErrInconsistencia, sku, suma, e.cantidad.Valor())
		}
	}
	return nil
}

// BajoMinimo lista los productos con existencias <= mínimo, ordenados por SKU.
func (i *Inventario) BajoMinimo() []Faltante {
	var r []Faltante
	for sku, e := range i.existencias {
		if e.cantidad.Valor() <= e.minimo {
			r = append(r, Faltante{sku, e.cantidad.Valor(), e.minimo, e.objetivo})
		}
	}
	sort.Slice(r, func(a, b int) bool { return r[a].SKU < r[b].SKU })
	return r
}
