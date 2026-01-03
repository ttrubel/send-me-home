# Firestore Setup Guide

This guide walks you through setting up Cloud Firestore for the Send Me Home game backend.

## Prerequisites

- A Google Cloud Platform (GCP) account
- A GCP project (create one at [console.cloud.google.com](https://console.cloud.google.com))
- Billing enabled on your project (Firestore has a free tier)

## Step 1: Enable Firestore API

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Select your project from the dropdown at the top
3. Navigate to **Firestore** in the left sidebar, or search for "Firestore" in the search bar
4. Click **"Select Native Mode"** when prompted
   - **Native Mode** is recommended (supports real-time updates, offline persistence)
   - **Datastore Mode** is for legacy compatibility (don't choose this)

## Step 2: Choose a Location

1. Select a region/multi-region for your Firestore database
   - **Recommended**: `us-central1` (Iowa) - low latency for US users
   - **Multi-region options**: `nam5` (US), `eur3` (Europe)
   - **Note**: Location cannot be changed after creation

2. Click **"Create Database"**
3. Wait for the database to be provisioned (takes ~1 minute)

## Step 3: Create a Service Account

You need a service account to authenticate your backend server.

### Option A: Using the Console (Recommended for Development)

1. Go to **IAM & Admin** > **Service Accounts**
   - Direct link: https://console.cloud.google.com/iam-admin/serviceaccounts

2. Click **"+ CREATE SERVICE ACCOUNT"**

3. Fill in the details:
   - **Service account name**: `send-me-home-backend`
   - **Description**: "Backend service for Send Me Home game"
   - Click **"CREATE AND CONTINUE"**

4. Grant roles:
   - Add role: **Cloud Datastore User** (for Firestore read/write)
   - Add role: **Vertex AI User** (if using Gemini/Vertex AI)
   - Click **"CONTINUE"**

5. Click **"DONE"**

6. Download the service account key:
   - Click on the service account you just created
   - Go to **"KEYS"** tab
   - Click **"ADD KEY"** > **"Create new key"**
   - Choose **JSON** format
   - Click **"CREATE"**
   - Save the JSON file securely (e.g., `~/gcp-keys/send-me-home-key.json`)

### Option B: Using gcloud CLI

```bash
# Set your project ID
export PROJECT_ID="your-project-id"

# Create service account
gcloud iam service-accounts create send-me-home-backend \
    --display-name="Send Me Home Backend" \
    --project=$PROJECT_ID

# Grant Firestore permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:send-me-home-backend@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/datastore.user"

# Grant Vertex AI permissions (if using Gemini)
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:send-me-home-backend@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/aiplatform.user"

# Create and download key
gcloud iam service-accounts keys create ~/gcp-keys/send-me-home-key.json \
    --iam-account=send-me-home-backend@${PROJECT_ID}.iam.gserviceaccount.com
```

## Step 4: Configure Your Backend

1. **Create the credentials directory** (if needed):
   ```bash
   mkdir -p ~/gcp-keys
   chmod 700 ~/gcp-keys  # Secure the directory
   ```

2. **Update your `.env` file** in the `backend/` directory:
   ```bash
   cd backend
   cp .env.example .env
   ```

3. **Edit `backend/.env`** with your project details:
   ```bash
   # Google Cloud Configuration
   GOOGLE_CLOUD_PROJECT=your-actual-project-id
   GOOGLE_CLOUD_LOCATION=us-central1
   GOOGLE_APPLICATION_CREDENTIALS=/Users/yourusername/gcp-keys/send-me-home-key.json

   # Firestore Configuration
   FIRESTORE_PROJECT_ID=your-actual-project-id

   # Enable Vertex AI (if using Gemini)
   GOOGLE_GENAI_USE_VERTEXAI=true
   ```

4. **Verify the credentials file path**:
   ```bash
   ls -l $GOOGLE_APPLICATION_CREDENTIALS
   # Should show your JSON key file
   ```

## Step 5: Test the Connection

Run the backend server to verify Firestore connection:

```bash
cd backend
go run cmd/server/main.go
```

You should see:
```
Server starting on port 8080
Game service available at http://localhost:8080/game.v1.GameService
```

If you see errors like `Failed to initialize Firestore`, check:
- Your `FIRESTORE_PROJECT_ID` matches your GCP project ID
- The credentials file path is correct and accessible
- The service account has the `Cloud Datastore User` role

## Step 6: View Data in Firestore Console

1. Go to [Firestore Console](https://console.cloud.google.com/firestore/databases)
2. Select your database
3. After starting a game session, you should see a `sessions` collection appear
4. Click on the collection to view stored game sessions

## Firestore Data Structure

Your game data is stored as follows:

```
sessions (collection)
  └── {sessionID} (document)
      ├── session_id: string
      ├── game_date: string
      ├── rules: array
      ├── cases: array of Case objects
      ├── current_case_index: number
      ├── score: number
      ├── correct_decisions: number
      ├── incorrect_decisions: number
      ├── secondary_checks_quota: number
      ├── remaining_secondary_checks: number
      └── completed_cases: array
```

## Cost Estimation

Firestore pricing (as of 2024):

**Free Tier (per day):**
- 50,000 document reads
- 20,000 document writes
- 20,000 document deletes
- 1 GB storage

**Your game usage:**
- Each game session: ~15-20 writes, ~30-50 reads
- Free tier covers ~1,000+ game sessions per day
- Storage: Minimal (sessions can be auto-deleted after 24 hours)

**To stay within free tier:**
- Monitor usage in Cloud Console
- Implement session cleanup (TTL) for old sessions
- Consider deleting sessions after game completion

## Production Deployment (Cloud Run)

When deploying to Cloud Run, you don't need to manage credentials manually:

1. **Deploy with default service account**:
   ```bash
   gcloud run deploy send-me-home \
     --source . \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated
   ```

2. **Grant permissions to Cloud Run service account**:
   ```bash
   PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")

   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
     --role="roles/datastore.user"
   ```

3. **Set environment variables in Cloud Run**:
   - Go to Cloud Run service settings
   - Add: `FIRESTORE_PROJECT_ID=your-project-id`
   - Cloud Run automatically uses the service identity for authentication

## Security Best Practices

1. **Never commit credentials**:
   ```bash
   # Add to .gitignore
   echo "*.json" >> .gitignore
   echo "backend/.env" >> .gitignore
   ```

2. **Use Secret Manager** (production):
   - Store API keys in Secret Manager instead of .env files
   - Grant Cloud Run service account access to secrets

3. **Set up Firestore Security Rules**:
   ```javascript
   // Firestore Rules (in Console)
   rules_version = '2';
   service cloud.firestore {
     match /databases/{database}/documents {
       // Sessions are only writable by backend service
       match /sessions/{sessionId} {
         allow read, write: if false;  // Backend uses service account
       }
     }
   }
   ```

4. **Limit service account permissions**:
   - Use principle of least privilege
   - Only grant `datastore.user`, not `owner` or `editor`

## Troubleshooting

### Error: "PERMISSION_DENIED"
- Check service account has `Cloud Datastore User` role
- Verify `GOOGLE_APPLICATION_CREDENTIALS` points to valid key file

### Error: "Project not found"
- Verify `FIRESTORE_PROJECT_ID` matches your GCP project ID exactly
- Check project ID in Cloud Console (not display name)

### Error: "Database not found"
- Ensure you created a Firestore database in Native Mode
- Database creation can take a few minutes

### Error: "Application Default Credentials not found"
- Set `GOOGLE_APPLICATION_CREDENTIALS` environment variable
- Point it to your service account JSON key file

## Next Steps

- ✅ Firestore is now configured
- [ ] Set up Vertex AI for Gemini integration
- [ ] Configure ElevenLabs for voice synthesis
- [ ] Deploy to Cloud Run for production

## Resources

- [Firestore Documentation](https://cloud.google.com/firestore/docs)
- [Firestore Pricing](https://cloud.google.com/firestore/pricing)
- [Service Account Best Practices](https://cloud.google.com/iam/docs/best-practices-service-accounts)
- [Cloud Run Authentication](https://cloud.google.com/run/docs/securing/service-identity)
