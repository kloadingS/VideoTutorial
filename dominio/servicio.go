package dominio

import (
	"fmt"
	"sort"
)

// ResultadoGeneracion resume qué ocurrió al generar solicitudes.
type ResultadoGeneracion struct {
	Solicitudes  []*SolicitudAbastecimiento
	Omitidos     []SKU // ya tenían una solicitud pendiente
	SinProveedor []SKU // ningún proveedor los ofrece
}

// ServicioAbastecimiento es un Domain Service: la lógica involucra varios
// agregados (Inventario, Producto, Proveedor, Solicitud) y no pertenece
// naturalmente a ninguno de ellos.
type ServicioAbastecimiento struct {
	productos   ProductoRepositorio
	proveedores ProveedorRepositorio
	solicitudes SolicitudRepositorio
	contador    int
}

func NuevoServicioAbastecimiento(p ProductoRepositorio, v ProveedorRepositorio, s SolicitudRepositorio) *ServicioAbastecimiento {
	return &ServicioAbastecimiento{productos: p, proveedores: v, solicitudes: s}
}

func (s *ServicioAbastecimiento) siguienteID() string {
	s.contador++
	return fmt.Sprintf("SOL-%04d", s.contador)
}

// GenerarSolicitudes revisa los productos bajo el mínimo y arma borradores:
//  1. omite productos que ya tienen una solicitud pendiente;
//  2. elige el proveedor con menor precio por unidad base;
//  3. calcula presentaciones para llegar al objetivo (redondeo hacia arriba);
//  4. ajusta al rango mínimo/máximo del proveedor;
//  5. agrupa las líneas en una solicitud por proveedor.
func (s *ServicioAbastecimiento) GenerarSolicitudes(inv *Inventario) (ResultadoGeneracion, error) {
	var res ResultadoGeneracion

	pendientes := map[SKU]bool{}
	previas, err := s.solicitudes.PendientesDe(inv.SucursalID())
	if err != nil {
		return res, err
	}
	for _, sol := range previas {
		for _, l := range sol.Lineas() {
			pendientes[l.SKU()] = true
		}
	}

	proveedores, err := s.proveedores.Todos()
	if err != nil {
		return res, err
	}

	porProveedor := map[string]*SolicitudAbastecimiento{}
	for _, f := range inv.BajoMinimo() {
		if pendientes[f.SKU] {
			res.Omitidos = append(res.Omitidos, f.SKU)
			continue
		}
		prod, err := s.productos.Obtener(f.SKU)
		if err != nil {
			return res, err
		}
		prov, cond, unidades, ok := elegirProveedor(prod, proveedores)
		if !ok {
			res.SinProveedor = append(res.SinProveedor, f.SKU)
			continue
		}

		faltante := f.Objetivo - f.Actual
		cant := (faltante + unidades - 1) / unidades
		if cant < cond.Minimo() {
			cant = cond.Minimo()
		}
		if cant > cond.Maximo() {
			cant = cond.Maximo()
		}

		linea, err := prov.CrearLinea(prod, cant)
		if err != nil {
			return res, err
		}
		sol := porProveedor[prov.ID()]
		if sol == nil {
			if sol, err = NuevaSolicitud(s.siguienteID(), inv.SucursalID(), prov.ID()); err != nil {
				return res, err
			}
			porProveedor[prov.ID()] = sol
		}
		if err := sol.AgregarLinea(linea); err != nil {
			return res, err
		}
	}

	ids := make([]string, 0, len(porProveedor))
	for id := range porProveedor {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := s.solicitudes.Guardar(porProveedor[id]); err != nil {
			return res, err
		}
		res.Solicitudes = append(res.Solicitudes, porProveedor[id])
	}
	return res, nil
}

// elegirProveedor compara precio por unidad base con multiplicación cruzada
// (sin decimales). En empate gana el proveedor de menor ID (la lista llega ordenada).
func elegirProveedor(prod *Producto, proveedores []*Proveedor) (*Proveedor, CondicionSuministro, int, bool) {
	var mejor *Proveedor
	var mejorCond CondicionSuministro
	mejorU := 0
	for _, p := range proveedores {
		c, ok := p.Condicion(prod.SKU())
		if !ok {
			continue
		}
		u, err := prod.UnidadesDe(c.Presentacion())
		if err != nil {
			continue
		}
		if mejor == nil || c.Precio().Colones()*int64(mejorU) < mejorCond.Precio().Colones()*int64(u) {
			mejor, mejorCond, mejorU = p, c, u
		}
	}
	return mejor, mejorCond, mejorU, mejor != nil
}

// RecibirSolicitud coordina dos agregados: marca la solicitud como recibida y
// registra una entrada por cada línea en el inventario de la sucursal.
func (s *ServicioAbastecimiento) RecibirSolicitud(sol *SolicitudAbastecimiento, inv *Inventario) error {
	if sol.SucursalID() != inv.SucursalID() {
		return fmt.Errorf("%w: la solicitud %s es de otra sucursal", ErrSolicitudInvalida, sol.ID())
	}
	for _, l := range sol.Lineas() { // validar antes de modificar nada
		if _, err := inv.Existencias(l.SKU()); err != nil {
			return err
		}
	}
	lineas, err := sol.Recibir()
	if err != nil {
		return err
	}
	for _, l := range lineas {
		if err := inv.RegistrarEntrada(l.SKU(), l.Unidades(), "Recepción "+sol.ID()); err != nil {
			return err
		}
	}
	return s.solicitudes.Guardar(sol)
}
