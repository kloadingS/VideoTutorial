package dominio

import (
	"fmt"
	"strings"
)

// CondicionSuministro es un Value Object: lo que un proveedor ofrece de un
// producto (presentación, precio por presentación y rango de cantidades).
type CondicionSuministro struct {
	sku          SKU
	presentacion string
	precio       Dinero
	minimo       int
	maximo       int
}

func NuevaCondicion(sku SKU, presentacion string, precio Dinero, minimo, maximo int) (CondicionSuministro, error) {
	if strings.TrimSpace(presentacion) == "" || precio.Colones() <= 0 || minimo < 1 || maximo < minimo {
		return CondicionSuministro{}, fmt.Errorf("%w: %s (%q, %s, mín %d, máx %d)", ErrCondicionInvalida, sku, presentacion, precio, minimo, maximo)
	}
	return CondicionSuministro{sku, presentacion, precio, minimo, maximo}, nil
}

func (c CondicionSuministro) SKU() SKU             { return c.sku }
func (c CondicionSuministro) Presentacion() string { return c.presentacion }
func (c CondicionSuministro) Precio() Dinero       { return c.precio }
func (c CondicionSuministro) Minimo() int          { return c.minimo }
func (c CondicionSuministro) Maximo() int          { return c.maximo }

// Admite verifica que la cantidad de presentaciones esté dentro del rango.
func (c CondicionSuministro) Admite(cantidad int) error {
	if cantidad < c.minimo || cantidad > c.maximo {
		return fmt.Errorf("%w: %d (permitido %d a %d %s)", ErrCantidadFueraDeRango, cantidad, c.minimo, c.maximo, c.presentacion)
	}
	return nil
}

// Proveedor es Entity y raíz de agregado: sus condiciones solo cambian a través de él.
type Proveedor struct {
	id          string
	nombre      string
	condiciones map[SKU]CondicionSuministro
}

func NuevoProveedor(id, nombre string) (*Proveedor, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(nombre) == "" {
		return nil, fmt.Errorf("%w: id y nombre son obligatorios", ErrProveedorInvalido)
	}
	return &Proveedor{id: id, nombre: nombre, condiciones: map[SKU]CondicionSuministro{}}, nil
}

func (p *Proveedor) ID() string     { return p.id }
func (p *Proveedor) Nombre() string { return p.nombre }

// AgregarCondicion registra o actualiza lo que el proveedor ofrece de un SKU.
func (p *Proveedor) AgregarCondicion(c CondicionSuministro) { p.condiciones[c.sku] = c }

func (p *Proveedor) Condicion(sku SKU) (CondicionSuministro, bool) {
	c, ok := p.condiciones[sku]
	return c, ok
}

// CrearLinea es la Factory de líneas de solicitud: aplica TODAS las reglas del
// proveedor (ofrece el producto, la presentación existe, cantidad en rango) y
// congela el precio vigente en la línea.
func (p *Proveedor) CrearLinea(prod *Producto, cantidad int) (LineaSolicitud, error) {
	c, ok := p.condiciones[prod.SKU()]
	if !ok {
		return LineaSolicitud{}, fmt.Errorf("%w: %s no vende %s", ErrProveedorNoOfrece, p.nombre, prod.SKU())
	}
	u, err := prod.UnidadesDe(c.presentacion)
	if err != nil {
		return LineaSolicitud{}, err
	}
	if err := c.Admite(cantidad); err != nil {
		return LineaSolicitud{}, fmt.Errorf("%s: %w", p.nombre, err)
	}
	return LineaSolicitud{
		proveedorID: p.id, sku: prod.SKU(), presentacion: c.presentacion,
		cantidad: cantidad, porPresentacion: u, precio: c.precio,
	}, nil
}
