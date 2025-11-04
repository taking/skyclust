'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { CheckCircle, XCircle, Loader2, X } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

interface BulkOperationProgressProps {
  operation: 'delete' | 'tag';
  total: number;
  completed: number;
  failed: number;
  isComplete: boolean;
  isCancelled?: boolean;
  cancelled?: number;
  onDismiss?: () => void;
  onCancel?: () => void;
}

export function BulkOperationProgress({
  operation,
  total,
  completed,
  failed,
  isComplete,
  isCancelled = false,
  cancelled = 0,
  onDismiss,
  onCancel,
}: BulkOperationProgressProps) {
  const progress = total > 0 ? ((completed + failed + cancelled) / total) * 100 : 0;
  const successCount = completed;
  
  const getOperationLabel = () => {
    switch (operation) {
      case 'delete':
        return 'Deleting';
      case 'tag':
        return 'Tagging';
      default:
        return 'Processing';
    }
  };

  // Cancelled state
  if (isCancelled) {
    return (
      <Card className="border-gray-200 bg-gray-50">
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <XCircle className="h-5 w-5 text-gray-600" />
              <div>
                <p className="text-sm font-medium text-gray-900">
                  Operation cancelled
                </p>
                <p className="text-xs text-gray-700">
                  {completed} completed, {failed} failed, {cancelled} cancelled
                </p>
              </div>
            </div>
            {onDismiss && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onDismiss}
                aria-label="Dismiss"
              >
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (isComplete && successCount === total && failed === 0) {
    return (
      <Card className="border-green-200 bg-green-50">
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <CheckCircle className="h-5 w-5 text-green-600" />
              <div>
                <p className="text-sm font-medium text-green-900">
                  Successfully {operation === 'delete' ? 'deleted' : 'tagged'} {successCount} item(s)
                </p>
                <p className="text-xs text-green-700">
                  Operation completed successfully
                </p>
              </div>
            </div>
            {onDismiss && (
              <button
                onClick={onDismiss}
                className="text-green-700 hover:text-green-900"
                aria-label="Dismiss"
              >
                <XCircle className="h-4 w-4" />
              </button>
            )}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (isComplete && failed > 0) {
    return (
      <Card className="border-yellow-200 bg-yellow-50">
        <CardContent className="pt-6">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <XCircle className="h-5 w-5 text-yellow-600" />
                <div>
                  <p className="text-sm font-medium text-yellow-900">
                    Operation completed with errors
                  </p>
                  <p className="text-xs text-yellow-700">
                    {successCount} succeeded, {failed} failed
                  </p>
                </div>
              </div>
              {onDismiss && (
                <button
                  onClick={onDismiss}
                  className="text-yellow-700 hover:text-yellow-900"
                  aria-label="Dismiss"
                >
                  <XCircle className="h-4 w-4" />
                </button>
              )}
            </div>
            <div className="flex space-x-4 text-xs">
              <Badge variant="outline" className="bg-green-100 text-green-800">
                {successCount} Successful
              </Badge>
              <Badge variant="outline" className="bg-red-100 text-red-800">
                {failed} Failed
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Loader2 className="h-4 w-4 animate-spin" />
          <span>{getOperationLabel()} Items</span>
        </CardTitle>
        <CardDescription>
          Processing {completed + failed} of {total} items
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span>Progress</span>
            <span>{Math.round(progress)}%</span>
          </div>
          <Progress value={progress} className="h-2" />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex space-x-4 text-xs text-gray-600">
            <span>Completed: {completed}</span>
            {failed > 0 && <span className="text-red-600">Failed: {failed}</span>}
            {cancelled > 0 && <span className="text-gray-500">Cancelled: {cancelled}</span>}
            <span>Remaining: {total - completed - failed - cancelled}</span>
          </div>
          {onCancel && !isComplete && (
            <Button
              variant="outline"
              size="sm"
              onClick={onCancel}
              className="text-red-600 hover:text-red-700"
            >
              <X className="mr-2 h-4 w-4" />
              Cancel
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

