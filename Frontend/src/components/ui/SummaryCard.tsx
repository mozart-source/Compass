import React from 'react';
import { cn } from '../../lib/utils';

interface SummaryCardProps {
  title: string;
  value: number;
  icon?: React.ReactNode;
  trend?: number;
  isPercentage?: boolean;
  darkMode?: boolean;
}

const SummaryCard: React.FC<SummaryCardProps> = ({
  title,
  value,
  icon,
  trend,
  isPercentage = false,
  darkMode = false,
}) => {
  return (
    <div
      className={cn(
        'rounded-lg p-6 shadow-sm',
        darkMode ? 'bg-gray-800 text-white' : 'bg-white text-gray-900'
      )}
    >
      <div className="flex items-center justify-between mb-4">
        <h3 className={cn(
          'text-sm font-medium',
          darkMode ? 'text-gray-300' : 'text-gray-500'
        )}>
          {title}
        </h3>
        {icon && (
          <div className={cn(
            'p-2 rounded-full',
            darkMode ? 'bg-gray-700' : 'bg-gray-100'
          )}>
            {icon}
          </div>
        )}
      </div>
      <div className="flex items-end justify-between">
        <div>
          <p className="text-2xl font-semibold">
            {isPercentage ? `${value.toFixed(1)}%` : value.toLocaleString()}
          </p>
          {trend !== undefined && (
            <div className={cn(
              'flex items-center mt-2 text-sm',
              trend >= 0 ? 'text-green-500' : 'text-red-500'
            )}>
              <span className="mr-1">
                {trend >= 0 ? '↑' : '↓'}
              </span>
              <span>{Math.abs(trend)}%</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SummaryCard;
