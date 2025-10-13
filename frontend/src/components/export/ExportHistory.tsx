/**
 * Export History Component
 * 내보내기 이력 관리 컴포넌트
 */

'use client';

import { useState } from 'react';
import { format } from 'date-fns';
import { ko } from 'date-fns/locale';
import {
  Download,
  FileText,
  RefreshCw,
  Eye,
  AlertCircle,
  CheckCircle,
  Clock,
  XCircle,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { useExportHistory, useDownloadExport } from '@/hooks/useExport';
import { exportService, ExportResult } from '@/services/export';
import { toast } from 'react-hot-toast';

interface ExportHistoryProps {
  limit?: number;
  showActions?: boolean;
}

export function ExportHistory({ limit = 20, showActions = true }: ExportHistoryProps) {
  const [offset, setOffset] = useState(0);
  const { data: history, isLoading, error, refetch } = useExportHistory(limit, offset);
  const downloadMutation = useDownloadExport();

  const handleDownload = async (exportItem: ExportResult) => {
    if (exportItem.status !== 'completed' || !exportItem.download_url) {
      toast.error('다운로드할 수 없는 파일입니다.');
      return;
    }

    try {
      await downloadMutation.mutateAsync(exportItem.id);
    } catch (error) {
      console.error('Download failed:', error);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case 'processing':
        return <RefreshCw className="h-4 w-4 text-blue-600 animate-spin" />;
      case 'pending':
        return <Clock className="h-4 w-4 text-yellow-600" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-600" />;
      default:
        return <AlertCircle className="h-4 w-4 text-gray-600" />;
    }
  };

  const getStatusText = (status: string) => {
    const statusMap: Record<string, string> = {
      completed: '완료',
      processing: '처리 중',
      pending: '대기 중',
      failed: '실패',
    };
    return statusMap[status] || status;
  };

  const getTypeText = (type: string) => {
    const typeMap: Record<string, string> = {
      vms: '가상머신',
      workspaces: '워크스페이스',
      credentials: '자격증명',
      audit_logs: '감사 로그',
      costs: '비용 분석',
    };
    return typeMap[type] || type;
  };

  const getFormatText = (format: string) => {
    return format.toUpperCase();
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>내보내기 이력</CardTitle>
          <CardDescription>데이터 내보내기 기록을 확인할 수 있습니다.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center space-x-4">
                <Skeleton className="h-4 w-4" />
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-4 w-20" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>내보내기 이력</CardTitle>
          <CardDescription>데이터 내보내기 기록을 확인할 수 있습니다.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <div className="text-center">
              <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
              <p className="text-red-600 mb-4">이력을 불러올 수 없습니다.</p>
              <Button onClick={() => refetch()} variant="outline">
                <RefreshCw className="mr-2 h-4 w-4" />
                다시 시도
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!history || history.exports.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>내보내기 이력</CardTitle>
          <CardDescription>데이터 내보내기 기록을 확인할 수 있습니다.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <div className="text-center">
              <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600">아직 내보내기 이력이 없습니다.</p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>내보내기 이력</CardTitle>
            <CardDescription>
              총 {history.total}개의 내보내기 기록
            </CardDescription>
          </div>
          <Button onClick={() => refetch()} variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            새로고침
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>상태</TableHead>
                <TableHead>타입</TableHead>
                <TableHead>형식</TableHead>
                <TableHead>파일 크기</TableHead>
                <TableHead>생성일</TableHead>
                <TableHead>완료일</TableHead>
                {showActions && <TableHead className="text-right">작업</TableHead>}
              </TableRow>
            </TableHeader>
            <TableBody>
              {history.exports.map((exportItem) => (
                <TableRow key={exportItem.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getStatusIcon(exportItem.status)}
                      <Badge
                        variant="outline"
                        className={exportService.getStatusBgColor(exportItem.status)}
                      >
                        {getStatusText(exportItem.status)}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="font-medium">
                      {getTypeText(exportItem.type)}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">
                      {getFormatText(exportItem.format)}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {exportItem.file_size
                      ? exportService.formatFileSize(exportItem.file_size)
                      : '-'}
                  </TableCell>
                  <TableCell>
                    {format(new Date(exportItem.created_at), 'yyyy-MM-dd HH:mm', {
                      locale: ko,
                    })}
                  </TableCell>
                  <TableCell>
                    {exportItem.completed_at
                      ? format(new Date(exportItem.completed_at), 'yyyy-MM-dd HH:mm', {
                          locale: ko,
                        })
                      : '-'}
                  </TableCell>
                  {showActions && (
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-2">
                        {exportItem.status === 'completed' && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleDownload(exportItem)}
                            disabled={downloadMutation.isPending}
                          >
                            <Download className="h-4 w-4" />
                          </Button>
                        )}
                        {exportItem.status === 'failed' && exportItem.error && (
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <Button size="sm" variant="outline">
                                <Eye className="h-4 w-4" />
                              </Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                              <AlertDialogHeader>
                                <AlertDialogTitle>오류 상세</AlertDialogTitle>
                                <AlertDialogDescription>
                                  내보내기 실패 원인을 확인하세요.
                                </AlertDialogDescription>
                              </AlertDialogHeader>
                              <div className="py-4">
                                <pre className="whitespace-pre-wrap text-sm bg-gray-100 p-3 rounded">
                                  {exportItem.error}
                                </pre>
                              </div>
                              <AlertDialogFooter>
                                <AlertDialogCancel>닫기</AlertDialogCancel>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        )}
                      </div>
                    </TableCell>
                  )}
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {/* 페이지네이션 */}
          {history.total > limit && (
            <div className="flex items-center justify-between pt-4">
              <div className="text-sm text-gray-600">
                {offset + 1}-{Math.min(offset + limit, history.total)} / {history.total}
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setOffset(Math.max(0, offset - limit))}
                  disabled={offset === 0}
                >
                  이전
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setOffset(offset + limit)}
                  disabled={offset + limit >= history.total}
                >
                  다음
                </Button>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
