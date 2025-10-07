import React from 'react';
import { Layers, Activity, Search, RefreshCw, Trash2 } from 'lucide-react';
import { Input } from '../../components/ui/input';
import { Button } from '../../components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '../../components/ui/tooltip';
import { AppBarContext } from '../../contexts/AppBarContext';
import { useQuery, useClient } from '../../hooks/api';
import type { components } from '../../api/v2/schema';
import QueueMetrics from '../../features/queues/components/QueueMetrics';
import QueueList from '../../features/queues/components/QueueList';
import { DAGRunDetailsModal } from '../../features/dag-runs/components/dag-run-details';
import ConfirmModal from '../../ui/ConfirmModal';
import { cn } from '../../lib/utils';

function Queues() {
  const appBarContext = React.useContext(AppBarContext);
  const client = useClient();
  const [searchText, setSearchText] = React.useState('');
  const [isRefreshing, setIsRefreshing] = React.useState(false);
  const [selectedQueueType, setSelectedQueueType] =
    React.useState<string>('all');
  const [isClearing, setIsClearing] = React.useState(false);
  const [showClearConfirm, setShowClearConfirm] = React.useState(false);

  const [selectedQueueType, setSelectedQueueType] =
    React.useState<string>('all');

  // State for DAG run modal
  const [modalDAGRun, setModalDAGRun] = React.useState<{
    name: string;
    dagRunId: string;
  } | null>(null);

  React.useEffect(() => {
    appBarContext.setTitle('Queue Dashboard');
  }, [appBarContext]);

  const { data, error, isLoading, mutate } = useQuery('/queues', {
    params: {
      query: {
        remoteNode: appBarContext.selectedRemoteNode || 'local',
      },
    },
    refreshInterval: 3000, // Refresh every 3 seconds for real-time updates
    revalidateOnFocus: true,
    revalidateOnReconnect: true,
  });

  const handleRefresh = async () => {
    setIsRefreshing(true);
    await mutate();
    setTimeout(() => setIsRefreshing(false), 500);
  };

  const handleClearQueue = async () => {
    setIsClearing(true);
    try {
      // Get all queued DAG runs from all queues
      const queuedRuns: Array<{ name: string; dagRunId: string }> = [];

      if (data?.queues) {
        for (const queue of data.queues) {
          if (queue.queued) {
            for (const dagRun of queue.queued) {
              queuedRuns.push({
                name: dagRun.name,
                dagRunId: dagRun.dagRunId,
              });
            }
          }
        }
      }

      // Dequeue all queued DAG runs
      await Promise.all(
        queuedRuns.map(async (dagRun) => {
          try {
            await client.GET('/dag-runs/{name}/{dagRunId}/dequeue', {
              params: {
                path: {
                  name: dagRun.name,
                  dagRunId: dagRun.dagRunId,
                },
                query: {
                  remoteNode: appBarContext.selectedRemoteNode || 'local',
                },
              },
            });
          } catch (error) {
            console.error(
              `Failed to dequeue ${dagRun.name}:${dagRun.dagRunId}:`,
              error
            );
          }
        })
      );

      // Refresh the data
      await mutate();
    } catch (error) {
      console.error('Failed to clear queue:', error);
    } finally {
      setIsClearing(false);
      setShowClearConfirm(false);
    }
  };

  // Filter queues based on search text and type
  const filteredQueues = React.useMemo(() => {
    if (!data?.queues) return [];

    let filtered = data.queues;

    // Filter by search text
    if (searchText) {
      const search = searchText.toLowerCase();
      filtered = filtered.filter((queue) =>
        queue.name.toLowerCase().includes(search)
      );
    }

    // Filter by queue type
    if (selectedQueueType !== 'all') {
      filtered = filtered.filter((queue) => queue.type === selectedQueueType);
    }

    // Sort alphabetically by queue name for stable display
    return filtered.sort((a, b) => a.name.localeCompare(b.name));
  }, [data?.queues, searchText, selectedQueueType]);

  // Calculate metrics
  const metrics = React.useMemo(() => {
    const queues = data?.queues || [];

    // Count queues by type
    const globalQueues = queues.filter((q) => q.type === 'global').length;
    const dagBasedQueues = queues.filter((q) => q.type === 'dag-based').length;

    // Count active queues (those with running or queued items)
    const activeQueues = queues.filter(
      (q) => (q.running?.length || 0) > 0 || (q.queued?.length || 0) > 0
    ).length;

    const totalRunning = queues.reduce(
      (sum, q) => sum + (q.running?.length || 0),
      0
    );
    const totalQueued = queues.reduce(
      (sum, q) => sum + (q.queued?.length || 0),
      0
    );
    const totalActive = totalRunning + totalQueued;

    // Calculate utilization for global queues only (DAG-based queues are isolated and don't compete for shared capacity)
    const globalQueuesList = queues.filter((q) => q.type === 'global');
    const globalRunning = globalQueuesList.reduce(
      (sum, q) => sum + (q.running?.length || 0),
      0
    );
    const globalCapacity = globalQueuesList
      .filter((q) => q.maxConcurrency)
      .reduce((sum, q) => sum + (q.maxConcurrency || 0), 0);
    const utilization =
      globalCapacity > 0
        ? Math.round((globalRunning / globalCapacity) * 100)
        : 0;

    return {
      globalQueues,
      dagBasedQueues,
      activeQueues,
      totalRunning,
      totalQueued,
      totalActive,
      utilization,
    };
  }, [data?.queues]);

  // Handle DAG run click
  const handleDAGRunClick = React.useCallback(
    (dagRun: components['schemas']['DAGRunSummary']) => {
      setModalDAGRun({
        name: dagRun.name,
        dagRunId: dagRun.dagRunId,
      });
    },
    []
  );

  if (error) {
    const errorData = error as components['schemas']['Error'];
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center space-y-2">
          <Layers className="h-12 w-12 text-muted-foreground mx-auto" />
          <p className="text-sm text-muted-foreground">
            {errorData?.message || 'Failed to load queue information'}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3 w-full h-full overflow-hidden">
      {/* Header with search and refresh */}
      <div className="border rounded bg-card flex-shrink-0">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 p-3">
          <div className="flex flex-col sm:flex-row sm:items-center gap-3">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-muted-foreground" />
              <h1 className="text-sm font-semibold">Execution Queues</h1>
            </div>
            <div className="flex items-center gap-2">
              <div className="relative">
                <Search className="absolute left-2 top-1.5 h-3 w-3 text-muted-foreground" />
                <Input
                  type="text"
                  placeholder="Search queues..."
                  value={searchText}
                  onChange={(e) => setSearchText(e.target.value)}
                  className="h-7 w-[200px] pl-7 text-xs"
                />
              </div>
              <select
                value={selectedQueueType}
                onChange={(e) => setSelectedQueueType(e.target.value)}
                className="h-7 px-2 text-xs border rounded bg-background"
              >
                <option value="all">All Types</option>
                <option value="global">Global</option>
                <option value="dag-based">DAG-based</option>
              </select>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {filteredQueues.length !== data?.queues?.length && (
              <span className="text-xs text-muted-foreground">
                ({filteredQueues.length} of {data?.queues?.length})
              </span>
            )}
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowClearConfirm(true)}
                  disabled={isClearing || metrics.totalQueued === 0}
                  className="h-7 px-2 text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/20"
                >
                  <Trash2
                    className={cn('h-3 w-3', isClearing && 'animate-pulse')}
                  />
                  <span className="ml-1 text-xs">Clear Queue</span>
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Remove all queued DAG runs from all queues</p>
              </TooltipContent>
            </Tooltip>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRefresh}
              disabled={isRefreshing}
              className="h-7 px-2"
            >
              <RefreshCw
                className={cn('h-3 w-3', isRefreshing && 'animate-spin')}
              />
              <span className="ml-1 text-xs">Refresh</span>
            </Button>
          </div>
        </div>
      </div>

      {/* Metrics */}
      <QueueMetrics metrics={metrics} isLoading={isLoading} />

      {/* Queue List */}
      <div className="border rounded bg-card flex-1 flex flex-col min-h-0">
        <div className="p-3 border-b flex-shrink-0">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Layers className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm font-semibold">Queue Status</span>
              {filteredQueues.length !== data?.queues?.length && (
                <span className="text-xs text-muted-foreground">
                  ({filteredQueues.length} queues)
                </span>
              )}
            </div>
            <div className="flex items-center gap-4 text-xs text-muted-foreground">
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-green-500" />
                <span>Running</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-purple-500" />
                <span>Queued</span>
              </div>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="flex items-center gap-1 cursor-help">
                    <div className="w-2 h-2 rounded-full bg-blue-500" />
                    <span>Global</span>
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-xs">
                    Shared queues with maxConcurrency limits that can process
                    DAG runs from multiple DAGs
                  </p>
                </TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="flex items-center gap-1 cursor-help">
                    <div className="w-2 h-2 rounded-full bg-gray-500" />
                    <span>DAG-based</span>
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="max-w-xs">
                    Dedicated queues where each DAG has its own queue with
                    maxActiveRuns limit (default 1)
                  </p>
                </TooltipContent>
              </Tooltip>
            </div>
          </div>
        </div>
        <div className="flex-1 min-h-0 overflow-auto">
          <QueueList
            queues={filteredQueues}
            isLoading={isLoading && !data}
            onDAGRunClick={handleDAGRunClick}
            onQueueCleared={handleRefresh}
          />
        </div>
      </div>

      {/* DAG Run Details Modal */}
      {modalDAGRun && (
        <DAGRunDetailsModal
          name={modalDAGRun.name}
          dagRunId={modalDAGRun.dagRunId}
          isOpen={!!modalDAGRun}
          onClose={() => setModalDAGRun(null)}
        />
      )}

      {/* Clear Queue Confirmation Modal */}
      <ConfirmModal
        title="Clear All Queues"
        buttonText="Clear Queue"
        visible={showClearConfirm}
        dismissModal={() => setShowClearConfirm(false)}
        onSubmit={handleClearQueue}
      >
        <div className="space-y-2">
          <p className="text-sm">
            This will remove all queued DAG runs from all queues. This action
            cannot be undone.
          </p>
          <p className="text-xs text-muted-foreground">
            Currently {metrics.totalQueued} DAG runs are queued across all
            queues.
          </p>
        </div>
      </ConfirmModal>
    </div>
  );
}

export default Queues;
