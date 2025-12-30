# Fiducia

> **Assistant de Production Comptable Zero-Friction**
>
> 🎯 100% des lignes 471 justifiées à M+1, sans appel téléphonique

## 🚀 Quick Start

### Prérequis

- Docker & Docker Compose
- Go 1.22+
- Node.js 20+
- FFmpeg (pour les notes vocales)

### Démarrage rapide

```bash
# Cloner le projet
cd fiduciaa

# Copier les variables d'environnement
cp backend/.env.example backend/.env

# Démarrer les services Docker
docker-compose up -d

# Backend disponible sur http://localhost:8080
# Frontend disponible sur http://localhost:3000
```

### Développement local

```bash
# Démarrer PostgreSQL et Redis
docker-compose up -d postgres redis

# Backend (dans un terminal)
cd backend
go run ./cmd/server

# Frontend (dans un autre terminal)
cd frontend
npm install
npm run dev
```

## 📁 Structure du Projet

```
fiduciaa/
├── backend/                 # API Go
│   ├── cmd/server/          # Point d'entrée
│   ├── internal/
│   │   ├── config/          # Configuration
│   │   ├── database/        # Connexion DB + migrations
│   │   ├── handlers/        # Routes HTTP
│   │   ├── middleware/      # Middleware (auth, logging, etc.)
│   │   ├── models/          # Modèles de données
│   │   └── services/        # Logique métier
│   └── pkg/
│       ├── whatsapp/        # Client Twilio WhatsApp
│       ├── voice/           # ElevenLabs + FFmpeg
│       └── ocr/             # GPT-4o-mini Vision
├── frontend/                # Next.js 14
│   └── src/app/             # App Router
├── docker-compose.yml
├── Makefile
└── README.md
```

## 🔧 Configuration

Créer `backend/.env` avec :

```env
# Server
PORT=8080
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/fiducia?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Twilio WhatsApp
TWILIO_ACCOUNT_SID=your_sid
TWILIO_AUTH_TOKEN=your_token
TWILIO_PHONE_NUMBER=+14155238886

# ElevenLabs
ELEVENLABS_API_KEY=your_key

# OpenAI (OCR)
OPENAI_API_KEY=your_key
```

## 📊 API Endpoints

### Health Check
```
GET /api/v1/health
```

### Cabinets
```
GET    /api/v1/cabinets
POST   /api/v1/cabinets
GET    /api/v1/cabinets/{id}
```

### Clients
```
GET    /api/v1/cabinets/{cabinet_id}/clients
POST   /api/v1/cabinets/{cabinet_id}/clients
GET    /api/v1/clients/{id}
```

### Pending Lines (471)
```
GET    /api/v1/cabinets/{cabinet_id}/pending-lines
POST   /api/v1/cabinets/{cabinet_id}/pending-lines
GET    /api/v1/pending-lines/{id}
PATCH  /api/v1/pending-lines/{id}
```

### Import
```
POST   /api/v1/cabinets/{cabinet_id}/import/csv
GET    /api/v1/import/{id}/status
```

### Messages
```
GET    /api/v1/pending-lines/{id}/messages
POST   /api/v1/pending-lines/{id}/messages
POST   /api/v1/webhook/whatsapp
```

### Validation
```
GET    /api/v1/cabinets/{cabinet_id}/proposals
POST   /api/v1/proposals/{id}/approve
POST   /api/v1/proposals/{id}/reject
```

### Export
```
POST   /api/v1/cabinets/{cabinet_id}/exports
GET    /api/v1/exports/{id}
```

## 🛠️ Commandes Make

```bash
make help           # Afficher l'aide
make dev            # Démarrer l'environnement de dev
make build          # Build tous les services
make test           # Lancer les tests
make docker-up      # Démarrer Docker
make docker-down    # Arrêter Docker
make migrate        # Exécuter les migrations
```

## 📖 Architecture

### Modules

| Module | Description |
|--------|-------------|
| **A - The Listener** | Ingestion données ERP (CSV, API) |
| **B - The Voice** | Clonage vocal IA (ElevenLabs) |
| **C - The Communicator** | Hub WhatsApp (Twilio) |
| **D - The Brain** | OCR + NLP (GPT-4o-mini) |
| **E - The Closer** | Validation + Export ERP |

### Stack

- **Backend**: Go 1.22
- **Database**: PostgreSQL 16
- **Queue**: Redis + BullMQ
- **Frontend**: Next.js 14
- **Voice**: ElevenLabs Turbo v2.5
- **OCR**: GPT-4o-mini Vision
- **WhatsApp**: Twilio Business API

## 📝 License

Proprietary - Fiducia SAS
