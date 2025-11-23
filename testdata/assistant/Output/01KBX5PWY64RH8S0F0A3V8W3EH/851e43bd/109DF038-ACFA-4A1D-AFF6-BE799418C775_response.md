---
provider: https://openrouter.ai/api/v1
model: google/gemini-2.5-flash
duration: 864.135458ms
input: 794t
output: 108t
executed_at: 2025-12-09T21:47:41.401642+03:00
---

El endpoint de la API acepta solicitudes POST con un payload JSON que contiene los siguientes campos obligatorios: "user_id" (cadena de texto), "timestamp" (formato ISO 8601) y "event_type" (enumeración: "click", "view", "purchase"). Los parámetros opcionales incluyen "metadata" (objeto) y "session_id" (cadena de texto). La limitación de la tasa está establecida en 100 solicitudes por minuto por clave de API.