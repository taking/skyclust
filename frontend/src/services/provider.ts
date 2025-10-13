import api from '@/lib/api';
import { ApiResponse, Provider, Instance, Region } from '@/lib/types';

export const providerService = {
  // Get all providers
  getProviders: async (): Promise<Provider[]> => {
    const response = await api.get<ApiResponse<{ providers: Provider[] }>>('/api/v1/providers');
    return response.data.data!.providers;
  },

  // Get provider by name
  getProvider: async (name: string): Promise<Provider> => {
    const response = await api.get<ApiResponse<Provider>>(`/api/v1/providers/${name}`);
    return response.data.data!;
  },

  // Get instances by provider
  getInstances: async (provider: string, region?: string): Promise<Instance[]> => {
    const params = region ? `?region=${region}` : '';
    const response = await api.get<ApiResponse<Instance[]>>(`/api/v1/providers/${provider}/instances${params}`);
    return response.data.data!;
  },

  // Get instance by ID
  getInstance: async (provider: string, instanceId: string): Promise<Instance> => {
    const response = await api.get<ApiResponse<Instance>>(`/api/v1/providers/${provider}/instances/${instanceId}`);
    return response.data.data!;
  },

  // Get regions by provider
  getRegions: async (provider: string): Promise<Region[]> => {
    const response = await api.get<ApiResponse<Region[]>>(`/api/v1/providers/${provider}/regions`);
    return response.data.data!;
  },

  // Get cost estimates
  getCostEstimates: async (provider: string): Promise<unknown[]> => {
    const response = await api.get<ApiResponse<unknown[]>>(`/api/v1/providers/${provider}/cost-estimates`);
    return response.data.data!;
  },

  // Create cost estimate
  createCostEstimate: async (provider: string, data: {
    instance_type: string;
    region: string;
    duration_hours: number;
  }): Promise<unknown> => {
    const response = await api.post<ApiResponse<unknown>>(`/api/v1/providers/${provider}/cost-estimates`, data);
    return response.data.data!;
  },
};
