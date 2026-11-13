package dominio

import (
	"fmt"
	"strconv"
	"strings"
)

// ---------- SKU (Value Object) ----------

// SKU identifica un producto. Se normaliza a mayúsculas y no admite espacios.
type SKU string

func NuevoSKU(texto string) (SKU, error) {
	t := strings.ToUpper(strings.TrimSpace(texto))
	if t == "" || strings.ContainsAny(t, " \t") {
		return "", fmt.Errorf("%w: %q", ErrSKUInvalido, texto)
	}
	return SKU(t), nil
}

// ---------- Cantidad (Value Object) ----------

// Cantidad son unidades base de producto. Nunca es negativa: el invariante
// se garantiza en el constructor y en Restar, así que no existe una
// Cantidad inválida en todo el sistema.
type Cantidad struct{ valor int }

func NuevaCantidad(n int) (Cantidad, error) {
	if n < 0 {
		return Cantidad{}, fmt.Errorf("%w: %d", ErrCantidadNegativa, n)
	}
	return Cantidad{n}, nil
}

func (c Cantidad) Valor() int { return c.valor }

func (c Cantidad) Sumar(o Cantidad) Cantidad { return Cantidad{c.valor + o.valor} }

func (c Cantidad) Restar(o Cantidad) (Cantidad, error) {
	if o.valor > c.valor {
		return Cantidad{}, fmt.Errorf("%w: hay %d y se piden %d", ErrExistenciasInsuficientes, c.valor, o.valor)
	}
	return Cantidad{c.valor - o.valor}, nil
}

// ---------- Dinero (Value Object) ----------

// Dinero en colones enteros (evita errores de redondeo de float).
type Dinero struct{ colones int64 }

func Colones(n int64) (Dinero, error) {
	if n < 0 {
		return Dinero{}, fmt.Errorf("%w: %d", ErrDineroInvalido, n)
	}
	return Dinero{n}, nil
}

func (d Dinero) Colones() int64 { return d.colones }

func (d Dinero) Por(n int) Dinero { return Dinero{d.colones * int64(n)} }

func (d Dinero) Sumar(o Dinero) Dinero { return Dinero{d.colones + o.colones} }

func (d Dinero) String() string {
	s := strconv.FormatInt(d.colones, 10)
	var partes []string
	for len(s) > 3 {
		partes = append([]string{s[len(s)-3:]}, partes...)
		s = s[:len(s)-3]
	}
	partes = append([]string{s}, partes...)
	return "₡" + strings.Join(partes, ".")
}

// ---------- Presentacion (Value Object) ----------

// Presentacion describe cómo se empaca un producto: "Unidad" = 1,
// "Caja x12" = 12 unidades base.
type Presentacion struct {
	nombre   string
	unidades int
}

func NuevaPresentacion(nombre string, unidades int) (Presentacion, error) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" || unidades <= 0 {
		return Presentacion{}, fmt.Errorf("%w: %q con %d unidades", ErrPresentacionInvalida, nombre, unidades)
	}
	return Presentacion{nombre, unidades}, nil
}

func (p Presentacion) Nombre() string { return p.nombre }
func (p Presentacion) Unidades() int  { return p.unidades }
