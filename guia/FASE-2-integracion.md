# FASE 2 — Integración: WhatsApp y Email (Local)

## Objetivo

Conectar el core lógico de la Fase 1 con servicios externos reales: enviar WhatsApp vía API HTTP y enviar emails vía SMTP de Gmail. Todo sigue corriendo localmente.

---

## Lo que vas a aprender

- Paquete `net/http` para hacer POST requests a una API REST
- Paquete `net/smtp` para envío de correos electrónicos
- Cómo estructurar un "notificador" que abstrae el canal de comunicación
- Manejo de errores en operaciones de red
- Cómo probar integraciones reales sin hacer mock de todo

---

## Antes de arrancar: Configuración de servicios reales

Necesitás credenciales reales. No hay forma de esquivarlo.

### WhatsApp

El PRD menciona dos opciones: Twilio Sandbox o UltraMsg. Elegí la que quieras.

**Independientemente de cuál elijas, el patrón es el mismo:**
1. Hacés un POST a una URL con un token de autenticación
2. Mandás un JSON con el número destino y el mensaje
3. La API se encarga de entregar el WhatsApp

**Requerimientos:**
- Crear cuenta en el proveedor que elijas
- Obtener un token/API key
- Configurar el número de teléfono de Gino como destino
- Verificar que puedas enviar un mensaje de prueba desde el dashboard web del proveedor (antes de escribir código)

### Gmail SMTP

Google requiere una "Contraseña de Aplicación" para usar SMTP desde código.

**Pasos necesarios:**
1. Activá la verificación en dos pasos en tu cuenta de Gmail (si no la tenés)
2. Generá una contraseña de aplicación desde la configuración de seguridad de Google
3. Guardá esa contraseña en tu `.env.local` — **nunca la commitees**

**Importante:** La contraseña de aplicación no es la misma que tu contraseña de Gmail. Es una contraseña de 16 caracteres que Google genera específicamente para apps.

---

## Paso a paso

### 1. Crear `internal/notifier/whatsapp.go`

Este archivo encapsula TODO el envío de WhatsApp. Nadie fuera de este paquete necesita saber qué proveedor usás ni cómo se llama la API.

**Lo que necesitás implementar:**

1. Una función/clase `WhatsAppClient` que reciba la configuración (token, números) en su constructor
   - **Referencia:** Mirá cómo en tu boilerplate `internal/auth/service.go` recibe `Repository` y `JWTConfig` en el constructor. Mismo patrón.
2. Un método `SendMessage(to, message string) error` que:
   - Arme un payload JSON con el número destino y el texto
   - Haga un POST HTTP a la URL del proveedor
   - Lea la respuesta y verifique que sea exitosa
   - Retorne error si algo falla

**Detalles técnicos:**

- Usá `net/http` para el request. No necesitas ninguna librería externa.
- Usá `encoding/json` para marshalling del body
- Configurá los headers: `Content-Type: application/json` y el header de autorización que tu proveedor requiera
- **Edge case:** ¿Qué pasa si la API responde con un error HTTP (4xx/5xx)? No te comas el error — propagalo con contexto.
- **Edge case:** ¿Qué pasa si hay un timeout de red? Por defecto `http.Client` no tiene timeout. Configurá uno explícito.

**Testing local:**
Antes de integrarlo con el resto del sistema, escribí un pequeño programa de prueba temporal que llame a esta función con un mensaje fijo. Si te llega el WhatsApp, sabés que funciona.

---

### 2. Crear `internal/notifier/email.go`

Este archivo encapsula el envío de emails vía SMTP.

**Lo que necesitás implementar:**

1. Una función/clase `EmailClient` que reciba la configuración SMTP (usuario, password, host, puerto)
2. Un método `Send(to, subject, body string) error` que:
   - Construya el mensaje en formato MIME (To, Subject, From, cuerpo)
   - Use `net/smtp.SendMail()` o `net/smtp.PlainAuth` + `net/smtp.SendMail`
   - Retorne error si falla

**La trampa de SMTP:**

A diferencia de una API REST que recibe JSON, SMTP tiene su propio protocolo. El paquete `net/smtp` te abstrae la parte complicada, pero tenés que armar el mensaje en el formato correcto.

**El mensaje MIME tiene esta estructura:**
```
From: remitente@gmail.com
To: destinatario@gmail.com
Subject: Asunto del correo
MIME-Version: 1.0
Content-Type: text/plain; charset="UTF-8"

Cuerpo del mensaje
```

**Problema clásico:** Si no ponés los headers `From`, `To` y `Subject` como parte del cuerpo del mensaje (no como parámetros de la función `SendMail`), el correo llega sin asunto. Investigá cómo funciona `net/smtp.SendMail`.

**Testing local:**
Mandate un email de prueba a vos mismo. Si te llega, funciona.

---

### 3. Integrar en el orquestador

Ahora que tenés los dos clientes de notificación, necesitás actualizar `cmd/check/main.go` para que:

1. Cree una instancia de `EmailClient` y `WhatsAppClient` con la configuración cargada
2. Después de la lógica de `kitchen` y `butcher`, **en lugar de imprimir por consola**, llame a los notificadores correspondientes:

**Reglas de envío:**

| Evento | WhatsApp | Email |
|--------|----------|-------|
| Recordatorio diario cocina | ✅ A Gino | ✅ A Gino |
| Alerta de pedido al carnicero | ✅ A Gino (con el texto pre-armado) | ✅ A Gino |

**Nota importante del PRD (Sección 7):** El sistema NO envía directamente al carnicero. Le envía el texto PRE-ARMADO a Gino para que él haga el copiar/pegar manual. Esto evita errores de software frente al proveedor. Respetá esta decisión.

---

### 4. Manejo de errores

Cuando falla un envío (y va a fallar — es una integración real), necesitás:

1. **NO detener todo el programa** porque falle el WhatsApp si el email puede enviarse
2. **LOGEAR el error** con contexto suficiente para debuguear
3. **Diferenciar** entre error temporal (timeout, red) y error permanente (token inválido)

**Implementación sugerida:**
- Hacé que los envíos sean paralelos o secuenciales independientes (no en cadena)
- Si uno falla, loguealo y seguí con el otro
- Al final del programa, mostrá un resumen: "WhatsApp: ✅, Email: ❌ (timeout)"

---

## Verificación

1. **Ejecutá el programa completo** y verificá que te llegue el mensaje de WhatsApp
2. **Verificá que el email llegue** a tu bandeja de entrada (revisá spam si no llega)
3. **Desconectá internet temporalmente** y verificá que el error se maneje graceful (no panic, no crash)
4. **Modificá el token de WhatsApp** a uno inválido y verificá que el error sea claro
5. **Corré el programa múltiples veces** en diferentes días (o simulando diferentes fechas) para verificar el detector de findes de compra

```bash
go run ./cmd/check/
```

---

## Lo que NO tenés que hacer todavía

- ❌ No subas nada a AWS
- ❌ No optimices nada — el código local no necesita ser eficiente a nivel Lambda
- ❌ No agregues colas de reintento ni backoff
- ❌ No te preocupes por secretos— las credenciales están en `.env.local` que está en `.gitignore`

---

## Para reflexionar antes de pasar a Fase 3

1. ¿Por qué separé `whatsapp.go` de `email.go` en vez de tener un solo `Notifier` que maneje ambos?
   - **Respuesta:** Porque cambian por diferentes razones. WhatsApp usa HTTP REST, Email usa SMTP. Si mañana cambiás de proveedor de WhatsApp, solo tocás `whatsapp.go`.
2. ¿Por qué los mensajes se envían secuencialmente y no en paralelo con goroutines?
   - **Respuesta:** Por ahora no hace falta. Son 2-4 mensajes, no 10,000. Cuando sea necesario, agregás concurrencia.
3. ¿Por qué el logger es `log.Println` y no un paquete fancy?
   - **Respuesta:** Porque para Lambda, los logs van a CloudWatch. `log.Println` escribe a stdout/stderr, que CloudWatch captura automáticamente. Cero dependencias innecesarias.

Checklist:
- [ ] WhatsAppClient creado con constructor que recibe configuración
- [ ] `SendMessage(to, message string) error` implementado
- [ ] Mensaje de prueba enviado y recibido exitosamente
- [ ] EmailClient creado con constructor que recibe configuración
- [ ] `Send(to, subject, body string) error` implementado
- [ ] Email de prueba enviado y recibido exitosamente
- [ ] `cmd/check/main.go` orquesta envíos según las reglas de negocio
- [ ] Manejo de errores: si un canal falla, el otro sigue
- [ ] Errores de red simulados no crashean el programa
