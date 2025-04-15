import { contextBridge, ipcRenderer } from 'electron';

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld('electron', {
  close: () => ipcRenderer.send('window-close'),
  minimize: () => ipcRenderer.send('window-minimize'),
  maximize: () => ipcRenderer.send('window-maximize'),
  hideCommand: () => ipcRenderer.send('hide-command'),
  onToggleCommandModal: (callback: () => void) => {
    console.log('Registering toggle-command-modal listener');
    const listener = () => {
      console.log('Received toggle-command-modal event');
      callback();
    };
    ipcRenderer.on('toggle-command-modal', listener);
    return () => {
      console.log('Removing toggle-command-modal listener');
      ipcRenderer.removeListener('toggle-command-modal', listener);
    };
  }
});
