# FASE 1 — Core Lógico en Go (Local)

## Objetivo

Implementar toda la lógica de negocio del sistema SIN depender de servicios externos. Todo debe funcionar con salida por consola.

---

## Lo que vas a aprender

- Paquete `time` de Go (fechas, días de semana, comparaciones)
- Slice + `math/rand` para selección aleatoria
- Structs con métodos vs funciones sueltas — cuándo usar cada una
- Separación por responsabilidad de dominio (cada package resuelve UN problema)
- Cómo estructurar lógica que después va a ser llamada desde un entry point

---

## Módulos a crear

### Módulo `internal/kitchen/` — Alerta de Cocina (Feature 1)

**Responsabilidad:** Saber si hoy a las 8 PM hay que mandar el recordatorio de la cena.

**La respuesta es trivial:** SI, todos los días. Pero igual necesitás modelarlo como un paquete porque:

1. Establece el patrón de dominio desde el principio
2. Si después querés modificar la lógica (ej: no mandar los domingos), el cambio está aislado
3. Es consistente con la arquitectura del boilerplate

**Paso a paso:**

1. Creá `internal/kitchen/check.go`
2. Definí una función (o struct con método) tipo `ShouldRemind(now time.Time) bool`
   - Por ahora, que siempre retorne `true`
   - ¿Por qué recibe `time.Time` como parámetro en vez de llamar a `time.Now()` adentro? **Pista:** testabilidad. Investigá por qué es mala práctica llamar a `time.Now()` directamente en una función de dominio.
3. Definí `Message() string` que retorna el mensaje del recordatorio
   - _"Gino, ve a revisar la cocina. ¿Está listo el arroz de los perros?"_
   - El mensaje debería ser una constante exportada o una función que lo retorne

**Referencia al boilerplate:** En tu `internal/auth/entity.go`, la entidad `User` tiene un método `Validate()`. Acá estás haciendo lo mismo — lógica de dominio pegada a la entidad/concepto.

---

### Módulo `internal/butcher/` — Planificador de Compras (Feature 2 + 3)

Este módulo tiene DOS responsabilidades. ¿Deberían ser dos packages separados? Pensalo. Si las plantillas "anti-bot" son solo para el carnicero, tiene sentido que estén juntas. Si después apareciera otro tipo de plantilla, ahí las separarías.

#### Sub-módulo A: Detector de findes de semana de compras

**La lógica de negocio:** El pedido al carnicero se hace el primer fin de semana del mes. La alerta se dispara el VIERNES anterior a ese fin de semana.

**Algoritmo conceptual:**

1. Recibís una fecha (hoy)
2. Calculás qué día es mañana (sábado) y pasado mañana (domingo)
3. Si mañana es sábado Y es el primer sábado del mes → compras este finde
4. O si pasado mañana es domingo Y es el primer domingo del mes → compras este finde
5. Si estamos en viernes Y el fin de semana es de compras → disparar alerta

**Paso a paso:**

1. Creá `internal/butcher/detector.go`
2. Definí una función `IsShoppingWeekend(today time.Time) bool`
   - Investigá: `time.Weekday()`, `time.Date()`, cómo obtener el primer día del mes
   - Edge case: ¿qué pasa si es viernes 30 y el sábado 1 es el primer día del mes siguiente?
3. Definí una función `ShouldOrderReminder(now time.Time) bool`
   - Compone a `IsShoppingWeekend` + verifica que hoy sea viernes
   - **Referencia:** esto es como el método `Validate()` en tu entity — composición de lógica simple

**Edge cases que TENÉS que considerar (y probar con prints):**

- ¿Qué pasa el 31 de enero si cae viernes y el 1 de febrero es sábado?
- ¿Qué pasa el primer viernes del mes si el sábado 1 fue el fin de semana de compras?
- ¿Qué pasa un miércoles cualquiera?

Probá cada uno de estos casos creando fechas manualmente con `time.Date()`.

#### Sub-módulo B: Motor de Plantillas Anti-Bot

**La lógica de negocio:** Cada vez que se envía el pedido al carnicero, se selecciona una variante de mensaje al azar para no sonar robótico.

**Paso a paso:**

1. Creá `internal/butcher/template.go`
2. Definí un slice de strings con las variantes del mensaje
   - Son solo 2 en el PRD, pero el slice puede tener más. La estructura permite crecer.
3. Definí una función `RandomButcherMessage() string`
   - Usá `math/rand` para seleccionar un índice al azar
   - **IMPORTANTE:** Investigá por qué necesitás seed en `math/rand` y cómo inicializarlo correctamente para que NO te dé siempre el mismo resultado
   - En Go 1.20+ ya no necesitás seed manual — investigá qué cambió

**Referencia al boilerplate:** En `internal/auth/jwt.go` tenés funciones que generan tokens con lógica específica. Acá es el mismo patrón: una función que encapsula lógica y retorna un resultado.

---

### Conectar todo en `cmd/check/main.go`

Ahora actualizá `cmd/check/main.go` para que:

1. Cargue la configuración (como en Fase 0)
2. Cree una instancia de la fecha actual (`time.Now()`)
3. Llame al módulo `kitchen` y muestre:
   - `[COCINA] Debería recordar: Sí/No`
   - `[COCINA] Mensaje: "..."` (solo si debe recordar)
4. Llame al módulo `butcher` y muestre:
   - `[BUTCHER] Es finde de compras: Sí/No`
   - `[BUTCHER] Debería recordar hoy: Sí/No`
   - `[BUTCHER] Mensaje para carnicero: "..."` (solo si debe recordar)

**La salida debería verse algo así (sin ser exactamente igual):**
```
[CONFIG] Entorno: local
[COCINA] Recordatorio diario: ACTIVADO
[COCINA] Mensaje: Gino, ve a revisar la cocina...
[BUTCHER] ¿Finde de compras? true
[BUTCHER] ¿Recordar hoy? true (es viernes)
[BUTCHER] Mensaje: Hola buenas, por fa para mañana...
```

**Importante:** NO imprimas las credenciales sensibles. Solo mostrá que se cargaron correctamente (ej: "Token configurado: sí/no").

---

## Verificación

Para confirmar que la lógica funciona:

1. **Modificá manualmente la fecha** — creá un `time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)` que caiga viernes y verificá que el detector funcione. Después probá con miércoles, domingo, etc.
2. **Ejecutá varias veces seguidas** el programa y verificá que el mensaje del carnicero cambie aleatoriamente.
3. **Comentá temporalmente** el `if` de `ShouldOrderReminder` para forzar la salida y verificar todos los caminos.

```bash
go run ./cmd/check/
```

---

## Lo que NO tenés que hacer todavía

- ❌ No envíes WhatsApps reales
- ❌ No configures SMTP
- ❌ No pienses en AWS Lambda
- ❌ No escribas tests todavía (aunque ya te esté picando el bichito)

---

## Para reflexionar antes de pasar a Fase 2

1. ¿Por qué separé `kitchen` de `butcher` si ambos son "alertas"?
   - **Respuesta:** Cambian por diferentes razones. La cocina es DIARIA y siempre igual. El carnicero es CONDICIONAL (solo findes de mes) y tiene lógica de fechas compleja.
2. ¿Por qué el detector recibe un `time.Time` en vez de usar `time.Now()` internamente?
   - **Respuesta:** Testabilidad y control. Si querés probar qué pasa un viernes 31, podés pasarle cualquier fecha.
3. ¿Por qué las plantillas están en `butcher` y no en un package separado "templates"?
   - **Respuesta:** Porque por ahora solo el carnicero usa plantillas. Si el recordatorio de cocina también tuviera variantes, ahí las separás. YAGNI.

Checklist:
- [ ] `kitchen/check.go` compila y retorna mensaje
- [ ] `butcher/detector.go` detecta correctamente findes de compra
- [ ] `butcher/template.go` selecciona variante aleatoria
- [ ] `go run ./cmd/check/` muestra toda la salida esperada
- [ ] Probaste con diferentes fechas modificando manualmente `time.Date()`
