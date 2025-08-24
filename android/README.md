# Feature Voting Platform - Android App

Simple Android app for the Feature Voting Platform.

## Setup Instructions

### Prerequisites
- Android Studio installed on MacBook
- Android emulator configured

### Quick Start

1. **Open in Android Studio:**
   ```bash
   # Open Android Studio
   # File -> Open -> Select this android folder
   ```

2. **Run the app:**
   - Click the green "Run" button
   - Select your emulator device
   - App will install and launch

### Architecture

Simple architecture with:
- **Activities**: Main screens (Login, Features, etc.)
- **API Service**: HTTP client for backend communication
- **Models**: Data classes for API responses
- **Utils**: Helper classes

### API Configuration

The app connects to your local backend:
- **Base URL**: `http://10.0.2.2:8080` (emulator localhost)
- **Authentication**: JWT tokens
- **Storage**: SharedPreferences

### Features

1. **Login**: Authenticate with backend
2. **Feature List**: Browse all features with vote counts
3. **Voting**: Vote/unvote on features
4. **Create Feature**: Add new feature proposals
5. **My Features**: View user's created features

### Testing

Use the Android emulator:
1. Start your backend: `make dev`
2. Run Android app in emulator
3. Login with test credentials
4. Test all functionality

## Project Structure

```
android/
├── app/
│   ├── src/main/java/com/featurevoting/
│   │   ├── activities/          # Screen activities
│   │   ├── api/                 # HTTP client
│   │   ├── models/              # Data models
│   │   └── utils/               # Utilities
│   ├── src/main/res/
│   │   ├── layout/              # XML layouts
│   │   └── values/              # Strings, colors
│   └── build.gradle             # Dependencies
└── build.gradle                 # Project config
```