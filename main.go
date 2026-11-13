// Demostración del dominio: ejecuta los escenarios normales y excepcionales
// del enunciado. Uso: go run .
package main

import (
	"fmt"
	"time"

	"supermercado/dominio"
	"supermercado/memoria"
)

func main() {
	// ---------- Infraestructura ----------
	productos := memoria.NuevosProductos()
	proveedores := memoria.NuevosProveedores()
	inventarios := memoria.NuevosInventarios()
	solicitudes := memoria.NuevasSolicitudes()
	servicio := dominio.NuevoServicioAbastecimiento(productos, proveedores, solicitudes)

	reloj := relojSimulado()
	cargarCatalogo(productos, proveedores)

	alajuela, _ := dominio.NuevaSucursal("ALA", "Sucursal Alajuela")
	heredia, _ := dominio.NuevaSucursal("HER", "Sucursal Heredia")
	invAla, invHer := alajuela.NuevoInventario(reloj), heredia.NuevoInventario(reloj)
	inventarios.Guardar(invAla)
	inventarios.Guardar(invHer)

	titulo("1. Inventario inicial por sucursal (cada una es un agregado independiente)")
	for _, c := range []struct {
		inv             *dominio.Inventario
		sku             dominio.SKU
		min, obj, stock int
	}{
		{invAla, "ARR-1KG", 20, 100, 15}, {invAla, "LEC-1L", 12, 60, 10},
		{invAla, "ACE-1L", 6, 24, 30}, {invAla, "CAF-500", 5, 20, 2},
		{invHer, "LEC-1L", 12, 30, 11}, {invHer, "ARR-1KG", 20, 100, 80},
	} {
		revisar(c.inv.RegistrarProducto(c.sku, c.min, c.obj))
		revisar(c.inv.RegistrarEntrada(c.sku, c.stock, "Inventario inicial"))
	}
	mostrar(alajuela, invAla, "ARR-1KG", "LEC-1L", "ACE-1L", "CAF-500")
	mostrar(heredia, invHer, "LEC-1L", "ARR-1KG")

	titulo("2. Entradas y salidas (escenario normal)")
	revisar(invAla.RegistrarEntrada("ACE-1L", 12, "Compra directa"))
	revisar(invAla.RegistrarSalida("ACE-1L", 36, "Ventas del día"))
	mostrar(alajuela, invAla, "ACE-1L")

	titulo("3. Reglas que protegen el inventario (escenarios excepcionales)")
	intentar("Salida mayor a las existencias", invAla.RegistrarSalida("ARR-1KG", 99, "Venta"))
	intentar("Entrada de cantidad cero", invAla.RegistrarEntrada("ARR-1KG", 0, "Error de captura"))
	intentar("Ajuste sin motivo", invAla.Ajustar("ARR-1KG", 12, ""))
	intentar("Producto no registrado en la sucursal", invHer.RegistrarSalida("CAF-500", 1, "Venta"))
	revisar(invAla.Ajustar("ARR-1KG", 12, "Conteo físico: faltaban 3 unidades"))

	titulo("4. Productos bajo el nivel mínimo")
	for _, f := range invAla.BajoMinimo() {
		fmt.Printf("  %-8s existencias %3d (mínimo %d, objetivo %d)\n", f.SKU, f.Actual, f.Minimo, f.Objetivo)
	}

	titulo("5. Generación de solicitudes de abastecimiento (servicio de dominio)")
	res, err := servicio.GenerarSolicitudes(invAla)
	revisar(err)
	imprimirSolicitudes(res.Solicitudes)
	fmt.Println("  Sin proveedor disponible:", res.SinProveedor)

	resHer, err := servicio.GenerarSolicitudes(invHer)
	revisar(err)
	fmt.Println("  Heredia (leche falta 19 unidades = 2 cajas, pero el proveedor exige mínimo 3):")
	imprimirSolicitudes(resHer.Solicitudes)

	titulo("6. No se duplican solicitudes pendientes")
	otra, _ := servicio.GenerarSolicitudes(invAla)
	fmt.Printf("  Nuevas: %d | Omitidos por tener solicitud pendiente: %v\n", len(otra.Solicitudes), otra.Omitidos)

	titulo("7. Condiciones del proveedor")
	p2, _ := proveedores.Obtener("P2")
	prodAceite, _ := productos.Obtener("ACE-1L")
	prodLeche, _ := productos.Obtener("LEC-1L")
	_, e1 := p2.CrearLinea(prodAceite, 1)
	intentar("Pedir a un proveedor que no vende el producto", e1)
	_, e2 := p2.CrearLinea(prodLeche, 11)
	intentar("Pedir más del máximo permitido", e2)

	titulo("8. Ciclo de vida de una solicitud y recepción")
	sol := res.Solicitudes[1] // proveedor P2 (leche)
	intentar("Recibir una solicitud que no fue enviada", servicio.RecibirSolicitud(sol, invAla))
	revisar(sol.Enviar())
	fmt.Println("  Estado:", sol.Estado())
	revisar(servicio.RecibirSolicitud(sol, invAla))
	fmt.Println("  Estado:", sol.Estado())
	intentar("Cancelar una solicitud ya recibida", sol.Cancelar())
	mostrar(alajuela, invAla, "LEC-1L")

	titulo("9. Historial: trazabilidad y reconstrucción")
	for _, m := range invAla.HistorialDe("LEC-1L") {
		fmt.Printf("  #%d %s %-7s %+4d  %s\n", m.ID(), m.Fecha().Format("02/01 15:04"), m.Tipo(), m.Delta(), m.Motivo())
	}
	corte := time.Date(2026, 11, 2, 10, 30, 0, 0, time.UTC)
	fmt.Printf("  Existencias de leche el %s: %d\n", corte.Format("02/01 15:04"), invAla.ExistenciasEn("LEC-1L", corte))
	if err := invAla.VerificarConsistencia(); err == nil {
		fmt.Println("  Consistencia historial vs existencias: OK")
	}
}

// ---------- Datos de ejemplo ----------

func cargarCatalogo(productos *memoria.Productos, proveedores *memoria.Proveedores) {
	unidad, _ := dominio.NuevaPresentacion("Unidad", 1)
	defs := []struct {
		sku, nombre string
		caja        string
		unidades    int
	}{
		{"ARR-1KG", "Arroz 1kg", "Caja x20", 20}, {"LEC-1L", "Leche 1L", "Caja x12", 12},
		{"ACE-1L", "Aceite 1L", "Caja x6", 6}, {"CAF-500", "Café 500g", "Caja x10", 10},
	}
	for _, d := range defs {
		caja, _ := dominio.NuevaPresentacion(d.caja, d.unidades)
		sku, _ := dominio.NuevoSKU(d.sku)
		p, err := dominio.NuevoProducto(sku, d.nombre, "Abarrotes", unidad, caja)
		revisar(err)
		productos.Guardar(p)
	}
	cond := func(p *dominio.Proveedor, sku, pres string, precio int64, min, max int) {
		d, _ := dominio.Colones(precio)
		c, err := dominio.NuevaCondicion(dominio.SKU(sku), pres, d, min, max)
		revisar(err)
		p.AgregarCondicion(c)
	}
	p1, _ := dominio.NuevoProveedor("P1", "Distribuidora Central")
	cond(p1, "ARR-1KG", "Caja x20", 16000, 2, 50)
	cond(p1, "LEC-1L", "Caja x12", 9600, 1, 20)
	cond(p1, "ACE-1L", "Caja x6", 15000, 1, 10)
	p2, _ := dominio.NuevoProveedor("P2", "Lácteos del Norte")
	cond(p2, "ARR-1KG", "Caja x20", 17000, 1, 100)
	cond(p2, "LEC-1L", "Caja x12", 8400, 3, 10)
	proveedores.Guardar(p1)
	proveedores.Guardar(p2)
}

// ---------- Utilidades de impresión ----------

func relojSimulado() func() time.Time {
	t := time.Date(2026, 11, 2, 8, 0, 0, 0, time.UTC)
	return func() time.Time { t = t.Add(time.Hour); return t }
}

func titulo(s string) { fmt.Printf("\n=== %s ===\n", s) }

func revisar(err error) {
	if err != nil {
		panic(err)
	}
}

func intentar(descripcion string, err error) {
	if err == nil {
		fmt.Printf("  ✗ %s: ¡debió fallar!\n", descripcion)
		return
	}
	fmt.Printf("  ✓ %s -> rechazado: %v\n", descripcion, err)
}

func mostrar(s dominio.Sucursal, inv *dominio.Inventario, skus ...dominio.SKU) {
	fmt.Printf("  %s:", s.Nombre())
	for _, sku := range skus {
		c, _ := inv.Existencias(sku)
		fmt.Printf("  %s=%d", sku, c.Valor())
	}
	fmt.Println()
}

func imprimirSolicitudes(ss []*dominio.SolicitudAbastecimiento) {
	for _, s := range ss {
		fmt.Printf("  %s [%s] proveedor %s, total %s\n", s.ID(), s.Estado(), s.ProveedorID(), s.Total())
		for _, l := range s.Lineas() {
			fmt.Printf("      %-8s %d x %s (%d unidades) = %s\n", l.SKU(), l.Cantidad(), l.Presentacion(), l.Unidades(), l.Subtotal())
		}
	}
}
