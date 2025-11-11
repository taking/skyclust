/**
 * CreateResourcePageHeader Component
 * 리소스 생성 페이지의 공통 Header 컴포넌트
 */

'use client';

import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { useRouter } from 'next/navigation';

export interface CreateResourcePageHeaderProps {
  /**
   * 뒤로 가기 버튼 클릭 시 이동할 경로
   */
  backPath: string;
  
  /**
   * 페이지 제목 (i18n key 또는 직접 텍스트)
   */
  title: string;
  
  /**
   * 페이지 설명 (i18n key 또는 직접 텍스트)
   */
  description?: string;
  
  /**
   * 설명에 사용할 파라미터 (i18n 사용 시)
   */
  descriptionParams?: Record<string, string | number>;
  
  /**
   * 취소 핸들러 (선택적, 기본값은 backPath로 이동)
   */
  onCancel?: () => void;
}

export function CreateResourcePageHeader({
  backPath,
  title,
  description,
  descriptionParams,
  onCancel,
}: CreateResourcePageHeaderProps) {
  const { t } = useTranslation();
  const router = useRouter();

  const handleCancel = () => {
    if (onCancel) {
      onCancel();
    } else {
      router.push(backPath);
    }
  };

  const titleText = title.startsWith('common.') || title.includes('.') 
    ? t(title) 
    : title;
  
  const descriptionText = description 
    ? (description.startsWith('common.') || description.includes('.')
        ? t(description, descriptionParams)
        : description)
    : undefined;

  return (
    <div className="mb-8">
      <Button
        variant="ghost"
        onClick={handleCancel}
        className="mb-4"
      >
        <ArrowLeft className="mr-2 h-4 w-4" />
        {t('common.back')}
      </Button>
      <h1 className="text-3xl font-bold text-gray-900">{titleText}</h1>
      {descriptionText && (
        <p className="text-gray-600 mt-2">{descriptionText}</p>
      )}
    </div>
  );
}


