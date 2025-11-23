---
provider: https://openrouter.ai/api/v1
model: openai/gpt-4o-mini
duration: 2.79s
input: 724t
output: 96t
executed_at: 2025-12-07T23:57:40.448059+03:00
rating: null
rated_at: null
---

El punto final de la API acepta solicitudes POST con una carga útil en formato JSON que contiene los siguientes campos obligatorios: "user_id" (cadena), "timestamp" (formato ISO 8601) y "event_type" (enum: "click", "view", "purchase"). Los parámetros opcionales incluyen "metadata" (objeto) y "session_id" (cadena). El límite de tasa está establecido en 100 solicitudes por minuto por clave de API.