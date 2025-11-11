/**
 * Export Dialog Component
 * 데이터 내보내기 다이얼로그
 */

'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { CalendarIcon, Download } from 'lucide-react';
import { format } from 'date-fns';
import { ko } from 'date-fns/locale';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { Calendar } from '@/components/ui/calendar';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { cn } from '@/lib/utils';
import { log } from '@/lib/logging';
import { useExportData } from '@/hooks/use-export';
import type { ExportRequest } from '@/lib/types/export';

const exportSchema = z.object({
  type: z.enum(['vms', 'workspaces', 'credentials', 'audit_logs', 'costs']),
  format: z.enum(['csv', 'json', 'xlsx', 'pdf']),
  workspace_id: z.string().optional(),
  date_from: z.date().optional(),
  date_to: z.date().optional(),
  include_deleted: z.boolean().default(false),
  filters: z.string().optional(),
});

type ExportFormData = z.infer<typeof exportSchema>;

interface ExportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  defaultType?: ExportRequest['type'];
  workspaceId?: string;
}

export function ExportDialog({
  open,
  onOpenChange,
  defaultType,
  workspaceId,
}: ExportDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const exportDataMutation = useExportData();

  const form = useForm({
    resolver: zodResolver(exportSchema),
    defaultValues: {
      type: defaultType || 'vms',
      format: 'csv',
      workspace_id: workspaceId,
      include_deleted: false,
    },
  });

  const selectedType = form.watch('type');
  const selectedFormat = form.watch('format');

  const onSubmit = async (data: ExportFormData) => {
    setIsSubmitting(true);
    try {
      const request: Omit<ExportRequest, 'user_id'> = {
        type: data.type,
        format: data.format,
        workspace_id: data.workspace_id,
        date_from: data.date_from?.toISOString().split('T')[0],
        date_to: data.date_to?.toISOString().split('T')[0],
        include_deleted: data.include_deleted,
        filters: data.filters ? JSON.parse(data.filters) : undefined,
      };

      await exportDataMutation.mutateAsync(request);
      onOpenChange(false);
      form.reset();
    } catch (error) {
      log.error('Export failed', error instanceof Error ? error : new Error(String(error)));
    } finally {
      setIsSubmitting(false);
    }
  };

  const getTypeDescription = (type: string) => {
    const descriptions: Record<string, string> = {
      vms: '가상머신 목록과 상세 정보',
      workspaces: '워크스페이스 목록과 멤버 정보',
      credentials: '클라우드 자격증명 정보',
      audit_logs: '사용자 활동 로그',
      costs: '비용 분석 데이터',
    };
    return descriptions[type] || '';
  };

  const getFormatDescription = (format: string) => {
    const descriptions: Record<string, string> = {
      csv: 'Excel에서 열 수 있는 CSV 형식',
      json: 'API 연동에 적합한 JSON 형식',
      xlsx: 'Excel 파일 형식 (차트 포함)',
      pdf: '인쇄에 적합한 PDF 형식',
    };
    return descriptions[format] || '';
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Download className="h-5 w-5" />
            데이터 내보내기
          </DialogTitle>
          <DialogDescription>
            원하는 데이터를 선택한 형식으로 내보낼 수 있습니다.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* 데이터 타입 선택 */}
              <FormField
                control={form.control}
                name="type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>데이터 타입</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="데이터 타입을 선택하세요" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="vms">가상머신</SelectItem>
                        <SelectItem value="workspaces">워크스페이스</SelectItem>
                        <SelectItem value="credentials">자격증명</SelectItem>
                        <SelectItem value="audit_logs">감사 로그</SelectItem>
                        <SelectItem value="costs">비용 분석</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      {getTypeDescription(selectedType)}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* 파일 형식 선택 */}
              <FormField
                control={form.control}
                name="format"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>파일 형식</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="파일 형식을 선택하세요" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="csv">CSV</SelectItem>
                        <SelectItem value="json">JSON</SelectItem>
                        <SelectItem value="xlsx">Excel (XLSX)</SelectItem>
                        <SelectItem value="pdf">PDF</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      {getFormatDescription(selectedFormat)}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* 워크스페이스 선택 (선택사항) */}
              <FormField
                control={form.control}
                name="workspace_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>워크스페이스 (선택사항)</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="워크스페이스 ID를 입력하세요"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      특정 워크스페이스의 데이터만 내보내려면 ID를 입력하세요.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

            {/* 날짜 범위 선택 */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="date_from"
                render={({ field }) => (
                  <FormItem className="flex flex-col">
                    <FormLabel>시작 날짜</FormLabel>
                    <Popover>
                      <PopoverTrigger asChild>
                        <FormControl>
                          <Button
                            variant="outline"
                            className={cn(
                              'w-full pl-3 text-left font-normal',
                              !field.value && 'text-muted-foreground'
                            )}
                          >
                            {field.value ? (
                              format(field.value, 'PPP', { locale: ko })
                            ) : (
                              <span>날짜를 선택하세요</span>
                            )}
                            <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                          </Button>
                        </FormControl>
                      </PopoverTrigger>
                      <PopoverContent className="w-auto p-0" align="start">
                        <Calendar
                          mode="single"
                          selected={field.value}
                          onSelect={field.onChange}
                          disabled={(date) =>
                            date > new Date() || date < new Date('1900-01-01')
                          }
                          initialFocus
                        />
                      </PopoverContent>
                    </Popover>
                    <FormDescription>
                      데이터 범위의 시작 날짜입니다.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="date_to"
                render={({ field }) => (
                  <FormItem className="flex flex-col">
                    <FormLabel>종료 날짜</FormLabel>
                    <Popover>
                      <PopoverTrigger asChild>
                        <FormControl>
                          <Button
                            variant="outline"
                            className={cn(
                              'w-full pl-3 text-left font-normal',
                              !field.value && 'text-muted-foreground'
                            )}
                          >
                            {field.value ? (
                              format(field.value, 'PPP', { locale: ko })
                            ) : (
                              <span>날짜를 선택하세요</span>
                            )}
                            <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                          </Button>
                        </FormControl>
                      </PopoverTrigger>
                      <PopoverContent className="w-auto p-0" align="start">
                        <Calendar
                          mode="single"
                          selected={field.value}
                          onSelect={field.onChange}
                          disabled={(date) =>
                            date > new Date() || date < new Date('1900-01-01')
                          }
                          initialFocus
                        />
                      </PopoverContent>
                    </Popover>
                    <FormDescription>
                      데이터 범위의 종료 날짜입니다.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* 추가 옵션 */}
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="include_deleted"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <div className="space-y-1 leading-none">
                      <FormLabel>삭제된 항목 포함</FormLabel>
                      <FormDescription>
                        삭제된 데이터도 내보내기에 포함합니다.
                      </FormDescription>
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="filters"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>추가 필터 (JSON 형식)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='{"status": "running", "provider": "aws"}'
                        className="min-h-[100px] font-mono text-sm"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      추가적인 필터 조건을 JSON 형식으로 입력하세요.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={isSubmitting}
              >
                취소
              </Button>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? (
                  <>
                    <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    내보내기 중...
                  </>
                ) : (
                  <>
                    <Download className="mr-2 h-4 w-4" />
                    내보내기 시작
                  </>
                )}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
