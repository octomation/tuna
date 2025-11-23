---
provider: https://openrouter.ai/api/v1
model: anthropic/claude-haiku-4.5
duration: 2.60s
input: 853t
output: 120t
executed_at: 2025-12-07T23:59:40.256076+03:00
rating: null
rated_at: null
---

El endpoint de la API acepta solicitudes POST con un payload JSON que contiene los siguientes campos obligatorios: "user_id" (cadena de texto), "timestamp" (formato ISO 8601) y "event_type" (enumeración: "click", "view", "purchase"). Los parámetros opcionales incluyen "metadata" (objeto) y "session_id" (cadena de texto). El límite de velocidad está configurado en 100 solicitudes por minuto por clave de API.