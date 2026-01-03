# Firestore Server-Side Configuration Guide

This guide covers server-side Firestore configuration including security rules, indexes, TTL policies, and database settings.

## Table of Contents
1. [Security Rules](#security-rules)
2. [Database Indexes](#database-indexes)
3. [TTL (Time-to-Live) Policy](#ttl-policy)
4. [Backup Configuration](#backup-configuration)
5. [Deployment](#deployment)

---

## Security Rules

Security rules control who can access your Firestore data. For this game, **only the backend service account should access Firestore** - never clients directly.

### Deploy Security Rules

#### Option 1: Using Firebase Console (Easiest)

1. Go to [Firestore Console](https://console.cloud.google.com/firestore)
2. Click on **"Rules"** tab
3. Replace the existing rules with the content from `firestore.rules`:

```javascript
rules_version = '2';

service cloud.firestore {
  match /databases/{database}/documents {
    // Sessions collection - only accessible by backend service account
    match /sessions/{sessionId} {
      allow read, write: if false;  // Deny all client access
    }

    // Default deny all
    match /{document=**} {
      allow read, write: if false;
    }
  }
}
```

4. Click **"Publish"**

#### Option 2: Using Firebase CLI

```bash
# Install Firebase CLI if you haven't
npm install -g firebase-tools

# Login to Firebase
firebase login

# Initialize Firestore in your project
firebase init firestore

# This creates firestore.rules and firestore.indexes.json
# We've already created these files for you

# Deploy the rules
firebase deploy --only firestore:rules

# Deploy indexes
firebase deploy --only firestore:indexes
```

#### Option 3: Using gcloud CLI

```bash
# Deploy security rules
gcloud firestore databases update \
  --project=your-project-id \
  --type=firestore-native \
  --location=us-central1

# Note: gcloud doesn't directly support rules deployment
# Use Firebase CLI for rules deployment
```

### Understanding the Rules

```javascript
allow read, write: if false;
```

This means:
- ✅ **Backend service account CAN access** (service accounts bypass rules)
- ❌ **Web/mobile clients CANNOT access** (all client requests denied)
- ✅ **Security**: Prevents unauthorized access to game sessions
- ✅ **Architecture**: Enforces all data access through your Connect-RPC API

---

## Database Indexes

Indexes improve query performance. Firestore automatically creates single-field indexes, but composite indexes must be created manually.

### Current Indexes

The `firestore.indexes.json` file defines:

```json
{
  "indexes": [
    {
      "collectionGroup": "sessions",
      "queryScope": "COLLECTION",
      "fields": [
        {"fieldPath": "session_id", "order": "ASCENDING"},
        {"fieldPath": "current_case_index", "order": "ASCENDING"}
      ]
    }
  ]
}
```

### Deploy Indexes

```bash
# Using Firebase CLI
firebase deploy --only firestore:indexes

# View index build status
firebase firestore:indexes
```

### When to Add More Indexes

Add indexes if you plan to query sessions by:
- Creation time: Add index on `created_at` field
- User ID: Add index on `user_id` field (if you add user authentication)
- Score: Add index on `score` field (for leaderboards)

Example index for queries like "get top 10 scores":
```json
{
  "collectionGroup": "sessions",
  "fields": [
    {"fieldPath": "score", "order": "DESCENDING"},
    {"fieldPath": "created_at", "order": "DESCENDING"}
  ]
}
```

---

## TTL Policy (Automatic Session Cleanup)

Time-to-Live policies automatically delete old sessions to save storage costs.

### Option 1: Firestore TTL (Recommended)

Firestore can automatically delete documents based on a timestamp field.

1. **Add timestamp field to sessions**:

```go
// In backend/internal/models/case.go
type Session struct {
    SessionID              string    `json:"session_id"`
    CreatedAt              time.Time `json:"created_at" firestore:"created_at,serverTimestamp"`
    ExpiresAt              time.Time `json:"expires_at" firestore:"expires_at"`
    // ... other fields
}
```

2. **Set expiration when creating sessions**:

```go
// In backend/internal/api/handler.go - StartSession
session := &models.Session{
    SessionID:  sessionID,
    CreatedAt:  time.Now(),
    ExpiresAt:  time.Now().Add(24 * time.Hour), // Delete after 24 hours
    // ... other fields
}
```

3. **Enable TTL in Firestore Console**:

```bash
# Using gcloud CLI
gcloud firestore fields ttls update expires_at \
  --collection-group=sessions \
  --enable-ttl \
  --project=your-project-id
```

Or in the Console:
- Go to Firestore > Select database
- Click on "Time-to-live" tab
- Click "Create TTL policy"
- Collection: `sessions`
- Timestamp field: `expires_at`

### Option 2: Cloud Scheduler + Cloud Functions

For more control, use a scheduled cleanup function:

```bash
# Create Cloud Function to delete old sessions
gcloud functions deploy cleanupOldSessions \
  --runtime go121 \
  --trigger-topic cleanup-sessions \
  --entry-point CleanupOldSessions \
  --project=your-project-id

# Create Cloud Scheduler job (runs daily at 2 AM)
gcloud scheduler jobs create pubsub cleanup-sessions-daily \
  --schedule="0 2 * * *" \
  --topic=cleanup-sessions \
  --message-body='{"action":"cleanup"}' \
  --project=your-project-id
```

---

## Backup Configuration

### Enable Point-in-Time Recovery (PITR)

Protects against accidental deletions:

```bash
# Enable PITR (retains backups for 7 days)
gcloud firestore databases update \
  --project=your-project-id \
  --location=us-central1 \
  --enable-point-in-time-recovery
```

Cost: ~$0.18 per GB/month

### Scheduled Backups

```bash
# Create weekly backup schedule
gcloud firestore backups schedules create \
  --database='(default)' \
  --recurrence=weekly \
  --retention=4w \
  --project=your-project-id
```

### Manual Backup

```bash
# One-time backup
gcloud firestore export gs://your-backup-bucket/firestore-backup-$(date +%Y%m%d) \
  --project=your-project-id
```

---

## Database Settings & Limits

### Configure Database Limits

In Firestore Console:
1. Go to **Settings** > **Limits**
2. Adjust:
   - **Maximum concurrent connections**: Default 1M (more than enough)
   - **Maximum document size**: 1 MB (default, sufficient for sessions)
   - **Maximum writes per second**: 10,000 (default)

### Monitor Usage

1. Go to [Firestore Usage Dashboard](https://console.cloud.google.com/firestore/usage)
2. Monitor:
   - Document reads/writes per day
   - Storage usage
   - Index entries

### Set Budget Alerts

```bash
# Create budget alert for Firestore costs
gcloud billing budgets create \
  --billing-account=YOUR_BILLING_ACCOUNT_ID \
  --display-name="Firestore Budget" \
  --budget-amount=10USD \
  --threshold-rule=percent=50 \
  --threshold-rule=percent=90 \
  --threshold-rule=percent=100
```

---

## Complete Deployment Checklist

### Initial Setup (One-time)

```bash
# 1. Install Firebase CLI
npm install -g firebase-tools

# 2. Login and initialize
firebase login
firebase init firestore
# Select your GCP project when prompted

# 3. Deploy rules and indexes
firebase deploy --only firestore:rules,firestore:indexes

# 4. Enable TTL policy
gcloud firestore fields ttls update expires_at \
  --collection-group=sessions \
  --enable-ttl \
  --project=your-project-id

# 5. Enable PITR (optional, for production)
gcloud firestore databases update \
  --project=your-project-id \
  --enable-point-in-time-recovery
```

### After Code Changes

```bash
# If you modified firestore.rules
firebase deploy --only firestore:rules

# If you added new queries requiring indexes
firebase deploy --only firestore:indexes
```

---

## Environment-Specific Configuration

### Development Environment

```bash
# .env (local development)
FIRESTORE_PROJECT_ID=send-me-home-dev
GOOGLE_APPLICATION_CREDENTIALS=./gcp-keys/dev-key.json
```

**Settings:**
- Relaxed security rules (for testing)
- No TTL (keep data for debugging)
- No backups

### Production Environment

```bash
# Cloud Run environment variables
FIRESTORE_PROJECT_ID=send-me-home-prod
# No credentials needed - uses Cloud Run service identity
```

**Settings:**
- Strict security rules (`allow read, write: if false`)
- TTL: 24 hours
- PITR enabled
- Weekly backups
- Budget alerts

---

## Advanced Configuration

### Multi-Region Setup (High Availability)

```bash
# Create multi-region database (higher cost)
gcloud firestore databases create \
  --type=firestore-native \
  --location=nam5 \
  --project=your-project-id
```

Multi-region locations:
- `nam5`: North America (US, Canada)
- `eur3`: Europe
- Single region is fine for most games

### Read Replicas (Low Latency)

For global games, consider using Firestore's multi-region deployment:

```bash
# This is set at database creation time
# Requires enterprise plan
```

---

## Monitoring & Alerts

### Set Up Alerting Policies

1. Go to [Cloud Monitoring](https://console.cloud.google.com/monitoring)
2. Create alerting policies for:
   - **High read/write rates** (approaching quota limits)
   - **Increased latency** (slow queries)
   - **Error rates** (permission denied errors)

Example alert using gcloud:

```bash
gcloud alpha monitoring policies create \
  --notification-channels=CHANNEL_ID \
  --display-name="Firestore High Write Rate" \
  --condition-display-name="Write rate > 5000/min" \
  --condition-threshold-value=5000 \
  --condition-threshold-duration=60s
```

---

## Testing Security Rules

### Local Emulator Testing

```bash
# Start Firestore emulator
firebase emulators:start --only firestore

# Run tests against emulator
FIRESTORE_EMULATOR_HOST=localhost:8080 go test ./...
```

### Rules Unit Tests

Create `firestore.test.js`:

```javascript
const { initializeTestEnvironment } = require('@firebase/rules-unit-testing');

describe('Firestore Security Rules', () => {
  it('denies client read access to sessions', async () => {
    const env = await initializeTestEnvironment({
      projectId: 'test-project',
      firestore: {
        rules: fs.readFileSync('firestore.rules', 'utf8'),
      },
    });

    const alice = env.authenticatedContext('alice');
    await assertFails(alice.firestore().collection('sessions').doc('test').get());
  });
});
```

---

## Quick Reference Commands

```bash
# Deploy everything
firebase deploy --only firestore

# Deploy only rules
firebase deploy --only firestore:rules

# Deploy only indexes
firebase deploy --only firestore:indexes

# View current rules
gcloud firestore databases describe --project=your-project-id

# Monitor in real-time
gcloud logging tail "resource.type=cloud_firestore_database" --project=your-project-id

# Export data
gcloud firestore export gs://your-bucket/backup --project=your-project-id

# Import data
gcloud firestore import gs://your-bucket/backup --project=your-project-id
```

---

## Cost Optimization Tips

1. **Use TTL policies** - Auto-delete old sessions (saves storage)
2. **Minimize document reads** - Cache session data in backend memory when possible
3. **Batch writes** - Use transactions instead of individual writes
4. **Clean up completed sessions** - Delete sessions after game ends
5. **Monitor usage** - Set up billing alerts

**Expected costs for 10,000 daily active users:**
- Reads: ~500K/day = Free tier
- Writes: ~200K/day = ~$0.60/day
- Storage: ~1GB = Free tier
- **Total**: ~$18/month

---

## Troubleshooting

### "Permission Denied" in Production
- Check security rules are deployed: `firebase deploy --only firestore:rules`
- Verify service account has `roles/datastore.user`

### "Index Required" Error
- Deploy indexes: `firebase deploy --only firestore:indexes`
- Wait 5-10 minutes for index to build

### High Latency
- Check database region matches your Cloud Run region
- Add indexes for frequently used queries
- Consider upgrading to multi-region

---

## Next Steps

✅ **Security rules deployed**
✅ **Indexes configured**
✅ **TTL policy enabled**
⬜ Set up monitoring alerts
⬜ Configure backups (production)
⬜ Load test with realistic traffic

For questions, see [Firestore Documentation](https://cloud.google.com/firestore/docs) or the main [FIRESTORE_SETUP.md](./FIRESTORE_SETUP.md) guide.
