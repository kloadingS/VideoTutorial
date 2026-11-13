// Package memoria implementa los repositorios del dominio en memoria
// (capa de infraestructura). Cambiar a una base de datos solo requiere
// otra implementación de las mismas interfaces.
package memoria

import (
	"fmt"
	"sort"

	"supermercado/dominio"
)

// ---- Productos ----

type Productos struct {
	datos map[dominio.SKU]*dominio.Producto
}

func NuevosProductos() *Productos { return &Productos{datos: map[dominio.SKU]*dominio.Producto{}} }

func (r *Productos) Guardar(p *dominio.Producto) error { r.datos[p.SKU()] = p; return nil }
func (r *Productos) Obtener(sku dominio.SKU) (*dominio.Producto, error) {
	if p, ok := r.datos[sku]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("%w: producto %s", dominio.ErrNoEncontrado, sku)
}

// ---- Proveedores ----

type Proveedores struct{ datos map[string]*dominio.Proveedor }

func NuevosProveedores() *Proveedores { return &Proveedores{datos: map[string]*dominio.Proveedor{}} }

func (r *Proveedores) Guardar(p *dominio.Proveedor) error { r.datos[p.ID()] = p; return nil }
func (r *Proveedores) Obtener(id string) (*dominio.Proveedor, error) {
	if p, ok := r.datos[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("%w: proveedor %s", dominio.ErrNoEncontrado, id)
}
func (r *Proveedores) Todos() ([]*dominio.Proveedor, error) {
	var l []*dominio.Proveedor
	for _, p := range r.datos {
		l = append(l, p)
	}
	sort.Slice(l, func(a, b int) bool { return l[a].ID() < l[b].ID() })
	return l, nil
}

// ---- Inventarios ----

type Inventarios struct {
	datos map[string]*dominio.Inventario
}

func NuevosInventarios() *Inventarios { return &Inventarios{datos: map[string]*dominio.Inventario{}} }

func (r *Inventarios) Guardar(i *dominio.Inventario) error { r.datos[i.SucursalID()] = i; return nil }
func (r *Inventarios) Obtener(sucursalID string) (*dominio.Inventario, error) {
	if i, ok := r.datos[sucursalID]; ok {
		return i, nil
	}
	return nil, fmt.Errorf("%w: inventario de %s", dominio.ErrNoEncontrado, sucursalID)
}

// ---- Solicitudes ----

type Solicitudes struct {
	datos map[string]*dominio.SolicitudAbastecimiento
}

func NuevasSolicitudes() *Solicitudes {
	return &Solicitudes{datos: map[string]*dominio.SolicitudAbastecimiento{}}
}

func (r *Solicitudes) Guardar(s *dominio.SolicitudAbastecimiento) error {
	r.datos[s.ID()] = s
	return nil
}
func (r *Solicitudes) Obtener(id string) (*dominio.SolicitudAbastecimiento, error) {
	if s, ok := r.datos[id]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("%w: solicitud %s", dominio.ErrNoEncontrado, id)
}
func (r *Solicitudes) PendientesDe(sucursalID string) ([]*dominio.SolicitudAbastecimiento, error) {
	var l []*dominio.SolicitudAbastecimiento
	for _, s := range r.datos {
		if s.SucursalID() == sucursalID && s.Pendiente() {
			l = append(l, s)
		}
	}
	sort.Slice(l, func(a, b int) bool { return l[a].ID() < l[b].ID() })
	return l, nil
}
