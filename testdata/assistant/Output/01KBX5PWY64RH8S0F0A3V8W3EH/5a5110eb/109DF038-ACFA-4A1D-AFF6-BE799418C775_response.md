---
provider: https://openrouter.ai/api/v1
model: anthropic/claude-haiku-4.5
duration: 2.595546959s
input: 853t
output: 123t
executed_at: 2025-12-09T21:48:43.133023+03:00
---

El punto de acceso de la API acepta solicitudes POST con un payload JSON que contiene los siguientes campos requeridos: "user_id" (cadena de texto), "timestamp" (formato ISO 8601) y "event_type" (enumeración: "click", "view", "purchase"). Los parámetros opcionales incluyen "metadata" (objeto) y "session_id" (cadena de texto). El límite de velocidad está configurado en 100 solicitudes por minuto por clave de API.