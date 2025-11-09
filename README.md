# 💳 Stori Challenge – Joseph Mauricio Gutiérrez Valero

### 🧠 Descripción general

Este repositorio contiene la solución al **Stori Software Engineer Technical Challenge**.

La Lambda principal:

- Lee un archivo **CSV** con transacciones de crédito y débito desde **S3**.
- Procesa las transacciones y calcula:
    - Balance total de la cuenta.
    - Número de transacciones agrupadas por mes.
    - Promedio de montos de **créditos** y **débitos** agrupados por mes.
- Persiste la información en **PostgreSQL**.
- Envía un **correo electrónico** con el resumen, usando **SES**, con:
    - Logo de Stori.
    - Tabla de resumen mensual.

El proyecto está diseñado con:

- **Arquitectura hexagonal (ports & adapters)**.
- Enfoque hacia **TDD** (tests de dominio, servicios y adaptadores).
- Ejecución tanto en:
    - Local con **Docker + docker-compose + LocalStack**.
    - Infraestructura real en AWS usando **Terraform**.

Además, existe una segunda Lambda (en otro repo) que, expuesta vía **API Gateway**, recibe un archivo, valida que sea un
CSV correcto y lo sube al bucket S3 para que dispare esta Lambda de procesamiento.

---

## 🏗️ Estructura del proyecto

```text
📁 stori-challenge
├── cmd/
│   └── lambda_api/
│       └── main.go                # Entrypoint Lambda (S3Event → SummaryService)
├── configs/
│   └── .env                       # Configuración local (variables de entorno)
├── deployments/
│   └── terraform/                 # Infraestructura como código (AWS)
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       └── terraform.tfvars
├── docs/
│   └── api/
│       ├── README_API.md          # Documentación de la API (segunda Lambda)
│       └── postman/
│           └── stori-api.postman_collection.json  # Colección para API Gateway
├── internal/
│   ├── core/                      # Dominio puro + casos de uso
│   │   ├── application/
│   │   │   ├── summary_service.go
│   │   │   └── summary_service_test.go
│   │   ├── domain/
│   │   │   ├── transaction.go
│   │   │   └── summary.go
│   │   └── ports/
│   │       └── in/
│   │           └── summary_port.go
│   ├── infra/                     # Infraestructura y cross-cutting concerns
│   │   ├── aws/
│   │   │   └── s3client/
│   │   │       └── s3client.go
│   │   ├── bootstrap/
│   │   │   ├── bootstrap.go
│   │   │   └── bootstrap_integration_test.go
│   │   ├── config/
│   │   │   ├── config.go
│   │   │   └── config_test.go
│   │   ├── database/
│   │   │   └── postgres.go
│   │   └── logger/
│   │       └── logger.go
│   └── interfaces/                # Adaptadores (S3, SES, RDS, etc.)
│       ├── out/
│       │   ├── csvreader/
│       │   ├── email/
│       │   └── rds/
│       └── in/
│           └── (futuros handlers API/CLI)
├── migrations/
│   ├── 0001_create_schema_transactions.up.sql
│   ├── 0001_create_schema_transactions.down.sql
│   ├── 0002_create_transactions_table.up.sql
│   └── 0002_create_transactions_table.down.sql
├── .dockerignore
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── event.json                     # Ejemplo de evento S3 para pruebas locales
├── txns.csv                       # Ejemplo de CSV de entrada
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## ⚙️ Requisitos del challenge

| Requisito           | Descripción                                                        | Estado |
|---------------------|--------------------------------------------------------------------|--------|
| 📊 Procesar CSV     | Lee transacciones de crédito y débito desde un archivo CSV         | ✅      |
| 💰 Calcular resumen | Balance total, resumen por mes y promedios de crédito/débito       | ✅      |
| 📧 Enviar email     | Envía un correo con formato, tabla y logo de Stori vía SES         | ✅      |
| 💾 Guardar datos    | Persiste transacciones y resumen usando GORM + PostgreSQL          | ✅      |
| ☁️ Cloud Ready      | Compatible con AWS Lambda + S3 + SES + RDS (Terraform)             | ✅      |
| 🧪 Pruebas          | Tests de dominio, servicio, adaptadores (incluidos de integración) | ✅      |

---

## 🧩 Ejemplo de CSV de entrada

Formato esperado:

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

- `Id`: un identificador de la fila (no se usa en el cálculo del resumen, pero se valida la estructura).
- `Date`: fecha en formato `M/D` (por ejemplo `7/15`).
- `Transaction`: monto con signo `+` o `-`.

---

## 📬 Ejemplo del resumen enviado por email

Versión **texto plano** (body de respaldo):

```text
Account Summary

Total balance is 39.74

Number of transactions in July: 2
Number of transactions in August: 2

Average debit amount in July: -15.38
Average credit amount in July: 35.25
Average debit amount in August: -10.00
Average credit amount in August: 10.00
```

La versión **HTML** incluye:

- Logo de Stori (configurable por `STORI_LOGO_URL`).
- Colores de marca (tonos verdes).
- Tarjeta con:
    - Balance total.
    - Tabla con resumen por mes (`mes`, `# transacciones`, `avg debit`, `avg credit`).
- Mensaje de aviso al usuario.

---

## 🧰 Tecnologías principales

| Categoría       | Herramienta / Librería                     |
|-----------------|--------------------------------------------|
| Lenguaje        | Go 1.22                                    |
| Arquitectura    | Hexagonal (ports & adapters)               |
| ORM             | GORM                                       |
| Base de datos   | PostgreSQL                                 |
| Cloud           | AWS Lambda, S3, SES, RDS                   |
| Infraestructura | Terraform                                  |
| Configuración   | Viper                                      |
| Logs            | Uber Zap                                   |
| Testing         | Go `testing`, fakes y tests de integración |
| Contenedores    | Docker, docker-compose                     |
| Local Cloud     | LocalStack                                 |
| Build / CI      | Makefile                                   |
| Estilo Go       | gofumpt, golangci-lint                     |

---

# 🧪 Tutorial: prueba local con Docker + LocalStack

Este tutorial te guía para probar el flujo completo **sin tocar AWS real**:

`CSV → S3 (LocalStack) → Lambda (contenedor) → PostgreSQL`

### 1. Crear el archivo `txns.csv`

En la raíz del proyecto:

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

### 2. Levantar el entorno con docker-compose

El `docker-compose.yml` levanta:

- `localstack` → simula S3, SES (limitado), etc.
- `pg-local` → PostgreSQL local.
- `stori-app` → la imagen Lambda corriendo con `aws-lambda-runtime` en modo contenedor.

Desde la raíz del repo:

**Windows (PowerShell):**

```powershell
make compose-up
# o
docker compose up -d
```

**macOS / Linux:**

```bash
make compose-up
# o
docker compose up -d
```

Verifica los contenedores:

```bash
docker ps
```

---

### 3. Configurar AWS CLI para hablar con LocalStack

La clave: mientras uses `--endpoint-url http://localhost:4566`, todo va contra LocalStack.

**Windows (PowerShell):**

```powershell
$env:AWS_ACCESS_KEY_ID="test"
$env:AWS_SECRET_ACCESS_KEY="test"
$env:AWS_DEFAULT_REGION="us-east-1"

function awslocal { aws --endpoint-url http://localhost:4566 @Args }
```

**macOS / Linux (bash/zsh):**

```bash
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1

awslocal() {
  aws --endpoint-url http://localhost:4566 "$@"
}
```

---

### 4. Crear el bucket S3 y subir el CSV

```bash
awslocal s3 mb s3://stori-transactions-local

awslocal s3 cp txns.csv s3://stori-transactions-local/input/txns.csv

awslocal s3 ls s3://stori-transactions-local/input/
```

Si ves `txns.csv` listado, está todo bien.

---

### 5. Crear el archivo `event.json`

```json
{
  "Records": [
    {
      "s3": {
        "bucket": {
          "name": "stori-transactions-local"
        },
        "object": {
          "key": "input/txns.csv"
        }
      }
    }
  ]
}
```

---

### 6. Invocar la Lambda localmente

El contenedor de la Lambda suele exponer `9001:8080`.

**Windows (PowerShell):**

```powershell
curl -Method Post "http://localhost:9001/2015-03-31/functions/function/invocations" `
  -ContentType "application/json" `
  -InFile "event.json"
```

**macOS / Linux:**

```bash
curl -X POST "http://localhost:9001/2015-03-31/functions/function/invocations" \
  -H "Content-Type: application/json" \
  -d @event.json
```

Esto simula el evento que dispara S3 en AWS.

---

### 7. Ver logs de la Lambda

```bash
docker logs stori-app
```

Ahí deberías ver:

- El evento S3 recibido.
- Lectura de `input/txns.csv` desde S3.
- Cálculo del resumen.
- Inserciones en DB.
- Intento de envío de email:
    - En AWS real: SES v2.
    - En LocalStack: en este challenge se usa un **NoopEmailSender** cuando `AWS_ENDPOINT_URL` está configurado, para
      evitar errores por cobertura parcial de SES.

---

### 8. Validar en PostgreSQL

La DB local suele estar en `localhost:5434` (expuesta por docker-compose).

```bash
psql "host=localhost port=5434 dbname=app user=app password=app"
```

Dentro de `psql`:

```sql
\d
SELECT *
FROM transactions;
SELECT *
FROM account_summaries;
```

Si ves filas que coinciden con tu CSV, el flujo está funcionando.

---

### 9. Limpiar / resetear entorno

**Apagar contenedores:**

```bash
make compose-down
# o
docker compose down
```

**Reset total (contenedores + volúmenes + datos locales):**

```bash
make reset
```

En Windows esto también limpia `C:\docker-data\stori`.

---

## 🧪 Testing y TDD

El proyecto trae varias capas de pruebas:

- **Dominio (`internal/core/domain`)**
    - Validación de estructuras y lógica básica.

- **Casos de uso (`internal/core/application`)**
    - Tests de `SummaryService`: cálculo de totales, agrupación por mes, promedio de débitos/créditos, interacción con
      puertos (repositorio, lector de archivos, email).

- **Adaptadores (`internal/interfaces/out`)**
    - CSV reader S3.
    - Repositorio RDS (GORM).
    - Envío de email (SES + Noop).

- **Infra / bootstrap**
    - Tests de integración de wiring entre componentes.

### Comandos de pruebas (Makefile)

**Unitarias (dominio, servicios, adaptadores):**

```bash
make test
```

Internamente ejecuta:

```bash
go test ./internal/... -v -cover
```

**Integración (cuando estén configuradas en `./tests/integration/...`):**

```bash
make test-integration
```

Este target está preparado para leer el endpoint de la DB desde Terraform (`terraform output db_endpoint`) cuando la
infraestructura está levantada.

**Ejecutar todo:**

```bash
make test-all
```

---

## ☁️ Infraestructura con Terraform (AWS real)

En `deployments/terraform` se define la infraestructura necesaria en AWS:

### Recursos principales

- **VPC por defecto** (`data "aws_vpc" "default"`)
- **RDS PostgreSQL**:
    - `aws_db_instance.stori`
    - Acceso público habilitado (solo para fines de demo).
- **Bucket S3 de transacciones**:
    - Versionado habilitado.
- **Rol IAM para Lambda** con permisos para:
    - Logs (CloudWatch).
    - Lectura S3 (`AmazonS3ReadOnlyAccess`).
    - SES (`AmazonSESFullAccess`).
- **Lambda 1 – s3_processor**:
    - `aws_lambda_function.s3_processor`
    - `package_type = "Image"` → imagen en **ECR** (`var.ecr_s3_processor_image`).
    - Variables de entorno para DB, S3, SES y logo Stori.
    - Disparada por evento **S3 ObjectCreated .csv**.
- **Lambda 2 – api_handler** (en otro repo, pero orquestada desde aquí):
    - `aws_lambda_function.api_handler`
    - También basada en imagen ECR (`var.ecr_api_handler_image`).
    - Expuesta vía **API Gateway HTTP API**.
- **API Gateway v2**:
    - `aws_apigatewayv2_api.http_api`
    - Integración proxy con `api_handler`.
    - Stage `$default` con `auto_deploy = true`.

### Comandos Terraform vía Makefile

Todos se ejecutan desde la raíz del repo:

**Inicializar Terraform:**

```bash
make tf-init
```

**Ver plan de cambios:**

```bash
make tf-plan
```

**Aplicar infraestructura (crear / actualizar):**

```bash
make tf-apply
```

**Destruir infraestructura (limpieza):**

```bash
make tf-destroy
```

Atajos:

```bash
make infra-up    # equivale a tf-init + tf-apply
make infra-down  # equivale a tf-destroy
```

> ⚠️ Nota: el `provider "aws"` usa `profile = "personal"` en `main.tf`.  
> Si tienes varios perfiles en tu AWS CLI, asegúrate de que `personal` apunte a la cuenta correcta.

---

## 🧪 Flujo end-to-end en AWS

Combinando todo:

1. Se despliega la infraestructura con **Terraform** (`make infra-up`).
2. Se publica la imagen de la Lambda en **ECR** (`make login && make publish`).
3. Terraform apunta la Lambda a esas imágenes (`ecr_s3_processor_image` y `ecr_api_handler_image`).
4. Llegan peticiones al **API Gateway** hacia la segunda Lambda que:
    - Valida el archivo subido (CSV no vacío, estructura correcta).
    - Lo sube al bucket S3.
5. El evento `ObjectCreated` dispara la Lambda `s3_processor`, que:
    - Lee el CSV.
    - Calcula el resumen.
    - Persiste en RDS.
    - Envía el correo con el resumen usando SES.

Para probar el API Gateway sin tocar código, puedes usar la colección de **Postman** en:

```text
docs/api/postman/stori-api.postman_collection.json
```

---

## 📬 Revisión de la prueba (entorno desplegado)

Durante el periodo de evaluación de esta prueba técnica:

- La solución estará desplegada en mi cuenta personal de AWS (perfil `personal`).
- Puedes usar la colección de Postman incluida en `docs/api/postman` para disparar el API Gateway.
- El correo de resumen se envía a un correo temporal:

```text
joseph-stori@yopmail.com
```

Puedes entrar a YOPmail y revisar el resumen que genera la Lambda (HTML + texto plano).

---

## 🧑‍💻 Autor

**Joseph Mauricio Gutiérrez Valero**  
💼 Backend / Go / AWS / Arquitectura Hexagonal  
📧 josephmauricio23@hotmail.com

---

##  📝 Evidencia correo con imagen 

<img width="1630" height="934" alt="image" src="https://github.com/user-attachments/assets/38652a25-01df-4580-b503-d8b203f4c6fd" />

---

## 🏁 Resumen rápido

- ✅ Arquitectura hexagonal real (dominio aislado, ports & adapters).
- ✅ Lambda que procesa CSV desde S3, persiste en RDS y envía correo vía SES.
- ✅ Infraestructura reproducible con Terraform.
- ✅ Ejecutable localmente con Docker + LocalStack.
- ✅ Tests unitarios y de integración.
- ✅ Colección de Postman para probar el flujo vía API Gateway.

Si quieres entender el sistema a vista de pájaro:  
**"Subo un CSV → aparece en S3 → Lambda lo procesa → guarda en DB → manda un correo bonito con el resumen."**
