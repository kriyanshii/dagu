import React from 'react';
import { Tab, Tabs } from '@/components/ui/tabs';
import {
  Play,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Pause,
  Activity,
} from 'lucide-react';

interface StatusSearchTabsProps {
  activeStatus: string;
  onStatusChange: (status: string) => void;
  statusCounts?: Record<string, number>;
  className?: string;
}

const statusOptions = [
  { value: 'all', label: 'All', icon: Activity, color: 'text-gray-500' },
  { value: '0', label: 'Not Started', icon: Clock, color: 'text-gray-500' },
  { value: '1', label: 'Running', icon: Play, color: 'text-[limegreen]' },
  { value: '2', label: 'Failed', icon: XCircle, color: 'text-red-500' },
  { value: '3', label: 'Cancelled', icon: Pause, color: 'text-orange-500' },
  { value: '4', label: 'Success', icon: CheckCircle, color: 'text-green-500' },
  { value: '5', label: 'Queued', icon: Clock, color: 'text-[purple]' },
  {
    value: '6',
    label: 'Partial Success',
    icon: AlertCircle,
    color: 'text-yellow-500',
  },
];

const StatusSearchTabs: React.FC<StatusSearchTabsProps> = ({
  activeStatus,
  onStatusChange,
  statusCounts = {},
  className = '',
}) => {
  return (
    <div className={`w-full ${className}`}>
      <Tabs className="p-1 w-full overflow-x-auto">
        {statusOptions.map((option) => {
          const Icon = option.icon;
          const count = statusCounts[option.value] || 0;
          const isActive = activeStatus === option.value;

          return (
            <Tab
              key={option.value}
              isActive={isActive}
              onClick={() => onStatusChange(option.value)}
              className="flex items-center gap-2 cursor-pointer whitespace-nowrap min-w-fit"
            >
              <Icon className={`h-4 w-4 ${option.color}`} />
              <span className="text-sm font-medium">{option.label}</span>
              {count > 0 && (
                <span className="text-xs bg-muted px-2 py-0.5 rounded-full">
                  {count}
                </span>
              )}
            </Tab>
          );
        })}
      </Tabs>
    </div>
  );
};

export default StatusSearchTabs;
