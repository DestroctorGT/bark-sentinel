# FASE 0 — Arranque: Inicialización del Proyecto

## Objetivo

Crear el esqueleto del proyecto desde cero: módulo Go, estructura de directorios, configuración por entorno, y un entry point que verifique que todo funciona.

---

## Lo que vas a aprender

- `go mod init` y dependencias mínimas
- Package layout al estilo Go estándar (no frameworks)
- Patrón de configuración por env vars (como en go-api-simple)
- Entry points múltiples con `cmd/`

---

## Paso a paso

### 1. Crear el módulo Go

```bash
mkdir bark-sentinel && cd bark-sentinel
go mod init github.com/<tu-usuario>/bark-sentinel
```

**Concepto a investigar:** ¿Qué es `go.mod`? ¿Qué significa el module path? ¿Por qué usamos una ruta de GitHub aunque el proyecto sea local?

### 2. Estructura de directorios

Creá esta estructura VACÍA (solo carpetas, sin archivos todavía):

```
bark-sentinel/
├── cmd/
│   ├── check/           # Entry point para pruebas locales
│   └── lambda/          # Entry point para AWS Lambda (Fase 3)
├── internal/
│   ├── config/          # Carga de configuración desde env vars
│   ├── kitchen/         # Módulo: alerta diaria de cocina
│   ├── butcher/         # Módulo: detector de findes de mes + plantillas
│   └── notifier/        # Módulo: envío de notificaciones (WhatsApp + Email)
├── .env.local           # Tus credenciales reales (NUNCA se commitea)
├── .env.local.example   # Template de variables para otros devs
└── .gitignore
```

**Referencia:** Mirá cómo tu boilerplate usa `cmd/main.go`, `cmd/migrate/main.go`, etc. — acá hacemos lo mismo pero con `check` y `lambda`.

**Concepto a investigar:** ¿Por qué Go usa `internal/` para restringir importaciones? Leé la spec de `internal` packages.

### 3. Crear `.gitignore`

Bajate el .gitignore estándar de Go. Los patrones clave que necesitás:

- Ignorar `.env.local` (y cualquier `.env.*` que no sea `.example`)
- Ignorar binarios compilados (`/bark-sentinel`, `*.exe`)
- Ignorar `vendor/` si no usás vendoring
- Ignorar archivos de IDE (`.idea/`, `.vscode/`)

**Inspiración:** Fijate cómo tu boilerplate maneja los `.env` — commitea `.env` base pero ignora `.env.local`.

### 4. Implementar `internal/config/config.go`

Este paquete es **casi un calco** del `internal/platform/config/config.go` de tu boilerplate. La idea:

1. Usá `github.com/joho/godotenv` para cargar archivos `.env`
2. Definí una struct `Config` con los campos que vas a necesitar:
   - `AppEnv` — para saber si estamos en local o producción
   - `WhatsAppToken` — token de la API de WhatsApp
   - `WhatsAppPhoneNumber` — tu número (Gino)
   - `ButcherPhoneNumber` — número del carnicero
   - `GmailUser` y `GmailPassword` — credenciales SMTP
   - `SmtpHost` y `SmtpPort` — servidor SMTP
3. Implementá `Load()` que retorna `(*Config, error)`
4. Implementá `validate()` que chequea campos requeridos

**Referencia directa:** Mirá cómo tu boilerplate hace `getEnv()` y `getEnvInt()` — son helpers reutilizables. También fijate cómo valida campos requeridos en `validate()`.

**Diferencia clave:** En el boilerplate hay sub-structs (DB, JWT, Admin). Para este proyecto, como las variables son más planas, podés tener una struct más chata o agrupar por dominio (ej: `WhatsAppConfig`, `SMTPConfig`). Elegí lo que te haga sentido.

### 5. Crear `.env.local.example`

Variables que vas a necesitar eventualmente:

```
APP_ENV=local

# WhatsApp
WHATSAPP_TOKEN=
WHATSAPP_PHONE_NUMBER=
BUTCHER_PHONE_NUMBER=

# Gmail SMTP
GMAIL_USER=
GMAIL_PASSWORD=
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
```

Dejá los valores vacíos o con placeholders. Nunca pongas credenciales reales.

### 6. Crear `cmd/check/main.go` — Entry point mínimo

El propósito de este archivo es **arrancar localmente y probar que todo funciona sin AWS**. Es análogo a `cmd/main.go` en tu boilerplate, pero sin router HTTP.

Lo que tiene que hacer este entry point:

1. Llamar a `config.Load()` para cargar la configuración
2. Hacer un `log.Printf` o `fmt.Println` con la configuración cargada
3. Por ahora no hace nada más — es el "Hola Mundo" del proyecto

**Referencia:** En tu boilerplate, `cmd/main.go` llama a `config.Load()`, después conecta DB, después crea servicios, después levanta el server. Acá es lo mismo pero sin DB ni servidor.

**Ejecutalo:**
```bash
go run ./cmd/check/
```

Si ves la configuración impresa en la terminal, esta fase está completa.

---

## Checklist de verificación

- [ ] `go mod init` ejecutado exitosamente
- [ ] `go build ./...` no da errores
- [ ] `go run ./cmd/check/` imprime la configuración
- [ ] `.gitignore` creado y funcional
- [ ] `.env.local.example` con TODAS las variables documentadas
- [ ] Entendés por qué `internal/` existe en Go (leé la spec)

---

## Lo que NO tenés que hacer todavía

- ❌ No escribas lógica de detección de fechas
- ❌ No conectes APIs externas
- ❌ No pienses en AWS
- ❌ No agregues dependencias que no necesites (empezá con solo `godotenv`)

Una vez que esto compile y corra, pasamos a la Fase 1.
