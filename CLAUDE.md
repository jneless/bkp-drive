# Claude Instructions for bkp-drive

## Project Overview
bkp-drive (不靠谱网盘) is a cloud storage project based on Volcano Engine's TOS (Object Storage) service. The project aims to evolve into a reliable cloud drive solution.

## Project Context
- **Language**: Chinese project with roadmap and documentation in Chinese
- **Platform**: Currently tested only on macOS
- **Storage Backend**: Volcano Engine TOS (Object Storage)
- **Authentication**: Uses AK/SK (Access Key/Secret Key) via environment variables

## Key Features (from Roadmap)
- Admin bucket management
- File/folder browsing with metadata
- Multi-select download/delete operations
- Multi-user support with individual buckets
- Photo gallery with thumbnails
- Search functionality with natural language support
- LLM integration for content summarization
- Recycle bin functionality
- Server-side compression and download throttling

## Development Guidelines
- Test environment: macOS only
- Configuration through environment variables for AK/SK
- Planned separation into frontend/backend projects
- MySQL metadata database integration planned
- Bandwidth target: 10Gbps up/down

## Commands
- No specific build/test/lint commands identified yet - check with user when needed

## Security Notes
- Handle AK/SK credentials securely via environment variables
- Implement proper error handling for all features
- Consider quota management for multi-user scenarios

## Reference
- Volcano Engine TOS API: https://www.volcengine.com/docs/6349/74837