# 💳 Stori Challenge – Joseph Mauricio Gutiérrez Valero

### 🧠 Descripción

Este proyecto fue desarrollado como solución al **Stori Software Engineer Technical Challenge**.  
Procesa un archivo CSV con transacciones de crédito y débito, calcula el resumen mensual y envía un **correo electrónico
con la información consolidada**.

Diseñado con **arquitectura hexagonal**, **principios de TDD** y preparado para ejecutarse tanto en **Docker** como en *
*AWS Lambda**.

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

## 🚀 Ejecución Local

### 🔧 Requisitos previos

- Go ≥ 1.22
- Docker y Docker Compose
- Archivo `.env` con variables de entorno

### ▶️ Con Makefile

```bash
make build      # Compila binarios
make test       # Ejecuta todos los tests
make run        # Ejecuta el servicio localmente
```

### 🐳 Con Docker

```bash
docker-compose up --build
```

Esto levantará:

- Contenedor de aplicación `stori-app`
- Contenedor de base de datos `postgres:latest`

*(Windows PowerShell)*

```powershell
docker-compose up --build
```

---

## 🧪 Testing (TDD aplicado)

- **Tests unitarios** en `internal/core/...`
- **Mocks** para adaptadores y servicios externos.
- **Cobertura** de entidades, casos de uso y repositorios.

```bash
go test -v ./...
```

---

## 🧑‍💻 Autor

**Joseph Mauricio Gutiérrez Valero**  
📧 joseph.gutierrez@example.com  
🔗 [GitHub](https://github.com/JosephMGutierrezV) · [LinkedIn](https://www.linkedin.com/in/joseph-gutierrez-v/)

---

## 🏁 Conclusión

✅ Desarrollado con **Go + GORM**  
✅ Arquitectura **Hexagonal / Clean Architecture**  
✅ Compatible con **Docker** y **AWS Lambda**  
✅ Probado con **TDD** y herramientas modernas de Go
