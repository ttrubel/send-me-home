# Send Me Home - Frontend

React TypeScript frontend using Connect-RPC and Vite.

## Setup

### Prerequisites
- Node.js 18+
- npm or yarn

### Install Dependencies

```bash
cd frontend
npm install
```

### Environment Variables

Create a `.env` file in the frontend directory (optional):

```bash
VITE_API_URL=http://localhost:8080
```

## Development

### Generate TypeScript Client from Proto

From the monorepo root:
```bash
buf generate
```

This generates TypeScript client code in `src/gen/`

### Run Development Server

```bash
npm run dev
```

Frontend runs on `http://localhost:3000`

### Build for Production

```bash
npm run build
```

Output in `dist/`

## Project Structure

```
frontend/
├── src/
│   ├── components/       # React components
│   │   ├── SessionStart.tsx
│   │   └── GameDesk.tsx
│   ├── api/              # API client setup
│   │   └── client.ts
│   ├── gen/              # Generated code (do not edit)
│   ├── App.tsx           # Main app component
│   └── main.tsx          # Entry point
├── public/               # Static assets
├── index.html
└── package.json
```

## Features

- **Session Start**: Initialize game session with progress tracking
- **Game Desk**: Main gameplay UI with document inspection
- **NPC Interaction**: Ask questions and get responses
- **Decision Making**: Approve/Deny workers
- **Score Tracking**: Real-time score and stats

## TODO

- [ ] Add audio playback for NPC voices
- [ ] Implement drag-and-drop for documents
- [ ] Add document field highlighting
- [ ] Add secondary check UI
- [ ] Add animations and polish
- [ ] Add responsive design for mobile
