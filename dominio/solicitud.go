package dominio

import (
	"fmt"
	"strings"
)

// LineaSolicitud es un Value Object: producto pedido, en cuántas presentaciones
// y a qué precio (precio congelado al crearla).
type LineaSolicitud struct {
	proveedorID     string
	sku             SKU
	presentacion    string
	cantidad        int
	porPresentacion int
	precio          Dinero
}

func (l LineaSolicitud) SKU() SKU             { return l.sku }
func (l LineaSolicitud) Presentacion() string { return l.presentacion }
func (l LineaSolicitud) Cantidad() int        { return l.cantidad }
func (l LineaSolicitud) Unidades() int        { return l.cantidad * l.porPresentacion }
func (l LineaSolicitud) Subtotal() Dinero     { return l.precio.Por(l.cantidad) }

type EstadoSolicitud string

const (
	Borrador  EstadoSolicitud = "BORRADOR"
	Enviada   EstadoSolicitud = "ENVIADA"
	Recibida  EstadoSolicitud = "RECIBIDA"
	Cancelada EstadoSolicitud = "CANCELADA"
)

// SolicitudAbastecimiento es raíz de agregado. Ciclo de vida:
// Borrador -> Enviada -> Recibida, y Cancelada desde Borrador o Enviada.
// Referencia a sucursal y proveedor solo por ID (no por objeto).
type SolicitudAbastecimiento struct {
	id          string
	sucursalID  string
	proveedorID string
	estado      EstadoSolicitud
	lineas      []LineaSolicitud
}

func NuevaSolicitud(id, sucursalID, proveedorID string) (*SolicitudAbastecimiento, error) {
	if strings.TrimSpace(id) == "" || sucursalID == "" || proveedorID == "" {
		return nil, fmt.Errorf("%w: id, sucursal y proveedor son obligatorios", ErrSolicitudInvalida)
	}
	return &SolicitudAbastecimiento{id: id, sucursalID: sucursalID, proveedorID: proveedorID, estado: Borrador}, nil
}

func (s *SolicitudAbastecimiento) ID() string              { return s.id }
func (s *SolicitudAbastecimiento) SucursalID() string      { return s.sucursalID }
func (s *SolicitudAbastecimiento) ProveedorID() string     { return s.proveedorID }
func (s *SolicitudAbastecimiento) Estado() EstadoSolicitud { return s.estado }
func (s *SolicitudAbastecimiento) Lineas() []LineaSolicitud {
	return append([]LineaSolicitud(nil), s.lineas...)
}

// Pendiente indica que la solicitud aún "cuenta" como pedido en camino.
func (s *SolicitudAbastecimiento) Pendiente() bool {
	return s.estado == Borrador || s.estado == Enviada
}

func (s *SolicitudAbastecimiento) Total() Dinero {
	var t Dinero
	for _, l := range s.lineas {
		t = t.Sumar(l.Subtotal())
	}
	return t
}

func (s *SolicitudAbastecimiento) AgregarLinea(l LineaSolicitud) error {
	if s.estado != Borrador {
		return fmt.Errorf("%w: no se agregan líneas en estado %s", ErrTransicionInvalida, s.estado)
	}
	if l.proveedorID != s.proveedorID {
		return fmt.Errorf("%w: %s", ErrLineaAjena, l.sku)
	}
	for _, x := range s.lineas {
		if x.sku == l.sku {
			return fmt.Errorf("%w: %s", ErrLineaDuplicada, l.sku)
		}
	}
	s.lineas = append(s.lineas, l)
	return nil
}

func (s *SolicitudAbastecimiento) Enviar() error {
	if s.estado != Borrador {
		return fmt.Errorf("%w: %s -> %s", ErrTransicionInvalida, s.estado, Enviada)
	}
	if len(s.lineas) == 0 {
		return ErrSolicitudSinLineas
	}
	s.estado = Enviada
	return nil
}

func (s *SolicitudAbastecimiento) Cancelar() error {
	if !s.Pendiente() {
		return fmt.Errorf("%w: %s -> %s", ErrTransicionInvalida, s.estado, Cancelada)
	}
	s.estado = Cancelada
	return nil
}

// Recibir marca la solicitud como recibida y devuelve las líneas para que el
// servicio de dominio registre las entradas en el inventario.
func (s *SolicitudAbastecimiento) Recibir() ([]LineaSolicitud, error) {
	if s.estado != Enviada {
		return nil, fmt.Errorf("%w: %s -> %s", ErrTransicionInvalida, s.estado, Recibida)
	}
	s.estado = Recibida
	return s.Lineas(), nil
}
