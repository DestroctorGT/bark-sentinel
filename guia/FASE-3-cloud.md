# FASE 3 — Cloud: AWS Lambda + EventBridge

## Objetivo

Migrar el programa local a AWS Lambda, compilarlo para Linux, empaquetarlo, subirlo a AWS, y configurar los disparadores cron con EventBridge para que se ejecute automáticamente.

---

## Lo que vas a aprender

- Cómo adaptar un programa Go a AWS Lambda con `aws-lambda-go`
- Compilación cruzada para Linux amd64 desde tu máquina
- Configuración de IAM roles con mínimos privilegios
- Reglas cron en EventBridge
- Variables de entorno en Lambda
- Debug de funciones Lambda con CloudWatch Logs

---

## Arquitectura final

```
EventBridge (cron) ──▶ Lambda (bark-sentinel) ──▶ WhatsApp API
                         │                           └─▶ Gino (celular)
                         └──▶ SMTP Gmail
                              └──▶ Gino (email)
```

**Dos reglas cron independientes:**
- `cron(0 1 * * ? *)` → 8:00 PM Colombia (UTC-5), todos los días → Recordatorio cocina
- `cron(0 18 * * FRI *)` → 1:00 PM Colombia viernes → Alerta carnicero (solo si aplica)

**Explicación de la conversión horaria:**
- Colombia está en UTC-5 todo el año (no tiene DST)
- 8:00 PM Colombia = 1:00 AM UTC del día SIGUIENTE
- Esperá, revisemos: 8 PM Colombia = 8 PM - 5 = 15 UTC? No.
- 8 PM = 20:00. 20:00 - 5 = 15:00 UTC. 
- Mejor pensalo así: cuando son las 8 PM en Colombia, son 1:00 AM UTC.
- 20:00 - 5 = 15:00... eso no es 1:00 AM.
- OK, 8 PM Colombia en UTC: 8 PM = 20:00. 20:00 + 5 = 01:00 (del día siguiente). 
- CORRECCIÓN: Colombia (UTC-5) → UTC = hora local + 5. 20:00 + 5 = 01:00 UTC del día siguiente.
- Para la regla cron de EventBridge: necesitás `cron(0 1 * * ? *)` — a la 1:00 AM UTC es 8:00 PM Colombia.
- Para el viernes a las 1 PM Colombia = 13:00 + 5 = 18:00 UTC. `cron(0 18 * * FRI *)`

**IMPORTANTE:** Verificá esta conversión antes de configurarla. Una hora mal puesta y el recordatorio llega a las 3 AM.

---

## Paso a paso

### 1. Agregar dependencia AWS Lambda

```bash
go get github.com/aws/aws-lambda-go/lambda
```

Esta es la única dependencia nueva que necesitás. `aws-lambda-go` es mantenido por AWS y es la forma oficial de escribir Lambda en Go.

**Concepto a investigar:** ¿Cómo funciona el handler de Lambda en Go? Leé la documentación de `aws-lambda-go`. La firma del handler puede ser `func()` o `func(context.Context)` o `func(context.Context, event)T` — todas son válidas.

### 2. Crear `cmd/lambda/main.go`

Este es el entry point para AWS Lambda. Es análogo a `cmd/check/main.go` pero con dos diferencias clave:

1. **No tiene loop infinito** — Lambda ejecuta el handler y termina
2. **No imprime a consola (solo logs)** — la salida de Lambda no es para humanos directos
3. **Usa `lambda.Start()`** en vez de ejecutar directamente

**Estructura del handler:**

```
func handler(ctx context.Context) error
```

Adentro de `handler`:
1. Llamá a `config.Load()` — las variables de entorno en Lambda se configuran en la consola de AWS
2. Determiná qué tipo de ejecución es:
   - Podés usar el `time.Now()` para detectar si hoy hay que mandar cocina, carnicero, ambos o ninguno
   - O podés recibir un event de EventBridge que te diga qué regla se disparó
3. Ejecutá los envíos correspondientes (igual que en `cmd/check/`)
4. Retorná `nil` si todo salió bien, o un error si algo falló

**¿Una función Lambda o dos?**

Tenés dos opciones:

**Opción A — Una Lambda que decide todo:**
- Un solo binario, una sola función
- Cuando se dispara, evalúa qué recordatorios corresponden hoy
- Ventaja: Simple, un solo deploy
- Desventaja: La Lambda de cocina se ejecuta todos los días aunque no haya carnicero (no es problema)

**Opción B — Dos Lambdas independientes:**
- Una para cocina, otra para carnicero
- Cada EventBridge llama a su Lambda específica
- Ventaja: Separación total de responsabilidades
- Desventaja: Dos deploys, dos roles IAM, más complejidad

**Recomendación:** Arrancá con la Opción A (una Lambda que decide todo). Podés pasar a la Opción B si necesitás. Es más simple y para dos alertas diarias no hay diferencia de costo.

**Referencia al boilerplate:** En `cmd/main.go` tenés el patrón de "cargar config → crear dependencias → ejecutar". Acá es lo mismo pero en vez de `http.ListenAndServe`, usás `lambda.Start()`.

### 3. Compilación cruzada para Lambda

Lambda corre en Linux amd64 (o ARM si elegís Graviton). Necesitás compilar para ese target.

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./cmd/lambda/
```

**Explicación de cada flag:**
- `GOOS=linux` — sistema operativo target
- `GOARCH=amd64` — arquitectura target (x86_64)
- `CGO_ENABLED=0` — desactiva CGO para tener un binario estático. **CRÍTICO** para Lambda porque el entorno no tiene las librerías C que esperaría tu compilador.
- `-o bootstrap` — el nombre del binario DEBE ser `bootstrap` para que Lambda lo reconozca (es el estándar de AWS para Go)

**Verificación:**
```bash
file bootstrap
# Debería decir: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked
```

**Problema común:** Si tenés alguna dependencia que usa CGO, no va a compilar con `CGO_ENABLED=0`. En ese caso, necesitás compilar en una máquina Linux o usar Docker. Para este proyecto, como solo usás `net/http`, `net/smtp`, `encoding/json` y `aws-lambda-go`, no deberías tener problemas. Todas son bibliotecas pure Go.

### 4. Empaquetar en ZIP

```bash
zip function.zip bootstrap
```

Eso es todo. Lambda para Go es solo el binario comprimido. No necesitas dependencias externas, runtime, ni nada más.

### 5. Configuración en AWS

#### a) Crear el rol IAM

Necesitás un rol que Lambda pueda asumir. Los permisos mínimos son:

- `AWSLambdaBasicExecutionRole` — para escribir logs a CloudWatch

**NO** le des más permisos de los necesarios. Este proyecto no toca S3, DynamoDB, ni ningún otro servicio de AWS.

**Para crear el rol:**
1. Andá a IAM > Roles > Create role
2. Trusted entity type: AWS service > Lambda
3. Adjuntá la política `AWSLambdaBasicExecutionRole`
4. Namedlo algo como `bark-sentinel-role`

#### b) Crear la función Lambda

1. Andá a Lambda > Create function
2. Author from scratch
3. Runtime: **Amazon Linux 2023** (no necesitás runtime específico porque el binario Go es auto-contenido)
4. Architecture: x86_64 (o ARM si compilaste para ARM)
5. Role: elegí el rol que creaste
6. Subí el ZIP (`function.zip`)
7. Handler: dejalo en blanco o como `bootstrap` (Go no necesita handler específico)

#### c) Configurar el timeout

Por defecto Lambda timeout es 3 segundos. Para este proyecto necesitás más:

- Configurá timeout en **30 segundos** (las llamadas HTTP a WhatsApp y SMTP pueden demorar)
- Si usás VPC, el timeout necesita ser mayor (pero no necesitás VPC para este proyecto)

#### d) Configurar las variables de entorno

En la consola de Lambda, andá a Configuration > Environment variables y agregá:

```
APP_ENV=production
WHATSAPP_TOKEN=<token real>
WHATSAPP_PHONE_NUMBER=<número de Gino>
BUTCHER_PHONE_NUMBER=<número del carnicero>
GMAIL_USER=<tu-email@gmail.com>
GMAIL_PASSWORD=<contraseña de aplicación>
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
```

**Consejo de seguridad:** Para producción considerá usar AWS Secrets Manager o SSM Parameter Store para las credenciales. Pero para este proyecto, variables de entorno en Lambda están bien — es un proyecto personal, el riesgo es bajo, y la simplicidad es el objetivo.

### 6. Configurar EventBridge (CloudWatch Events)

Necesitás crear DOS reglas.

#### Regla 1: Recordatorio de cocina (diario a las 8 PM Colombia)

1. Andá a EventBridge > Rules > Create rule
2. Name: `bark-sentinel-cocina-daily`
3. Rule type: Schedule
4. Schedule pattern: Cron expression
5. Cron: `0 1 * * ? *` (1 AM UTC = 8 PM Colombia)
6. Target: Seleccioná la función Lambda `bark-sentinel`
7. Create

**Probá la expresión cron:** Poné una fecha cercana y verificá en la consola que se muestre la próxima ejecución.

#### Regla 2: Alerta de carnicero (viernes a la 1 PM Colombia)

1. Name: `bark-sentinel-butcher-friday`
2. Schedule pattern: Cron expression
3. Cron: `0 18 * * FRI *` (6 PM UTC = 1 PM Colombia)
4. Target: misma función Lambda

**Explicación:** La Lambda recibe el evento, pero como es la misma función, ella misma decide qué hacer según la fecha actual. No necesita saber qué regla la disparó.

### 7. Probar la Lambda

#### Prueba manual desde la consola:

1. Andá a tu función Lambda en la consola
2. Test > Create new test event
3. Elegí "CloudWatch" como event template (opcional — nuestro handler no usa el event)
4. Ejecutá el test
5. Andá a CloudWatch Logs para ver la salida

#### Prueba real con EventBridge:

- Creá una regla temporal que se dispare en 5 minutos para verificar que el cron funciona
- O usá el botón "Test" de Lambda con un event simulado

**Qué esperar en los logs de CloudWatch:**
```
[COCINA] Recordatorio enviado por WhatsApp ✅
[COCINA] Recordatorio enviado por Email ✅
[BUTCHER] Hoy no es finde de compras. Omitido.
```

### 8. Debugging remoto

Cuando algo falle en Lambda (y va a fallar), la estrategia de debugging es:

1. **Revisá CloudWatch Logs** — todo lo que imprimís con `log.Println` aparece ahí
2. **Agregá más logs** si hace falta — no hay debugger remoto
3. **Simulá el entorno Lambda localmente** con el evento que recibís
4. **Verificá las variables de entorno** — es la causa más común de fallos

**Herramienta útil:** AWS Lambda tiene un "console emulator" que podés activar para ver errores de arranque. Si la Lambda falla antes de ejecutar tu handler, el único log que ves es "Task timed out" o "Task exited without message".

---

## Verificación final

1. **Compilación cruzada exitosa:**
   ```bash
   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./cmd/lambda/
   ```
   → Sin errores, archivo `bootstrap` generado

2. **Binario verificado como estático:**
   ```bash
   file bootstrap | grep "statically linked"
   ```

3. **Prueba manual en Lambda desde consola:**
   → Ejecución exitosa, logs en CloudWatch

4. **EventBridge configurado correctamente:**
   → Revisá la consola de EventBridge > Rules > bark-sentinel-* > ver próxima ejecución

5. **Recibís los mensajes reales:**
   → WhatsApp y email llegan a Gino

---

## Lo que NO tenés que hacer

- ❌ No agregues API Gateway (no necesitás HTTP público)
- ❌ No configures VPC (no necesitás acceso a recursos privados)
- ❌ No uses SAM/CloudFormation/CDK todavía — hacelo manual primero para entender cada paso
- ❌ No configures alarms de CloudWatch (es overkill para un proyecto personal)

---

## Para reflexionar sobre el proyecto completo

1. **¿Por qué funciona sin base de datos?** Porque la lógica es puramente algorítmica (fechas) y las plantillas son estáticas. No hay estado que persistir.
2. **¿Cuánto tiempo corre la Lambda por día?** Menos de 1 segundo. Multiplicado por 2 ejecuciones diarias = inmensamente dentro de la capa gratuita (400,000 GB-segundos por mes).
3. **¿Qué tan caro es?** AWS Lambda tiene 1 millón de ejecuciones gratis por mes. Este proyecto hace 60 ejecuciones por mes. Literalmente cuesta $0.00.

---

## Bonus: Lo que aprendiste (y deberías poder explicar sin leer el código)

- ☐ Cómo inicializar un módulo Go y estructurar un proyecto real
- ☐ Cómo cargar configuración desde variables de entorno con godotenv
- ☐ Cómo trabajar con fechas y días de semana usando `time` package
- ☐ Cómo seleccionar elementos aleatorios de un slice
- ☐ Cómo hacer requests HTTP POST con `net/http`
- ☐ Cómo enviar emails con `net/smtp`
- ☐ Cómo compilar un binario Go para una plataforma diferente
- ☐ Cómo empaquetar y subir un binario a AWS Lambda
- ☐ Cómo configurar reglas cron con EventBridge
- ☐ Cómo debuggear funciones Lambda con CloudWatch Logs

Si podés explicar cada uno de estos puntos sin leer el código, el proyecto cumplió su objetivo.
