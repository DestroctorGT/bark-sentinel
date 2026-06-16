# bark-sentinel — Guía de Aprendizaje

Este directorio contiene las guías paso a paso para construir **bark-sentinel** desde cero. Cada guía es independiente pero están diseñadas para leerse en orden.

El objetivo NO es copiar y pegar código. El objetivo es que LEAS cada paso, INVESTIGUÉS los conceptos mencionados, y ESCRIBAS el código vos mismo.

---

## Orden de lectura

| Guía | Qué hacés | Dependencias |
|------|-----------|-------------|
| **[FASE-0-arranque.md](FASE-0-arranque.md)** | Inicializás el proyecto, estructura de carpetas, config, entry point mínimo | Nada |
| **[FASE-1-core.md](FASE-1-core.md)** | Implementás la lógica de negocio: fechas, plantillas, consola | Fase 0 |
| **[FASE-2-integracion.md](FASE-2-integracion.md)** | Conectás WhatsApp real y Email real, probás localmente | Fase 1 |
| **[FASE-3-cloud.md](FASE-3-cloud.md)** | Subís todo a AWS Lambda, configurás EventBridge, olvidás que existe | Fase 2 |

---

## Cómo usar estas guías

1. **Leé la guía entera primero** antes de escribir una línea de código
2. **Investigá los conceptos** que no conozcas (hay pistas en cada paso)
3. **Escribí el código vos mismo** — no lo copies de ningún lado
4. **Verificá con el checklist** al final de cada fase
5. **NO pases a la siguiente fase** hasta que la actual funcione

---

## Referencia: Patrones del boilerplate `go-api-simple`

Durante las guías se mencionan patrones de tu boilerplate. Este es el mapeo conceptual entre ambos proyectos:

| go-api-simple (boilerplate) | bark-sentinel (este proyecto) |
|---|---|
| `cmd/main.go` — entry point HTTP | `cmd/check/main.go` — entry point local |
| `cmd/migrate/` + `cmd/seed/` — entry points adicionales | `cmd/lambda/main.go` — entry point cloud |
| `internal/platform/config/` — env loading | `internal/config/` — mismo patrón |
| `internal/auth/{entity,service,repository,handler}.go` | `internal/kitchen/check.go`, `internal/butcher/{detector,template}.go` |
| `internal/platform/storage/postgres.go` | NO APLICA (no hay DB) |
| `internal/platform/jsonio/` | NO APLICA (no hay API REST) |
| `internal/platform/middleware/` | NO APLICA (no hay HTTP) |

---

## Filosofía del proyecto

> "No es un proyecto para tener un sistema de alertas. Es un proyecto para APRENDER Go."

Cada decisión de arquitectura en estas guías tiene una razón pedagógica:
- Separar `kitchen` de `butcher` → responsabilidad única
- Usar `time.Time` como parámetro → testabilidad
- Una sola Lambda con decisión interna → simplicidad primero
- Sin DB → entender que no todo necesita persistencia
- Sin frameworks HTTP → entender que `net/http` alcanza para mucho

Disfrutá el proceso. Codeá con intención.
