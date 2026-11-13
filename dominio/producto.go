package dominio

import (
	"fmt"
	"strings"
)

// Producto es una Entity (identidad = SKU) y raíz de su propio agregado.
// Su invariante: siempre tiene al menos una presentación y sus nombres son únicos.
type Producto struct {
	sku            SKU
	nombre         string
	categoria      string
	presentaciones map[string]Presentacion
}

// NuevoProducto es la Factory: no se puede construir un producto inválido.
func NuevoProducto(sku SKU, nombre, categoria string, presentaciones ...Presentacion) (*Producto, error) {
	if strings.TrimSpace(nombre) == "" {
		return nil, fmt.Errorf("%w: el nombre es obligatorio", ErrProductoInvalido)
	}
	if len(presentaciones) == 0 {
		return nil, fmt.Errorf("%w: requiere al menos una presentación", ErrProductoInvalido)
	}
	p := &Producto{sku: sku, nombre: nombre, categoria: categoria, presentaciones: map[string]Presentacion{}}
	for _, pr := range presentaciones {
		if err := p.AgregarPresentacion(pr); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *Producto) SKU() SKU          { return p.sku }
func (p *Producto) Nombre() string    { return p.nombre }
func (p *Producto) Categoria() string { return p.categoria }

func (p *Producto) AgregarPresentacion(pr Presentacion) error {
	if _, existe := p.presentaciones[pr.Nombre()]; existe {
		return fmt.Errorf("%w: %q repetida en %s", ErrPresentacionInvalida, pr.Nombre(), p.sku)
	}
	p.presentaciones[pr.Nombre()] = pr
	return nil
}

// UnidadesDe devuelve cuántas unidades base trae una presentación.
func (p *Producto) UnidadesDe(presentacion string) (int, error) {
	pr, ok := p.presentaciones[presentacion]
	if !ok {
		return 0, fmt.Errorf("%w: %s no tiene %q", ErrPresentacionDesconocida, p.sku, presentacion)
	}
	return pr.Unidades(), nil
}
