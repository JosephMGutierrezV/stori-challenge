# 🚀 Pruebas de integración con API Gateway (Postman)

Esta colección permite validar el flujo completo entre **API Gateway**, **Lambda**, y **S3**.

## 🧭 Archivos incluidos

- **stori_apigw_collection.json** — colección de peticiones (importar en Postman o ejecutar con `newman`).
- **stori_env.json** — archivo de entorno con variables como `base_url` o `bucket_name`.

## ⚙️ Variables esperadas

| Variable      | Ejemplo                                              | Descripción                           |
|---------------|------------------------------------------------------|---------------------------------------|
| `base_url`    | `https://abc123.execute-api.us-east-1.amazonaws.com` | URL del API Gateway                   |
| `bucket_name` | `stori-transactions-dev`                             | Bucket destino para los CSV           |
| `api_key`     | _(opcional)_                                         | Clave si el Gateway usa autenticación |

## 💡 Uso rápido

1. Despliega la infraestructura (Lambda + API Gateway + S3).
2. Copia la URL base del API Gateway.
3. En Postman:

- Importa la colección y el entorno.
- Ajusta las variables (`base_url`, etc.).
- Envía la petición **POST /upload** adjuntando un archivo CSV válido.

Ejemplo de CSV válido:

```
Id,Date,Transaction
0,7/15,+60.5
1,7/28,-10.3
```

## 🧪 Ejecución automática (CLI)

Si prefieres probar desde terminal:

```bash
npx newman run tests/postman/stori_apigw_collection.json   -e tests/postman/stori_env.json   --reporters cli
```

---

**Autor:** Joseph Gutiérrez  
**Propósito:** Validar conexión funcional API Gateway → Lambda → S3.
