# Product Requirement Document (PRD)

## Proyecto: `bark-sentinel` (Servicio Serverless de Automatización y Notificaciones)

---

## 1. Introducción y Objetivos

**`bark-sentinel`** es un servicio backend serverless desarrollado en **Go** y desplegado en **AWS**. Su objetivo principal es automatizar la coordinación logística de la alimentación natural de dos perros en un entorno familiar, eliminando la fricción de los olvidos diarios y optimizando la comunicación mensual con el proveedor de insumos (carnicero) sin depender de infraestructura local encendida.

El proyecto está diseñado bajo restricciones estrictas de eficiencia, portabilidad y nulo mantenimiento manual, sirviendo además como un proyecto técnico de alto valor para demostrar el dominio de Go y arquitectura Serverless.

---

## 2. El Problema y Puntos de Fricción

### El Contexto Diario

- Dos perros con dieta natural (arroz con vegetales, hueso carnudo para el Labrador de 8 años y menudencias para la perra pequeña).
- La comida se prepara el día anterior por la mamá del usuario.
- Los insumos se compran de forma mensual en el mercado local un sábado o domingo a primera hora (7:00 AM).
- El usuario maneja un horario estricto que inicia a las 5:30 AM con el primer paseo y alimentación de los perros.

### Los Puntos de Fricción Reales

1. **El Factor "Olvido de la Cena":** Si a la mamá se le olvida preparar el arroz la noche anterior y al usuario se le pasa recordarle antes de las 8:00 PM, el problema estalla a las 5:30 AM del día siguiente al no haber comida lista tras el paseo, generando estrés e improvisación.
2. **El Factor "Pedido de Fin de Semana":** Las compras se realizan el primer fin de semana del mes (o el último del mes anterior). Si el usuario olvida escribirle al carnicero por WhatsApp el viernes previo para apartar las 10 libras de hueso de pecho y menudencias, se arriesga a desabastecer la logística mensual.
3. **El Factor "Mensaje Robótico":** Escribir exactamente el mismo texto de WhatsApp todos los meses al carnicero puede hacer que el usuario sea percibido como un bot o generar interacciones impersonales.

---

## 3. Público Objetivo y Usuarios

- **Usuario Principal:** Gino (Fullstack Developer), quien interactúa con las alertas en sus canales de comunicación nativos (Celular vía WhatsApp y Correo vía Gmail).
- **Usuario Indirecto:** La mamá de Gino (co-responsable de la preparación) y el carnicero del mercado (receptor del pedido).

---

## 4. Alcance del Producto (Features)

### Feature 1: Módulo de Alerta Diaria ("Check de la Cocina")

- **Descripción:** El sistema debe emitir un recordatorio preventivo todas las noches para verificar el estado de la comida del día siguiente.
- **Criterios de Aceptación:**
- Debe ejecutarse automáticamente todos los días a las **8:00 PM** (Hora Colombia).
- Debe enviar un correo electrónico a través de **Gmail** y un mensaje de texto automático a **WhatsApp**.
- **Mensaje requerido:** _"Gino, ve a revisar la cocina. ¿Está listo el arroz de los perros?"_

### Feature 2: Módulo del Carnicero ("Planificador de Compras")

- **Descripción:** Un algoritmo basado en fechas que identifique si el fin de semana actual (sábado/domingo) corresponde al inicio de mes y genere la orden de compra anticipada.
- **Criterios de Aceptación:**
- Debe ejecutarse los **viernes por la tarde**.
- Debe evaluar mediante el paquete `time` de Go si el día de mañana (sábado) o pasado mañana (domingo) es el momento de mercar el mes.
- Si la condición se cumple, debe disparar la alerta al celular y correo de Gino.

### Feature 3: Motor de Plantillas "Anti-Bot"

- **Descripción:** Lógica encargada de alternar y formatear dinámicamente el mensaje de WhatsApp enviado al carnicero para mantener una comunicación natural.
- **Criterios de Aceptación:**
- Debe almacenar un _slice_ con al menos 3 variantes del mensaje real provisto por el usuario.
- Debe seleccionar de forma aleatoria (`math/rand`) una de las variantes en cada ejecución.
- **Variantes requeridas:**
- _Opción 1:_ _"Hola buenas, por fa para mañana me puede apartar 10 libras de hueso de pecho. Me confirmas"_
- _Opción 2:_ _"Hola buenos días cómo va todo? Por fa para mañana me puede apartar 10 libras de hueso de pecho. Me confirmas."_

---

## 5. Requerimientos Técnicos y Arquitectura

### Stack Tecnológico

- **Lenguaje de Programación:** Go (Golang) 1.20+.
- **Infraestructura:** AWS (Capa Gratuita).
- **AWS Lambda:** Para ejecutar el binario de Go de manera serverless (0% consumo local, encendido por milisegundos).
- **Amazon EventBridge (CloudWatch Events):** Gobernanza del tiempo mediante reglas tipo Cron (`cron(0 20 * * ? *)` para la alerta diaria).

- **Integraciones Externas:**
- **SMTP de Gmail:** Uso del paquete nativo `net/smtp` de Go para envíos de correo.
- **API de WhatsApp (Twilio Sandbox / UltraMsg):** Uso del paquete nativo `net/http` para realizar peticiones POST hacia el Gateway de WhatsApp.

### Seguridad y Configuración

- Las credenciales sensibles (Token de WhatsApp, Contraseña de aplicación de Gmail, IDs) **no deben** estar harcodeadas en el código. Se consumirán como **Variables de Entorno** configuradas de forma segura en la consola de AWS Lambda.

---

## 6. Fases de Desarrollo Técnicas (Plan de Trabajo)

```
[Fase 1: Core en Go] ---> [Fase 2: Conectores SMTP/HTTP] ---> [Fase 3: Implementación AWS]

```

- **Fase 1 (Local):** Inicialización del módulo de Go, desarrollo del algoritmo de detección de fines de semana del mes y el selector aleatorio de strings (plantillas del carnicero). Salida por consola de la terminal.
- **Fase 2 (Integración):** Configuración del cliente HTTP para WhatsApp y el cliente SMTP de Gmail. Pruebas de envío reales desde la máquina local hacia el celular/correo del usuario.
- **Fase 3 (Cloud):** Migración del código principal al handler de `aws-lambda-go`, compilación del binario para Linux de 64 bits, empaquetado en `.zip`, subida a AWS y configuración de los disparadores cron en EventBridge.

---

## 7. Restricciones y Exclusiones (Fuera de Alcance)

- **No** tendrá interfaz gráfica (Frontend/UI).
- **No** tendrá base de datos relacional (PostgreSQL/MySQL); la persistencia de las plantillas y números de teléfono se gestionará de manera estática en el código o mediante variables de entorno para simplificar la infraestructura.
- El sistema no automatizará el envío _directo_ al WhatsApp del carnicero; le enviará el texto pre-armado a Gino para que él mantenga el control final de la interacción con un simple copiar/pegar, evitando errores de software frente al proveedor.
