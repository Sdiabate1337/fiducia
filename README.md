# 🏛️ Fiducia

> **SaaS platform for accounting firms to automate client document collection via WhatsApp AI**

Fiducia transforms how accounting firms (cabinets comptables) handle the tedious task of collecting supporting documents for pending accounting entries (compte 471). Instead of endless emails and phone calls, Fiducia automates the entire process through WhatsApp with AI-powered voice messages and intelligent document recognition.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-14+-black?style=flat&logo=next.js)](https://nextjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Twilio](https://img.shields.io/badge/Twilio-WhatsApp-F22F46?style=flat&logo=twilio)](https://www.twilio.com/)
[![ElevenLabs](https://img.shields.io/badge/ElevenLabs-Voice%20AI-000000?style=flat)](https://elevenlabs.io/)
[![OpenAI](https://img.shields.io/badge/OpenAI-GPT--4o%20Vision-412991?style=flat&logo=openai)](https://openai.com/)

---

## ✨ Features

### 📤 Smart Document Request
- **CSV Import** - Bulk import pending lines from your accounting software
- **Client Matching** - Automatic client detection and assignment
- **WhatsApp Integration** - Send requests directly via WhatsApp Business API

### 🎙️ AI Voice Messages
- **ElevenLabs TTS** - Natural French voice synthesis
- **Voice Cloning** - Clone your staff's voice for personalized messages
- **OGG/Opus Conversion** - FFmpeg conversion for WhatsApp compatibility

### 📄 OCR Document Processing
- **GPT-4o Vision** - Intelligent document text extraction
- **Structured Data** - Automatic extraction of date, amount, vendor, invoice number
- **Auto-Matching** - AI matches documents to pending entries with confidence scoring

### 🔄 Workflow Automation
- **Anti-Ban Queue** - Smart message scheduling with jitter (30-180s delays)
- **Status Tracking** - Real-time status: pending → contacted → received → validated
- **Webhook Integration** - Receive client responses and documents automatically

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         FIDUCIA PLATFORM                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Next.js   │    │   Go API    │    │ PostgreSQL  │         │
│  │  Frontend   │◄──►│   Backend   │◄──►│  Database   │         │
│  └─────────────┘    └──────┬──────┘    └─────────────┘         │
│                            │                                    │
│         ┌──────────────────┼──────────────────┐                │
│         │                  │                  │                │
│         ▼                  ▼                  ▼                │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Twilio    │    │ ElevenLabs  │    │   OpenAI    │         │
│  │  WhatsApp   │    │  Voice AI   │    │ GPT-4o OCR  │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **Node.js 18+**
- **PostgreSQL 17**
- **FFmpeg** (for voice conversion)
- **ngrok** (for local webhook testing)

### 1. Clone & Setup

```bash
git clone https://github.com/your-org/fiducia.git
cd fiducia
```

### 2. Database Setup

```bash
# Create database
createdb fiducia

# Run migrations
psql fiducia < backend/migrations/001_schema.sql
psql fiducia < backend/migrations/002_messages.sql
psql fiducia < backend/migrations/003_seed.sql
psql fiducia < backend/migrations/004_documents.sql
```

### 3. Backend Configuration

```bash
cd backend
cp .env.example .env
```

Edit `.env` with your credentials:

```env
# Database
DATABASE_URL=postgres://localhost:5432/fiducia?sslmode=disable

# Twilio WhatsApp
TWILIO_ACCOUNT_SID=your_account_sid
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_PHONE_NUMBER=+14155238886

# ElevenLabs Voice AI
ELEVENLABS_API_KEY=your_api_key
ELEVENLABS_VOICE_ID=your_voice_id

# OpenAI (GPT-4o Vision)
OPENAI_API_KEY=your_api_key

# Environment
ENVIRONMENT=development
BASE_URL=https://your-ngrok-url.ngrok-free.dev
```

### 4. Start Backend

```bash
cd backend
go run ./cmd/server
```

### 5. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

### 6. Setup Webhook (Development)

```bash
# In a new terminal
ngrok http 8080

# Configure in Twilio Console:
# Webhook URL: https://your-url.ngrok-free.dev/webhook/whatsapp
```

---

## 📱 Usage

### Import Pending Lines

1. Navigate to **Dashboard** → **📤 Importer CSV**
2. Upload your CSV file with columns: `date`, `libellé`, `montant`, `client` (optional)
3. Preview and confirm import

### Send Document Requests

1. Click **Voir** on any pending line
2. Choose message type: **📝 Texte** or **🎙️ Vocal**
3. Click **⚡ Envoyer immédiat** or schedule with anti-ban delay

### Process Received Documents

1. Client sends photo of document via WhatsApp
2. System automatically:
   - Downloads the image
   - Extracts text with GPT-4o Vision
   - Matches to pending line
   - Creates validation proposal
3. Review and **✓ Valider** or reject

---

## 🗂️ Project Structure

```
fiducia/
├── backend/
│   ├── cmd/server/          # Application entrypoint
│   ├── internal/
│   │   ├── config/          # Configuration
│   │   ├── database/        # Database connection
│   │   ├── handlers/        # HTTP handlers
│   │   ├── models/          # Domain models
│   │   ├── repository/      # Data access
│   │   └── services/        # Business logic
│   ├── pkg/
│   │   ├── voice/           # ElevenLabs integration
│   │   └── whatsapp/        # Twilio integration
│   └── migrations/          # SQL migrations
│
├── frontend/
│   ├── src/app/             # Next.js App Router
│   └── public/              # Static assets
│
└── README.md
```

---

## 🔌 API Endpoints

### Pending Lines
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/cabinets/{id}/pending-lines` | List pending lines |
| POST | `/api/v1/cabinets/{id}/import/csv` | Import CSV |
| GET | `/api/v1/pending-lines/{id}` | Get line details |

### Messages
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/pending-lines/{id}/messages` | List messages |
| POST | `/api/v1/pending-lines/{id}/messages` | Send relance |

### Documents
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/pending-lines/{id}/documents` | List documents |
| POST | `/api/v1/documents/{id}/approve` | Approve document |
| POST | `/api/v1/documents/{id}/reject` | Reject document |

### Webhooks
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/webhook/whatsapp` | Twilio incoming webhook |

---

## 🛡️ Security

- **Multi-tenant Architecture** - Data isolation per cabinet
- **Twilio Signature Verification** - Webhook authentication
- **Environment-based Config** - Secrets via environment variables
- **CORS Protection** - Configured for frontend origin

---

## 📈 Roadmap

- [x] **Sprint 0** - Foundation (Database, API, Auth)
- [x] **Sprint 1** - CSV Import with client matching
- [x] **Sprint 2** - WhatsApp Hub (Twilio, message queue)
- [x] **Sprint 3** - Voice AI (ElevenLabs, FFmpeg)
- [x] **Sprint 4** - OCR Documents (GPT-4o Vision, auto-matching)
- [ ] **Sprint 5** - Export to accounting software (FEC format)
- [ ] **Sprint 6** - Analytics dashboard
- [ ] **Sprint 7** - Multi-user roles & permissions

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is proprietary software. All rights reserved.

---

## 🙏 Acknowledgments

- [Twilio](https://www.twilio.com/) - WhatsApp Business API
- [ElevenLabs](https://elevenlabs.io/) - Voice AI
- [OpenAI](https://openai.com/) - GPT-4o Vision
- [FFmpeg](https://ffmpeg.org/) - Audio conversion

---

<p align="center">
  <strong>Built with ❤️ for accounting firms</strong><br>
  <em>Fiducia - La confiance numérique</em>
</p>
