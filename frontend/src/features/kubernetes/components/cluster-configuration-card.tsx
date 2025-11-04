/**
 * Cluster Configuration Card Component
 * 클러스터 설정 카드 컴포넌트
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { Credential } from '@/lib/types';

interface ClusterConfigurationCardProps {
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  selectedRegion: string;
  onRegionChange: (region: string) => void;
}

export function ClusterConfigurationCard({
  credentials,
  selectedCredentialId,
  onCredentialChange,
  selectedRegion,
  onRegionChange,
  onFormValueChange,
}: ClusterConfigurationCardProps) {
  const handleCredentialChange = (value: string) => {
    onCredentialChange(value);
    onFormValueChange('credential_id', value);
  };

  const handleRegionChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    onRegionChange(value);
    onFormValueChange('region', value);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Configuration</CardTitle>
        <CardDescription>Select credential and region to view cluster details</CardDescription>
      </CardHeader>
      <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Credential</Label>
          <Select value={selectedCredentialId} onValueChange={handleCredentialChange}>
            <SelectTrigger>
              <SelectValue placeholder="Select credential" />
            </SelectTrigger>
            <SelectContent>
              {credentials.map((cred) => (
                <SelectItem key={cred.id} value={cred.id}>
                  {cred.provider} - {cred.id.substring(0, 8)}...
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>Region</Label>
          <Input
            value={selectedRegion}
            onChange={handleRegionChange}
            placeholder="ap-northeast-2"
          />
        </div>
      </CardContent>
    </Card>
  );
}

