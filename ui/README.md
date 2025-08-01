# Kubin UI

Lens-like interface for viewing Kubernetes cluster snapshots.

## What it does

- View cluster state as it was at snapshot time
- Navigate resources like in Lens (namespaces → pods → details)
- See pod logs and resource details
- Share snapshot links with team members

## Features

- **Historical Viewing** - Explore cluster state at any point in time
- **Lens-like Navigation** - Familiar Kubernetes dashboard experience
- **Pod-focused** - Detailed pod information and logs
- **Shareable Links** - One-click sharing for debugging

## Usage

1. Get a snapshot link from CLI: `kubin create`
2. Open link in browser
3. Navigate cluster resources like in Lens
4. Share link with team for debugging

## Development

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build
```

## Deploy

The UI is deployed alongside the Kubin server.

## Tech Stack

- React/Next.js
- TypeScript
- Tailwind CSS 