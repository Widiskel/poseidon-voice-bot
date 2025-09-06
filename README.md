# Poseidon Voice Bot (Testing Project)

## Overview
This repository contains a testing-only automation bot for interacting with [Poseidon Voice Data Campaigns](https://app.psdn.ai/).  
⚠️ **Disclaimer:** This project is **strictly for testing purposes only**.  
If you want to participate in Poseidon's official airdrop and campaigns, please use your **real voice recordings** to contribute valid data.  

---

## Event Information
### Poseidon Airdrop Campaign
- **Reward:** Tokens  
- **Registration:** [Join here](https://app.psdn.ai/login?ref=D3VN34EO)  
- **Steps to Participate:**
  - Login with your email  
  - Record your voice  
  - Submit your recording  
  - Participate daily in any campaign that matches your native language  
  - Every submission will be reviewed  
  - LFG 🚀  

📌 Poseidon has raised **$15,000,000 from a16z**  

---

## Prerequisites

### 1. System Requirements
- Linux or macOS recommended  
- Go 1.21+  

### 2. Dependencies
This project uses:  
- **htgo-tts** for temporary speech synthesis  
- **Poseidon API** for uploading recordings  

### 3. Google Cloud Credentials
This project requires Gmail API to verify OTP (email-based authentication).  
To set it up:  

1. Go to [Google Cloud Console](https://console.cloud.google.com/).  
2. Create a new **Project**.  
3. Enable the **Gmail API**.  
4. Go to **APIs & Services > Credentials**.  
5. Create **OAuth 2.0 Client ID** for a **Desktop Application**.  
6. Download the JSON credentials file.  
7. Rename the file to `credentials.json`.  
8. Create `config` folder on the root of this project and paste `credentials.json` inside `config` folder.  
9. Add your Google account as a **Test User** in OAuth Consent Screen.  

---

## Installation

```bash
# Clone the repository
git clone https://github.com/your-username/poseidon-voice-bot.git
cd poseidon-voice-bot

# Install Go dependencies
go mod tidy

# Prepare account template file
cp accounts/accounts_tmp.json accounts/account.json
```

---

## Usage

```bash
# Run the bot
go run cmd/poseidon-voice-bot/main.go
```

Logs will show account progress, JWT management, campaign checks, and file uploads.  
Generated audio (temporary) will be created and validated before uploading.  

---

## Notes
- This bot is for **testing purposes only**.  
- If you want to actually contribute to Poseidon's campaign, use your **own real voice** recordings.  
- All synthesized voices used here are **invalid** for official contribution and will be rejected.  

---

## Referral
👉 [Sign up with referral code D3VN34EO](https://app.psdn.ai/login?ref=D3VN34EO)
