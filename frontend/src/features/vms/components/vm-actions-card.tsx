/**
 * VM Actions Card Component
 * VM 상세 페이지의 액션 버튼 카드 컴포넌트
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Play, Pause, RotateCcw } from 'lucide-react';
import type { VM } from '@/lib/types';

interface VMActionsCardProps {
  vm: VM;
  onStart: () => void;
  onStop: () => void;
  onRestart: () => void;
  isStarting: boolean;
  isStopping: boolean;
}

export function VMActionsCard({
  vm,
  onStart,
  onStop,
  onRestart,
  isStarting,
  isStopping,
}: VMActionsCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>VM Actions</CardTitle>
        <CardDescription>Control your virtual machine</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex space-x-2">
          <Button
            onClick={onStart}
            disabled={vm.status === 'running' || isStarting}
            className="flex-1"
          >
            <Play className="mr-2 h-4 w-4" />
            {isStarting ? 'Starting...' : 'Start'}
          </Button>
          <Button
            variant="outline"
            onClick={onStop}
            disabled={vm.status === 'stopped' || isStopping}
            className="flex-1"
          >
            <Pause className="mr-2 h-4 w-4" />
            {isStopping ? 'Stopping...' : 'Stop'}
          </Button>
          <Button
            variant="outline"
            onClick={onRestart}
            disabled={vm.status === 'stopped'}
            className="flex-1"
          >
            <RotateCcw className="mr-2 h-4 w-4" />
            Restart
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

