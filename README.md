# 💳 Stori Challenge – Joseph Mauricio Gutiérrez Valero

### 🧠 Descripción

Este proyecto fue desarrollado como solución al **Stori Software Engineer Technical Challenge**.  
Procesa un archivo CSV con transacciones de crédito y débito, calcula el resumen mensual y envía un **correo electrónico con la información consolidada**.

Está diseñado bajo **arquitectura hexagonal**, siguiendo **principios de TDD**, y preparado para ejecutarse tanto en **Docker** como en **AWS Lambda**.

---

## 🏗️ Estructura del Proyecto

```text
stori/
├── cmd/
│   ├── lambda_api/                 # Entrypoint para AWS Lambda (API Gateway)
│   │   └── main.go
│   └── local_runner/               # Entrypoint local (CLI o Docker)
│
├── configs/
│   └── .env                        # Variables de entorno locales
│
├── deployments/
│   ├── docker/                     # Dockerfile y docker-compose.yml
│   └── aws/                        # Template SAM o Terraform
│
├── internal/
│   ├── core/                       # Dominio puro (entidades y reglas de negocio)
│   │   ├── domain/                 # Entidades y objetos de valor
│   │   │   ├── transaction.go
│   │   │   ├── account.go
│   │   │   └── errors.go
│   │   ├── application/            # Casos de uso (servicios)
│   │   │   └── summary_service.go
│   │   └── ports/                  # Interfaces (puertos IN/OUT)
│   │       ├── in/
│   │       │   └── summary_port.go
│   │       └── out/
│   │           ├── email_sender.go
│   │           └── transaction_repo.go
│   │
│   ├── interfaces/                 # Adaptadores (entrada/salida)
│   │   ├── in/
│   │   │   ├── lambdahandler/      # Adaptador para AWS Lambda
│   │   │   └── cli/                # Adaptador CLI
│   │   └── out/
│   │       ├── csvreader/          # Lector de archivos CSV
│   │       ├── email/              # Envío de correos (SMTP / SES)
│   │       ├── persistence/        # Persistencia (ORM / DynamoDB / RDS)
│   │       │   ├── rds/            # Adaptador GORM / PostgreSQL
│   │       │   │   ├── mappers/    # Mapeo entre entidades y modelos GORM
│   │       │   │   └── models/     # Modelos GORM con tags
│   │       │   └── dynamo/         # (opcional) DynamoDB Adapter
│   │       └── notifier/           # SNS / Email Notifications
│   │
│   ├── infra/                      # Configuración e infraestructura
│   │   ├── aws/                    # Clientes AWS (S3, SES, DynamoDB)
│   │   ├── bootstrap/              # Wiring de dependencias
│   │   ├── config/                 # Carga de configuración (Viper)
│   │   └── logger/                 # Logging centralizado
│   │
│   └── shared/                     # Utilidades puras (sin dependencias externas)
│       └── uuid.go
│
├── test/                           # Tests unitarios y de integración
│
├── transactions.csv                # Archivo CSV de ejemplo
│
├── .dockerignore
├── .gitignore
├── Dockerfile                      # Construcción del contenedor
├── docker-compose.yml              # Orquestación local con DB y app
├── Makefile                        # Comandos automatizados (build, test, run)
├── go.mod                          # Dependencias Go
└── README.md                       # Documentación principal
```

---

## ⚙️ Requisitos del Challenge

| Requisito               | Descripción                                | Estado |
|-------------------------|--------------------------------------------|--------|
| 📊 **Procesar CSV**     | Lee transacciones de crédito y débito      | ✅      |
| 💰 **Calcular resumen** | Balance total, totales por mes y promedios | ✅      |
| 📧 **Enviar email**     | Envía resumen con formato y logo Stori     | ✅      |
| 💾 **Guardar datos**    | Persistencia con GORM / PostgreSQL         | ✅      |
| ☁️ **Cloud Ready**      | Compatible con AWS Lambda + SES + S3       | ✅      |

---

## 🧩 Ejemplo de CSV de Entrada

```csv
date,transaction
2021-07-15,+60.5
2021-07-20,-20.46
2021-08-10,+10.0
2021-08-15,-10.3
```

---

## 📬 Ejemplo de Resumen Enviado

```
💳 Account Summary

Total balance is 39.74

Number of transactions in July: 2
Number of transactions in August: 2

Average debit amount: -15.38
Average credit amount: 35.25
```

---

## 🧰 Tecnologías Principales

| Categoría       | Herramienta             |
|-----------------|-------------------------|
| Lenguaje        | Go (1.22)               |
| ORM             | GORM                    |
| Infraestructura | AWS Lambda, S3, SES     |
| Configuración   | Viper                   |
| Base de datos   | PostgreSQL              |
| Testing         | Go `testing` + mocks    |
| Contenedores    | Docker / docker-compose |
| Build / CI      | Makefile                |
| Estilo          | gofumpt + golangci-lint |

---

# 🧪 Tutorial: Prueba Local con LocalStack

Este tutorial muestra cómo levantar y probar el flujo completo del proyecto **Stori Challenge** localmente, sin usar recursos reales de AWS.  
Podrás simular un evento S3, ejecutar la Lambda y verificar los resultados en una base de datos PostgreSQL.

---

## 1. Crear el archivo `txns.csv`

Crea un archivo llamado `txns.csv` en la raíz del proyecto con este contenido:

```csv
Id,Date,Transaction
0,7/15,+60.5
1,7/28,-10.3
2,8/2,-20.46
3,8/13,+10
4,8/14,+15.75
5,8/21,-5.25
6,8/30,+120
7,9/1,-40
8,9/10,+5.5
9,9/15,-12
```

---

## 2. Levantar el entorno Docker

El `docker-compose.yml` debe incluir los servicios:

- `localstack` (simula AWS)
- `postgres` (base de datos local)
- `stori-app` (tu Lambda como contenedor)

Ejecuta:

```bash
docker compose up -d
```

*(o `make compose-up` si tienes Makefile configurado)*

Verifica los contenedores activos:

```bash
docker ps
```

Debes ver algo como:
```
localstack
pg-local
stori-app
```

---

## 3. Configurar AWS CLI para usar LocalStack

No se necesitan recursos en la nube real.  
Mientras uses `--endpoint-url http://localhost:4566`, todos los comandos apuntan a LocalStack.

**Windows (PowerShell):**
```powershell
$env:AWS_ACCESS_KEY_ID="test"
$env:AWS_SECRET_ACCESS_KEY="test"
$env:AWS_DEFAULT_REGION="us-east-1"
function awslocal { aws --endpoint-url http://localhost:4566 @Args }
```

**macOS / Linux:**
```bash
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
alias awslocal='aws --endpoint-url http://localhost:4566'
```

---

## 4. Crear el bucket S3 y subir el archivo

Crea el bucket dentro de LocalStack:

```bash
awslocal s3 mb s3://stori-transactions-local
```

Sube el archivo:

```bash
awslocal s3 cp txns.csv s3://stori-transactions-local/input/txns.csv
```

Confirma que se subió correctamente:

```bash
awslocal s3 ls s3://stori-transactions-local/input/
```

---

## 5. Crear el evento `event.json`

Este archivo emula el evento que S3 enviaría a Lambda al subir el CSV.

```json
{
  "Records": [
    {
      "s3": {
        "bucket": { "name": "stori-transactions-local" },
        "object": { "key": "input/txns.csv" }
      }
    }
  ]
}
```

---

## 6. Invocar la Lambda manualmente

Si el contenedor de la Lambda expone `9001:8080`, ejecuta:

```bash
curl -Method Post "http://localhost:9001/2015-03-31/functions/function/invocations" `
  -ContentType "application/json" `
  -InFile "event.json"
```

Esto simula la invocación automática que haría AWS cuando S3 genera un evento.

---

## 7. Ver logs de ejecución

Consulta los logs para ver el flujo de procesamiento:

```bash
docker logs stori-app
```

Deberías encontrar mensajes de:
- Lectura del archivo desde S3
- Procesamiento de las transacciones
- Inserciones en la base de datos
- Posible envío de email simulado (SES local)

---

## 8. Validar en PostgreSQL

Conéctate al contenedor de Postgres (expuesto en `5434`):

```bash
psql "host=localhost port=5434 dbname=app user=app password=app"
```

Ejecuta:

```sql
\d
SELECT * FROM transactions;
```

Si ves los registros procesados, el flujo funciona correctamente.

---

## 9. Limpiar el entorno

Cuando quieras reiniciar todo:

```bash
docker compose down -v
```

En Windows:

```powershell
Remove-Item -Recurse -Force C:\docker-data\stori
```

---

## 🏁 Resultado

Con este tutorial podrás reproducir localmente el flujo:

```
CSV -> S3 (LocalStack) -> Lambda (contenedor) -> PostgreSQL (como RDS)
```

Sin tocar recursos reales de AWS. Ideal para pruebas de integración o desarrollo sin costo.
