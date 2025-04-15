import React from 'react';
import { ChevronUp, ChevronDown, Minus } from 'lucide-react';
import cn from 'classnames';
import { TodoPriority } from '@/components/todo/types-todo';

interface PriorityIndicatorProps {
  priority: string;
}

const PriorityIndicator: React.FC<PriorityIndicatorProps> = ({ priority }) => {
  const getIcon = () => {
    switch (priority.toLowerCase()) {
      case TodoPriority.HIGH:
        return <ChevronUp className="h-4 w-4" />;
      case TodoPriority.MEDIUM:
        return <Minus className="h-4 w-4" />;
      case TodoPriority.LOW:
        return <ChevronDown className="h-4 w-4" />;
      default:
        return <Minus className="h-4 w-4" />; // Default to medium
    }
  };

  const getPriorityClass = () => {
    switch (priority.toLowerCase()) {
      case TodoPriority.HIGH:
        return 'text-red-500';
      case TodoPriority.MEDIUM:
        return 'text-white-500';
      case TodoPriority.LOW:
        return 'text-green-500';
      default:
        return 'text-white-500'; // Default to medium
    }
  };

  return (
    <div className={cn('flex items-center justify-center mr-2', getPriorityClass())}>
      {getIcon()}
    </div>
  );
};

export default PriorityIndicator;