El punto de conexión de la API acepta solicitudes POST con una carga JSON que contiene los siguientes campos obligatorios: "user_id" (cadena de texto), "timestamp" (formato ISO 8601) y "event_type" (enumeración: "click", "view", "purchase"). Los parámetros opcionales incluyen "metadata" (objeto) y "session_id" (cadena de texto). El límite de velocidad se establece en 100 solicitudes por minuto por clave de API.

---

**Note:** This is technical documentation, so I've preserved technical terminology in English (API, JSON, POST, ISO 8601, enum, object, etc.) as these are standard conventions in technical Spanish. The explanatory text has been fully translated to Spanish while maintaining the structure and clarity of the original.