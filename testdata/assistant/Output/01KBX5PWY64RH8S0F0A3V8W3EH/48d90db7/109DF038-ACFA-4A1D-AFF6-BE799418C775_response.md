---
provider: https://openrouter.ai/api/v1
model: openai/gpt-4o-mini
duration: 2.727021709s
input: 724t
output: 100t
executed_at: 2025-12-09T21:46:42.653629+03:00
---

El endpoint de la API acepta solicitudes POST con una carga útil en formato JSON que contiene los siguientes campos obligatorios: "user_id" (cadena de texto), "timestamp" (formato ISO 8601) y "event_type" (enumeración: "click", "view", "purchase"). Los parámetros opcionales incluyen "metadata" (objeto) y "session_id" (cadena de texto). El límite de tasa está establecido en 100 solicitudes por minuto por clave de API.