/**
 * Exports Page
 * 데이터 내보내기 관리 페이지
 */

'use client';

import { useState } from 'react';
import { Download, Settings } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ExportDialog } from '@/components/export/ExportDialog';
import { ExportHistory } from '@/components/export/ExportHistory';
import { useRequireAuth } from '@/hooks/use-auth';

export default function ExportsPage() {
  const [exportDialogOpen, setExportDialogOpen] = useState(false);
  const [selectedType, setSelectedType] = useState<'vms' | 'workspaces' | 'credentials' | 'audit_logs' | 'costs'>('vms');

  // 인증 확인
  useRequireAuth();

  const handleQuickExport = (type: typeof selectedType) => {
    setSelectedType(type);
    setExportDialogOpen(true);
  };

  const exportTypes = [
    {
      id: 'vms',
      name: '가상머신',
      description: 'VM 목록과 상세 정보를 내보냅니다.',
      icon: '🖥️',
    },
    {
      id: 'workspaces',
      name: '워크스페이스',
      description: '워크스페이스와 멤버 정보를 내보냅니다.',
      icon: '🏢',
    },
    {
      id: 'credentials',
      name: '자격증명',
      description: '클라우드 자격증명 정보를 내보냅니다.',
      icon: '🔐',
    },
    {
      id: 'audit_logs',
      name: '감사 로그',
      description: '사용자 활동 로그를 내보냅니다.',
      icon: '📋',
    },
    {
      id: 'costs',
      name: '비용 분석',
      description: '비용 분석 데이터를 내보냅니다.',
      icon: '💰',
    },
  ] as const;

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* 헤더 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">데이터 내보내기</h1>
          <p className="text-gray-600 mt-2">
            원하는 데이터를 다양한 형식으로 내보낼 수 있습니다.
          </p>
        </div>
        <Button onClick={() => setExportDialogOpen(true)}>
          <Download className="mr-2 h-4 w-4" />
          새로 내보내기
        </Button>
      </div>

      {/* 메인 컨텐츠 */}
      <Tabs defaultValue="quick" className="space-y-6">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="quick">빠른 내보내기</TabsTrigger>
          <TabsTrigger value="history">내보내기 이력</TabsTrigger>
          <TabsTrigger value="settings">설정</TabsTrigger>
        </TabsList>

        {/* 빠른 내보내기 */}
        <TabsContent value="quick" className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {exportTypes.map((type) => (
              <Card key={type.id} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="flex items-center gap-3">
                    <span className="text-2xl">{type.icon}</span>
                    <div>
                      <CardTitle className="text-lg">{type.name}</CardTitle>
                      <CardDescription>{type.description}</CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <Button
                    onClick={() => handleQuickExport(type.id as typeof selectedType)}
                    className="w-full"
                    variant="outline"
                  >
                    <Download className="mr-2 h-4 w-4" />
                    내보내기
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* 사용 팁 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-5 w-5" />
                사용 팁
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                <div>
                  <h4 className="font-semibold mb-2">📊 파일 형식 선택</h4>
                  <ul className="space-y-1 text-gray-600">
                    <li>• <strong>CSV:</strong> Excel에서 열기, 간단한 데이터</li>
                    <li>• <strong>JSON:</strong> API 연동, 프로그래밍</li>
                    <li>• <strong>Excel:</strong> 차트 포함, 복잡한 데이터</li>
                    <li>• <strong>PDF:</strong> 인쇄, 공식 문서</li>
                  </ul>
                </div>
                <div>
                  <h4 className="font-semibold mb-2">🔍 필터링 옵션</h4>
                  <ul className="space-y-1 text-gray-600">
                    <li>• <strong>날짜 범위:</strong> 특정 기간 데이터</li>
                    <li>• <strong>워크스페이스:</strong> 특정 팀 데이터</li>
                    <li>• <strong>JSON 필터:</strong> 고급 조건</li>
                    <li>• <strong>삭제된 항목:</strong> 포함 여부 선택</li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 내보내기 이력 */}
        <TabsContent value="history">
          <ExportHistory limit={20} showActions={true} />
        </TabsContent>

        {/* 설정 */}
        <TabsContent value="settings" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>내보내기 설정</CardTitle>
              <CardDescription>
                데이터 내보내기 관련 설정을 관리합니다.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <h4 className="font-semibold mb-2">기본 설정</h4>
                  <p className="text-sm text-gray-600">
                    기본 파일 형식과 데이터 타입을 설정할 수 있습니다.
                  </p>
                </div>
                <div>
                  <h4 className="font-semibold mb-2">자동 정리</h4>
                  <p className="text-sm text-gray-600">
                    30일 이상 된 내보내기 파일은 자동으로 삭제됩니다.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>지원되는 형식</CardTitle>
              <CardDescription>
                현재 지원되는 파일 형식과 데이터 타입입니다.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <h4 className="font-semibold mb-3">파일 형식</h4>
                  <div className="space-y-2">
                    {['CSV', 'JSON', 'Excel (XLSX)', 'PDF'].map((format) => (
                      <div key={format} className="flex items-center gap-2">
                        <div className="w-2 h-2 bg-green-500 rounded-full" />
                        <span className="text-sm">{format}</span>
                      </div>
                    ))}
                  </div>
                </div>
                <div>
                  <h4 className="font-semibold mb-3">데이터 타입</h4>
                  <div className="space-y-2">
                    {exportTypes.map((type) => (
                      <div key={type.id} className="flex items-center gap-2">
                        <div className="w-2 h-2 bg-blue-500 rounded-full" />
                        <span className="text-sm">{type.name}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 내보내기 다이얼로그 */}
      <ExportDialog
        open={exportDialogOpen}
        onOpenChange={setExportDialogOpen}
        defaultType={selectedType}
      />
    </div>
  );
}
