import React from 'react';
import { X, Minus, Square } from 'lucide-react';
import { cn } from '@/lib/utils';

interface TitleBarProps {
  darkMode?: boolean;
}

const TitleBar: React.FC<TitleBarProps> = ({ darkMode = false }) => {

  const handleClose = () => {
    if (window.electron) {
      window.electron.close();
    }
  };

  const handleMinimize = () => {
    if (window.electron) {
      window.electron.minimize();
    }
  };

  const handleMaximize = () => {
    if (window.electron) {
      window.electron.maximize();
    }
  };

  return (
    <div className={cn(
      "fixed top-0 left-0 w-full h-11 flex items-center justify-between px-4 select-none z-10 drag-region bg-background"
    )}>
      <div className="flex items-center">
      </div>
      <div className={cn("flex items-center no-drag space-x-2",
      )}>
        <button
          onClick={handleMinimize}
          className={cn(
            "p-1 rounded transition-colors",
            darkMode 
              ? 'hover:bg-[#2c2c2e] text-gray-400 hover:text-white' 
              : 'hover:bg-gray-100 text-gray-600 hover:text-gray-900'
          )}
        >
          <Minus className="w-4 h-4" />
        </button>
        <button
          onClick={handleMaximize}
          className={cn(
            "p-1 rounded transition-colors",
            darkMode 
              ? 'hover:bg-[#2c2c2e] text-gray-400 hover:text-white' 
              : 'hover:bg-gray-100 text-gray-600 hover:text-gray-900'
          )}
        >
          <Square className="w-4 h-4" />
        </button>
        <button
          onClick={handleClose}
          className="p-1 hover:bg-red-500 hover:text-white rounded transition-colors"
        >
          <X className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
};

export default TitleBar;