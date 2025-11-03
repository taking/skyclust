'use client';

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { CloudProvider } from '@/lib/types';
import { useProviderStore } from '@/store/provider';

const providers: { value: CloudProvider; label: string }[] = [
  { value: 'aws', label: 'AWS' },
  { value: 'gcp', label: 'Google Cloud Platform' },
  { value: 'azure', label: 'Azure' },
  { value: 'ncp', label: 'Naver Cloud Platform' },
];

interface ProviderSelectorProps {
  value?: CloudProvider | null;
  onValueChange?: (value: CloudProvider) => void;
  disabled?: boolean;
}

export function ProviderSelector({ value, onValueChange, disabled }: ProviderSelectorProps) {
  const { selectedProvider, setSelectedProvider } = useProviderStore();

  const currentValue = value ?? selectedProvider;

  const handleChange = (newValue: CloudProvider) => {
    if (onValueChange) {
      onValueChange(newValue);
    } else {
      setSelectedProvider(newValue);
    }
  };

  return (
    <Select
      value={currentValue || undefined}
      onValueChange={handleChange}
      disabled={disabled}
    >
      <SelectTrigger className="w-[200px]">
        <SelectValue placeholder="Select Provider" />
      </SelectTrigger>
      <SelectContent>
        {providers.map((provider) => (
          <SelectItem key={provider.value} value={provider.value}>
            {provider.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

