import * as z from 'zod';

// Kubernetes validations
export const createClusterSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  version: z.string().min(1, 'Version is required'),
  region: z.string().min(1, 'Region is required'),
  zone: z.string().optional(),
  subnet_ids: z.array(z.string()).min(1, 'At least one subnet is required'),
  vpc_id: z.string().optional(),
  role_arn: z.string().optional(),
  tags: z.record(z.string(), z.string()).optional(),
  access_config: z.object({
    authentication_mode: z.string().optional(),
    bootstrap_cluster_creator_admin_permissions: z.boolean().optional(),
  }).optional(),
});

export const createNodePoolSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required'),
  cluster_name: z.string().min(1, 'Cluster name is required'),
  version: z.string().optional(),
  region: z.string().min(1, 'Region is required'),
  zone: z.string().optional(),
  instance_type: z.string().min(1, 'Instance type is required'),
  disk_size_gb: z.number().min(10).optional(),
  disk_type: z.string().optional(),
  min_nodes: z.number().min(0),
  max_nodes: z.number().min(1),
  node_count: z.number().min(0),
  auto_scaling: z.boolean().optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

export const createNodeGroupSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required'),
  cluster_name: z.string().min(1, 'Cluster name is required'),
  instance_type: z.string().min(1, 'Instance type is required'),
  disk_size_gb: z.number().min(10).optional(),
  min_size: z.number().min(0),
  max_size: z.number().min(1),
  desired_size: z.number().min(0),
  region: z.string().min(1, 'Region is required'),
  tags: z.record(z.string(), z.string()).optional(),
});

// Network validations
export const createVPCSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters'),
  description: z.string().max(500).optional(),
  cidr_block: z.string().regex(/^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/, 'Invalid CIDR format').optional(),
  region: z.string().optional(),
  project_id: z.string().optional(),
  auto_create_subnets: z.boolean().optional(),
  routing_mode: z.string().optional(),
  mtu: z.number().min(1280).max(8896).optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

export const createSubnetSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters'),
  vpc_id: z.string().min(1, 'VPC ID is required'),
  cidr_block: z.string().regex(/^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/, 'Invalid CIDR format'),
  availability_zone: z.string().min(1, 'Availability zone is required'),
  region: z.string().min(1, 'Region is required'),
  description: z.string().max(500).optional(),
  private_ip_google_access: z.boolean().optional(),
  flow_logs: z.boolean().optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

export const createSecurityGroupSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters'),
  description: z.string().min(1, 'Description is required').max(255, 'Description must be less than 255 characters'),
  vpc_id: z.string().min(1, 'VPC ID is required'),
  region: z.string().min(1, 'Region is required'),
  project_id: z.string().optional(),
  direction: z.enum(['INGRESS', 'EGRESS']).optional(),
  priority: z.number().min(0).max(65535).optional(),
  action: z.enum(['ALLOW', 'DENY']).optional(),
  protocol: z.string().optional(),
  ports: z.array(z.string()).optional(),
  source_ranges: z.array(z.string()).optional(),
  target_tags: z.array(z.string()).optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

// Common validations
export const credentialSchema = z.object({
  provider: z.enum(['aws', 'gcp', 'azure', 'ncp']),
  credentials: z.record(z.string(), z.string()),
});

export const workspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  description: z.string().max(1000).optional(),
});

export const regionSchema = z.string().min(1, 'Region is required');

export const uuidSchema = z.string().uuid('Invalid UUID format');

