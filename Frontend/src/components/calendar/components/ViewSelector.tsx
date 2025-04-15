import React, { useRef, useEffect, useState } from 'react';
import { cn } from '@/lib/utils';

type ViewType = 'day' | 'threeDays' | 'week' | 'month';

interface ViewSelectorProps {
  currentView: ViewType;
  onViewChange: (view: ViewType) => void;
}

const ViewSelector: React.FC<ViewSelectorProps> = ({ currentView, onViewChange }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [highlightStyle, setHighlightStyle] = useState({
    width: 0,
    transform: 'translateX(0)',
  });

  const views: { value: ViewType; label: string }[] = [
    { value: 'day', label: 'Day' },
    { value: 'threeDays', label: '3 Days' },
    { value: 'week', label: 'Week' },
    { value: 'month', label: 'Month' },
  ];

  useEffect(() => {
    if (containerRef.current) {
      const buttons = containerRef.current.querySelectorAll('button');
      const currentButton = Array.from(buttons).find(
        button => button.dataset.view === currentView
      );

      if (currentButton) {
        const containerLeft = containerRef.current.getBoundingClientRect().left;
        const buttonRect = currentButton.getBoundingClientRect();
        
        setHighlightStyle({
          width: buttonRect.width,
          transform: `translateX(${buttonRect.left - containerLeft}px)`,
        });
      }
    }
  }, [currentView]);

  return (
    <div 
      ref={containerRef}
      className="flex rounded-md shadow-sm relative view-selector-container bg-secondary"
    >
      <div
        className="view-selector-highlight bg-primary rounded-md"
        style={highlightStyle}
      />
      {views.map(({ value, label }) => (
        <button
          key={value}
          data-view={value}
          onClick={() => onViewChange(value)}
          className={cn(
            "px-3 py-1.5 text-sm font-medium relative view-selector-text first:rounded-l-md",
            currentView === value
              ? "text-primary-foreground"
              : "text-secondary-foreground hover:text-secondary-foreground/80"
          )}
        >
          {label}
        </button>
      ))}
    </div>
  );
};

export default ViewSelector;
