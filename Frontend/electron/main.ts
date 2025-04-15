import { app, BrowserWindow, ipcMain, globalShortcut, screen } from 'electron';
import * as path from 'path';

// Handle any uncaught exceptions
process.on('uncaughtException', (error: Error) => {
  console.error('An uncaught error occurred:', error);
});

let mainWindow: BrowserWindow | null = null;
let commandWindow: BrowserWindow | null = null;

function createCommandWindow(): void {
  const { width: screenWidth, height: screenHeight } = screen.getPrimaryDisplay().workAreaSize;
  
  commandWindow = new BrowserWindow({
    width: 600,
    height: 160,
    x: (screenWidth - 600) / 2,
    y: screenHeight / 4,
    frame: false,
    resizable: false,
    titleBarStyle: 'hidden',
    alwaysOnTop: true,
    backgroundColor: '#ffffff',
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: true,
      preload: path.join(__dirname, '../dist-electron/preload.js')
    }
  });

  const isDev = process.env.NODE_ENV === 'development';
  
  if (isDev) {
    commandWindow.loadURL('https://localhost:3000/#command');
  } else {
    commandWindow.loadFile(path.join(__dirname, '../dist/index.html'), {
      hash: 'command'
    });
  }

  // Hide window when it loses focus
  commandWindow.on('blur', () => {
    commandWindow?.hide();
  });

  // Handle window closing
  commandWindow.on('closed', () => {
    commandWindow = null;
  });
}

function toggleCommandWindow() {
  if (!commandWindow) {
    createCommandWindow();
    return;
  }

  if (commandWindow.isVisible()) {
    commandWindow.hide();
  } else {
    // Recenter the window before showing
    const { width: screenWidth, height: screenHeight } = screen.getPrimaryDisplay().workAreaSize;
    commandWindow.setPosition(
      Math.floor((screenWidth - 600) / 2),
      Math.floor(screenHeight / 4)
    );
    commandWindow.show();
    commandWindow.focus();
  }
}

function createWindow(): void {
  // Create the browser window.
  mainWindow = new BrowserWindow({
    width: 1600,
    height: 900,
    minWidth: 1600,
    minHeight: 900,
    frame: false,
    titleBarStyle: 'hidden',
    backgroundColor: '#ffffff',
    icon: path.join(__dirname, '../AppLogo.png'),
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: true,
      preload: path.join(__dirname, '../dist-electron/preload.js')
    }
  });

  // In development, load the dev server URL
  const isDev = process.env.NODE_ENV === 'development';
  
  if (isDev) {
    // Disable certificate validation in development
    mainWindow.webContents.session.setCertificateVerifyProc((request, callback) => {
      // Always trust the certificate in development mode
      callback(0);
    });
    
    mainWindow.loadURL('https://localhost:3000');
    mainWindow.webContents.openDevTools();
    
    mainWindow.webContents.on('did-fail-load', () => {
      console.log('Failed to load URL, retrying...');
      setTimeout(() => {
        if (mainWindow) {
          mainWindow.loadURL('https://localhost:3000');
        }
      }, 1000);
    });
  } else {
    // In production, load from the dist directory
    mainWindow.loadFile(path.join(__dirname, '../dist/index.html'));
  }

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

// Window control IPC handlers
ipcMain.on('window-close', () => {
  mainWindow?.close();
});

ipcMain.on('window-minimize', () => {
  mainWindow?.minimize();
});

ipcMain.on('window-maximize', () => {
  if (mainWindow?.isMaximized()) {
    mainWindow.unmaximize();
  } else {
    mainWindow?.maximize();
  }
});

// Add handler for hiding command window
ipcMain.on('hide-command', () => {
  commandWindow?.hide();
});

// Register global shortcut when app is ready
app.whenReady().then(() => {
  createWindow();
  
  // Register the tilde shortcut
  console.log('Attempting to register tilde shortcut...');
  const registered = globalShortcut.register('Shift+`', () => {
    console.log('Tilde shortcut triggered!');
    toggleCommandWindow();
  });

  if (registered) {
    console.log('Tilde shortcut registered successfully');
  } else {
    console.log('Failed to register tilde shortcut');
  }
});

// Unregister shortcuts when app is quitting
app.on('will-quit', () => {
  globalShortcut.unregisterAll();
});

// Quit when all windows are closed.
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (!mainWindow) {
    createWindow();
  }
});
