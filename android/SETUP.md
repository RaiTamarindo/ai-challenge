# Android Setup Guide

This guide will help you set up Android Studio and test the Feature Voting Android app on your MacBook Pro.

## Step 1: Install Android Studio

1. **Download Android Studio**
   - Go to https://developer.android.com/studio
   - Download Android Studio for Mac (Apple Silicon or Intel based on your MacBook)

2. **Install Android Studio**
   - Open the downloaded `.dmg` file
   - Drag Android Studio to your Applications folder
   - Launch Android Studio from Applications

3. **Complete Setup Wizard**
   - Choose "Standard" installation
   - Accept license agreements
   - Android Studio will download Android SDK, platform tools, and other components
   - This may take 10-15 minutes

## Step 2: Create Android Virtual Device (Emulator)

1. **Open AVD Manager**
   - In Android Studio welcome screen, click "More Actions" → "AVD Manager"
   - Or use menu: Tools → AVD Manager

2. **Create Virtual Device**
   - Click "Create Virtual Device"
   - Select "Phone" category
   - Choose "Pixel 7" (recommended for good performance)
   - Click "Next"

3. **Select System Image**
   - Choose "UpsideDownCake" (API Level 34, Android 14)
   - Click "Download" if not already downloaded
   - Click "Next" after download completes

4. **Configure AVD**
   - Name: "Pixel_7_API_34"
   - Leave other settings as default
   - Click "Finish"

## Step 3: Import and Run the Project

1. **Import Project**
   - In Android Studio welcome screen, click "Open"
   - Navigate to `/Users/raitamarindo/Repos/MetaCTO/AIChallenge/android`
   - Select the `android` folder and click "Open"

2. **Wait for Project Sync**
   - Android Studio will sync the project and download dependencies
   - This may take 5-10 minutes on first import
   - You'll see "Sync successful" in the status bar when done

3. **Start the Emulator**
   - Click the AVD Manager icon in the toolbar
   - Click the "Play" button next to your "Pixel_7_API_34" device
   - Wait for the emulator to boot up (2-3 minutes first time)

4. **Run the App**
   - Make sure your backend is running on `localhost:8080`
   - In Android Studio, click the green "Run" button (or press Ctrl+R)
   - Select your emulator device
   - The app will install and launch automatically

## Step 4: Test the App

1. **Start Backend Server**
   ```bash
   cd /Users/raitamarindo/Repos/MetaCTO/AIChallenge/backend
   go run main.go
   ```
   - Make sure it's running on port 8080

2. **Test Login**
   - The app will open to the login screen
   - Pre-filled credentials are already loaded:
     - Email: admin@example.com
     - Password: admin123
   - Tap "LOGIN"

3. **Test Feature Listing**
   - After successful login, you'll see the features list
   - Pull down to refresh the list
   - Tap vote buttons to vote for features

4. **Test Feature Creation**
   - Tap the "+" button in the top right
   - Fill in title and description
   - Tap "CREATE" to add a new feature

## Troubleshooting

### Emulator Issues
- **Slow performance**: Enable "Hardware Acceleration" in AVD settings
- **Network issues**: Make sure backend is running on `localhost:8080`
- **App crashes**: Check Android Studio logs in the "Logcat" tab

### API Connection Issues
- **Connection refused**: Ensure backend is running
- **Wrong endpoint**: App uses `10.0.2.2:8080` (emulator's localhost)
- **CORS issues**: Backend should allow all origins in development

### Build Issues
- **Sync failed**: Try "File" → "Sync Project with Gradle Files"
- **Dependencies**: Check internet connection for downloads

## Quick Start Commands

Once everything is set up:

1. **Start Backend**:
   ```bash
   cd backend && go run main.go
   ```

2. **Run Android App**:
   - Open Android Studio
   - Click green "Run" button
   - App launches in emulator

## App Features

✅ **Login Screen**: Authenticate with backend  
✅ **Feature List**: View all features with vote counts  
✅ **Voting**: Vote for features (one vote per user per feature)  
✅ **Create Feature**: Add new feature requests  
✅ **Pull to Refresh**: Update feature list  
✅ **Logout**: Clear session and return to login  

The app is now ready for testing! The entire flow from login to feature creation should work seamlessly with your backend API.