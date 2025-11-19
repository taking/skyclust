/**
 * AWS Network Configuration Component
 * AWS EKS는 network step에서 추가 필드가 필요 없음
 */

'use client';

import type { UseFormReturn } from 'react-hook-form';
import type { CreateClusterForm } from '@/lib/types';

interface AWSNetworkConfigProps {
  form: UseFormReturn<CreateClusterForm>;
  onDataChange: (data: Partial<CreateClusterForm>) => void;
  selectedVPCId: string;
  selectedSubnetIds: string[];
}

export function AWSNetworkConfig({
  form: _form,
  onDataChange: _onDataChange,
  selectedVPCId: _selectedVPCId,
  selectedSubnetIds: _selectedSubnetIds,
}: AWSNetworkConfigProps) {
  // AWS는 network step에서 추가 필드가 필요 없음
  return null;
}

