/**
 * Provider Form Fields Component
 * 각 CSP별 폼 필드 컴포넌트
 */

'use client';

import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { UseFormRegister, UseFormSetValue } from 'react-hook-form';
import type { CreateCredentialForm } from '@/lib/types';

interface ProviderFormFieldsProps {
  provider: string;
  gcpInputMode?: 'json' | 'file';
  onGcpInputModeChange?: (mode: 'json' | 'file') => void;
  register: UseFormRegister<CreateCredentialForm>;
  setValue: UseFormSetValue<CreateCredentialForm>;
}

export function ProviderFormFields({
  provider,
  gcpInputMode = 'json',
  onGcpInputModeChange,
  register,
  setValue,
}: ProviderFormFieldsProps) {
  if (provider === 'aws') {
    return (
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="access_key">Access Key ID *</Label>
          <Input
            id="access_key"
            placeholder="AKIA..."
            {...register('credentials.access_key')}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="secret_key">Secret Access Key *</Label>
          <Input
            id="secret_key"
            type="password"
            placeholder="Enter secret key"
            {...register('credentials.secret_key')}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="region">Region</Label>
          <Input
            id="region"
            placeholder="us-east-1"
            {...register('credentials.region')}
          />
        </div>
      </div>
    );
  }

  if (provider === 'gcp') {
    return (
      <div className="space-y-4">
        <div className="flex items-center space-x-4">
          <Button
            type="button"
            variant={gcpInputMode === 'json' ? 'default' : 'outline'}
            onClick={() => onGcpInputModeChange?.('json')}
          >
            JSON Input
          </Button>
          <Button
            type="button"
            variant={gcpInputMode === 'file' ? 'default' : 'outline'}
            onClick={() => onGcpInputModeChange?.('file')}
          >
            File Upload
          </Button>
        </div>
        
        {gcpInputMode === 'json' ? (
          <>
            <div className="text-sm font-medium">GCP Service Account JSON Fields:</div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="type">Type *</Label>
                <Input
                  id="type"
                  placeholder="service_account"
                  defaultValue="service_account"
                  {...register('credentials.type')}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="project_id">Project ID *</Label>
                <Input
                  id="project_id"
                  placeholder="my-project-123"
                  {...register('credentials.project_id')}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="private_key_id">Private Key ID *</Label>
                <Input
                  id="private_key_id"
                  placeholder="Enter private key ID"
                  {...register('credentials.private_key_id')}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="private_key">Private Key *</Label>
                <Input
                  id="private_key"
                  type="password"
                  placeholder="-----BEGIN PRIVATE KEY-----"
                  {...register('credentials.private_key')}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="client_email">Client Email *</Label>
                <Input
                  id="client_email"
                  type="email"
                  placeholder="service-account@project.iam.gserviceaccount.com"
                  {...register('credentials.client_email')}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="client_id">Client ID *</Label>
                <Input
                  id="client_id"
                  placeholder="Enter client ID"
                  {...register('credentials.client_id')}
                />
              </div>
              <div className="space-y-2 col-span-2">
                <Label htmlFor="auth_uri">Auth URI</Label>
                <Input
                  id="auth_uri"
                  defaultValue="https://accounts.google.com/o/oauth2/auth"
                  {...register('credentials.auth_uri')}
                />
              </div>
              <div className="space-y-2 col-span-2">
                <Label htmlFor="token_uri">Token URI</Label>
                <Input
                  id="token_uri"
                  defaultValue="https://oauth2.googleapis.com/token"
                  {...register('credentials.token_uri')}
                />
              </div>
              <div className="space-y-2 col-span-2">
                <Label htmlFor="auth_provider_x509_cert_url">Auth Provider X509 Cert URL</Label>
                <Input
                  id="auth_provider_x509_cert_url"
                  defaultValue="https://www.googleapis.com/oauth2/v1/certs"
                  {...register('credentials.auth_provider_x509_cert_url')}
                />
              </div>
              <div className="space-y-2 col-span-2">
                <Label htmlFor="client_x509_cert_url">Client X509 Cert URL</Label>
                <Input
                  id="client_x509_cert_url"
                  placeholder="Enter client x509 cert URL"
                  {...register('credentials.client_x509_cert_url')}
                />
              </div>
              <div className="space-y-2 col-span-2">
                <Label htmlFor="universe_domain">Universe Domain</Label>
                <Input
                  id="universe_domain"
                  defaultValue="googleapis.com"
                  {...register('credentials.universe_domain')}
                />
              </div>
            </div>
          </>
        ) : (
          <div className="space-y-2">
            <Label htmlFor="gcp_file">Service Account JSON File *</Label>
            <Input
              id="gcp_file"
              type="file"
              accept=".json"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) {
                  setValue('credentials._file', file as any); // eslint-disable-line @typescript-eslint/no-explicit-any
                }
              }}
            />
            <p className="text-sm text-gray-500">
              Upload your GCP service account JSON file
            </p>
          </div>
        )}
      </div>
    );
  }

  if (provider === 'azure') {
    return (
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="subscription_id">Subscription ID</Label>
          <Input
            id="subscription_id"
            placeholder="12345678-1234-1234-1234-123456789012"
            {...register('credentials.subscription_id')}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="client_id">Client ID</Label>
          <Input
            id="client_id"
            placeholder="12345678-1234-1234-1234-123456789012"
            {...register('credentials.client_id')}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="client_secret">Client Secret</Label>
          <Input
            id="client_secret"
            type="password"
            placeholder="Enter client secret"
            {...register('credentials.client_secret')}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="tenant_id">Tenant ID</Label>
          <Input
            id="tenant_id"
            placeholder="12345678-1234-1234-1234-123456789012"
            {...register('credentials.tenant_id')}
          />
        </div>
      </div>
    );
  }

  return null;
}

