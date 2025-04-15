declare global {
  interface Window {
    electron: {
      close: () => void;
      minimize: () => void;
      maximize: () => void;
      hideCommand: () => void;
      onToggleCommandModal: (callback: () => void) => () => void;
    };
  }
}

export {}; 