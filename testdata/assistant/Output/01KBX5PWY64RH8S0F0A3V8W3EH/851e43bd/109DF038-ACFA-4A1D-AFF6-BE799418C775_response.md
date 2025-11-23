---
provider: https://openrouter.ai/api/v1
model: google/gemini-2.5-flash
duration: 811ms
input: 794t
output: 103t
executed_at: 2025-12-07T23:58:38.471681+03:00
rating: null
rated_at: null
---

El endpoint de la API acepta solicitudes POST con una carga útil JSON que contiene los siguientes campos obligatorios: "user_id" (cadena), "timestamp" (formato ISO 8601) y "event_type" (enumeración: "click", "view", "purchase"). Los parámetros opcionales incluyen "metadata" (objeto) y "session_id" (cadena). El límite de tasa está establecido en 100 solicitudes por minuto por clave de API.