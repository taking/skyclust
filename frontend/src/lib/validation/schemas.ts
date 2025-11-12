import * as z from 'zod';
import { VALIDATION } from '../constants';

/**
 * Create validation schemas with translation support
 * 번역 함수를 받아서 validation schema를 생성하는 함수들
 */

export type TranslationFunction = (key: string, params?: Record<string, string | number>) => string;

/**
 * Create validation schemas with translation
 * 
 * 번역 함수를 받아서 모든 validation schema를 생성하는 팩토리 함수입니다.
 * Zod를 사용하여 타입 안전한 validation을 제공하며, 에러 메시지는 번역 함수를 통해 다국어를 지원합니다.
 * 
 * @param t - 번역 함수 (i18n의 t 함수)
 * @returns 모든 validation schema를 포함한 객체
 * 
 * @example
 * ```tsx
 * const { t } = useTranslation();
 * const schemas = createValidationSchemas(t);
 * 
 * // 사용 예시
 * const form = useForm({
 *   resolver: zodResolver(schemas.createClusterSchema),
 * });
 * ```
 */
export function createValidationSchemas(t: TranslationFunction) {
  // ===== Kubernetes 관련 Validation Schemas =====
  const createClusterSchema = z.object({
    credential_id: z.string().uuid(t('form.validation.invalidCredentialId')),
    name: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.nameRequired')).max(VALIDATION.STRING.MAX_NAME_LENGTH, t('form.validation.nameMaxLength', { max: String(VALIDATION.STRING.MAX_NAME_LENGTH) })),
    version: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.versionRequired')),
    region: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.regionRequired')),
    zone: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.required', { field: 'Availability Zone' })),
    subnet_ids: z.array(z.string()).min(VALIDATION.ARRAY.MIN_SUBNET_COUNT, t('form.validation.atLeastOneSubnet')),
    vpc_id: z.string().optional(),
    role_arn: z.string().optional(),
    tags: z.record(z.string(), z.string()).optional(),
    access_config: z.object({
      authentication_mode: z.string().optional(),
      bootstrap_cluster_creator_admin_permissions: z.boolean().optional(),
    }).optional(),
  });

  const updateClusterSchema = z.object({
    name: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.nameRequired')).max(VALIDATION.STRING.MAX_NAME_LENGTH, t('form.validation.nameMaxLength', { max: String(VALIDATION.STRING.MAX_NAME_LENGTH) })).optional(),
    version: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.versionRequired')).optional(),
    tags: z.record(z.string(), z.string()).optional(),
  });

  const createNodePoolSchema = z.object({
    credential_id: z.string().uuid(t('form.validation.invalidCredentialId')),
    name: z.string().min(1, t('form.validation.nameRequired')),
    cluster_name: z.string().min(1, t('form.validation.clusterNameRequired')),
    version: z.string().optional(),
    region: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.regionRequired')),
    zone: z.string().optional(),
    instance_type: z.string().min(1, t('form.validation.instanceTypeRequired')),
    disk_size_gb: z.number().min(10).optional(),
    disk_type: z.string().optional(),
    min_nodes: z.number().min(0),
    max_nodes: z.number().min(1),
    node_count: z.number().min(0),
    auto_scaling: z.boolean().optional(),
    tags: z.record(z.string(), z.string()).optional(),
  });

  const updateNodePoolSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).optional(),
    version: z.string().optional(),
    min_nodes: z.number().min(0).optional(),
    max_nodes: z.number().min(1).optional(),
    node_count: z.number().min(0).optional(),
    auto_scaling: z.boolean().optional(),
    tags: z.record(z.string(), z.string()).optional(),
  });

  const createNodeGroupSchema = z.object({
    credential_id: z.string().uuid(t('form.validation.invalidCredentialId')),
    name: z.string().min(1, t('form.validation.nameRequired')),
    cluster_name: z.string().min(1, t('form.validation.clusterNameRequired')),
    instance_type: z.string().min(1, t('form.validation.instanceTypeRequired')),
    disk_size_gb: z.number().min(10).optional(),
    min_size: z.number().min(0),
    max_size: z.number().min(1),
    desired_size: z.number().min(0),
    region: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.regionRequired')),
    tags: z.record(z.string(), z.string()).optional(),
  });

  // Network validations
  const createVPCSchema = z.object({
    credential_id: z.string().uuid(t('form.validation.invalidCredentialId')),
    name: z.string().min(1, t('form.validation.nameRequired')).max(255, t('form.validation.nameMaxLength', { max: '255' })),
    description: z.string().max(500).optional(),
    // cidr_block은 AWS/Azure에만 필요, GCP는 optional (auto-mode VPC)
    cidr_block: z.string().regex(/^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/, t('form.validation.invalidCidrFormat')).optional(),
    region: z.string().optional(), // GCP는 region이 optional일 수 있음
    project_id: z.string().optional(), // GCP에만 필요
    auto_create_subnets: z.boolean().optional(), // GCP에만 필요
    routing_mode: z.string().optional(), // GCP에만 필요
    mtu: z.number().min(1280).max(8896).optional(), // GCP에만 필요
    // Azure specific fields
    location: z.string().optional(), // Azure uses 'location' instead of 'region'
    resource_group: z.string().optional(), // Azure에만 필요
    address_space: z.array(z.string()).optional(), // Azure에만 필요
    tags: z.record(z.string(), z.string()).optional(),
  });

  const updateVPCSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).max(255, t('form.validation.nameMaxLength', { max: '255' })).optional(),
    description: z.string().max(500).optional(),
    tags: z.record(z.string(), z.string()).optional(),
  });

  const createSubnetSchema = z.object({
    credential_id: z.string().uuid(t('form.validation.invalidCredentialId')),
    name: z.string().min(1, t('form.validation.nameRequired')).max(255, t('form.validation.nameMaxLength', { max: '255' })),
    vpc_id: z.string().min(1, t('form.validation.vpcIdRequired')),
    cidr_block: z.string().regex(/^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/, t('form.validation.invalidCidrFormat')).min(1, t('form.validation.cidrBlockRequired')),
    availability_zone: z.string().optional(), // AWS/Azure에만 필요, GCP는 zone 사용
    region: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.regionRequired')),
    description: z.string().max(500).optional(),
    // GCP specific fields
    project_id: z.string().optional(), // GCP에만 필요
    zone: z.string().optional(), // GCP uses 'zone' instead of 'availability_zone'
    private_ip_google_access: z.boolean().optional(), // GCP에만 필요
    flow_logs: z.boolean().optional(), // GCP에만 필요
    tags: z.record(z.string(), z.string()).optional(),
  });

  const createResourceGroupSchema = z.object({
    credential_id: z.string().uuid(t('form.validation.invalidCredentialId')),
    name: z.string()
      .min(1, t('form.validation.nameRequired'))
      .max(90, t('form.validation.nameMaxLength', { max: '90' }))
      .regex(/^[a-zA-Z0-9._()-]+$/, 'Resource group name can only contain alphanumeric characters, periods, underscores, hyphens, and parentheses'),
    location: z.string().min(1, 'Location is required'),
    tags: z.record(z.string(), z.string()).optional(),
  });

  const updateSubnetSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).max(255, t('form.validation.nameMaxLength', { max: '255' })).optional(),
    description: z.string().max(500).optional(),
    tags: z.record(z.string(), z.string()).optional(),
  });

  const createSecurityGroupSchema = z.object({
    credential_id: z.string().uuid(t('form.validation.invalidCredentialId')),
    name: z.string().min(1, t('form.validation.nameRequired')).max(255, t('form.validation.nameMaxLength', { max: '255' })),
    description: z.string().min(1, t('form.validation.descriptionRequired')).max(255, t('form.validation.descriptionMaxLength', { max: '255' })),
    vpc_id: z.string().min(1, t('form.validation.vpcIdRequired')),
    region: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.regionRequired')),
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

  const updateSecurityGroupSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).max(255, t('form.validation.nameMaxLength', { max: '255' })).optional(),
    description: z.string().min(1, t('form.validation.descriptionRequired')).max(255, t('form.validation.descriptionMaxLength', { max: '255' })).optional(),
    tags: z.record(z.string(), z.string()).optional(),
  });

  // VM validations
  const createVMSchema = z.object({
    credential_id: z.string().uuid(t('form.validation.invalidCredentialId')).optional(),
    name: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.nameRequired')).max(VALIDATION.STRING.MAX_NAME_LENGTH, t('form.validation.nameMaxLength', { max: String(VALIDATION.STRING.MAX_NAME_LENGTH) })),
    provider: z.enum(['aws', 'gcp', 'azure', 'ncp']),
    instance_type: z.string().min(1, t('form.validation.instanceTypeRequired')),
    region: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.regionRequired')),
    image_id: z.string().optional(),
    workspace_id: z.string().optional(),
    metadata: z.record(z.string(), z.unknown()).optional(),
  });

  const updateVMSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).max(100, t('form.validation.nameMaxLength', { max: '100' })).optional(),
    type: z.string().optional(),
    metadata: z.record(z.string(), z.unknown()).optional(),
  });

  // Credential validations
  const createCredentialSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).max(100, t('form.validation.nameMaxLength', { max: '100' })).optional(),
    provider: z.string().min(1, t('form.validation.providerRequired')).refine((val) => ['aws', 'gcp', 'azure', 'ncp'].includes(val), {
      message: t('form.validation.providerMustBeOneOf'),
    }),
    credentials: z.record(z.string(), z.unknown()),
  });

  const updateCredentialSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).max(100, t('form.validation.nameMaxLength', { max: '100' })).optional(),
    credentials: z.record(z.string(), z.string()).optional(),
  });

  // Auth validations
  const loginSchema = z.object({
    email: z.string().email(t('form.validation.invalidEmail')),
    password: z.string().min(6, t('form.validation.passwordMinLength', { min: '6' })),
  });

  const registerSchema = z.object({
    name: z.string().min(3, t('form.validation.nameMinLength', { min: '3' })).max(50, t('form.validation.nameMaxLength', { max: '50' })),
    email: z.string().email(t('form.validation.invalidEmail')),
    password: z.string().min(8, t('form.validation.passwordMinLength', { min: '8' })),
  });

  const profileSchema = z.object({
    username: z.string().min(3, t('form.validation.usernameMinLength', { min: '3' })),
    email: z.string().email(t('form.validation.invalidEmail')),
  });

  const passwordSchema = z.object({
    currentPassword: z.string().min(6, t('form.validation.currentPasswordRequired')),
    newPassword: z.string().min(8, t('form.validation.newPasswordMinLength', { min: '8' })),
    confirmPassword: z.string().min(8, t('form.validation.confirmPasswordRequired')),
  }).refine((data) => data.newPassword === data.confirmPassword, {
    message: t('form.validation.passwordsDontMatch'),
    path: ['confirmPassword'],
  });

  const notificationSchema = z.object({
    emailNotifications: z.boolean(),
    pushNotifications: z.boolean(),
    securityAlerts: z.boolean(),
    systemUpdates: z.boolean(),
  });

  // Common validations
  const credentialSchema = z.object({
    provider: z.enum(['aws', 'gcp', 'azure', 'ncp']),
    credentials: z.record(z.string(), z.string()),
  });

  const workspaceSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).max(255),
    description: z.string().max(1000).optional(),
  });

  const createWorkspaceSchema = z.object({
    name: z.string().min(VALIDATION.STRING.MIN_LENGTH, t('form.validation.nameRequired')).max(VALIDATION.STRING.MAX_NAME_LENGTH, t('form.validation.nameMaxLength', { max: String(VALIDATION.STRING.MAX_NAME_LENGTH) })),
    description: z.string().min(1, t('form.validation.descriptionRequired')).max(500, t('form.validation.descriptionMaxLength', { max: '500' })),
  });

  const updateWorkspaceSchema = z.object({
    name: z.string().min(1, t('form.validation.nameRequired')).max(100, t('form.validation.nameMaxLength', { max: '100' })).optional(),
    description: z.string().min(1, t('form.validation.descriptionRequired')).max(500, t('form.validation.descriptionMaxLength', { max: '500' })).optional(),
  });

  const addMemberSchema = z.object({
    email: z.string().email(t('form.validation.invalidEmail')),
    role: z.enum(['admin', 'member']),
  });

  const regionSchema = z.string().min(1, t('form.validation.regionRequired'));

  const uuidSchema = z.string().uuid(t('form.validation.invalidUuid'));

  return {
    loginSchema,
    registerSchema,
    profileSchema,
    passwordSchema,
    notificationSchema,
    createWorkspaceSchema,
    updateWorkspaceSchema,
    addMemberSchema,
    createClusterSchema,
    updateClusterSchema,
    createNodePoolSchema,
    updateNodePoolSchema,
    createNodeGroupSchema,
    createVPCSchema,
    updateVPCSchema,
    createSubnetSchema,
    updateSubnetSchema,
    createSecurityGroupSchema,
    updateSecurityGroupSchema,
    createResourceGroupSchema,
    createVMSchema,
    updateVMSchema,
    createCredentialSchema,
    updateCredentialSchema,
    credentialSchema,
    workspaceSchema,
    regionSchema,
    uuidSchema,
  };
}

// Default schemas for backward compatibility (English messages)
// These will be replaced by translated schemas in components
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

export const updateClusterSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters').optional(),
  version: z.string().min(1, 'Version is required').optional(),
  tags: z.record(z.string(), z.string()).optional(),
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

export const updateNodePoolSchema = z.object({
  name: z.string().min(1, 'Name is required').optional(),
  version: z.string().optional(),
  min_nodes: z.number().min(0).optional(),
  max_nodes: z.number().min(1).optional(),
  node_count: z.number().min(0).optional(),
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

export const updateVPCSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters').optional(),
  description: z.string().max(500).optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

export const createSubnetSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters'),
  vpc_id: z.string().min(1, 'VPC ID is required'),
  cidr_block: z.string().regex(/^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/, 'Invalid CIDR format').min(1, 'CIDR block is required'),
  availability_zone: z.string().min(1, 'Availability zone is required'),
  region: z.string().min(1, 'Region is required'),
  description: z.string().max(500).optional(),
  private_ip_google_access: z.boolean().optional(),
  flow_logs: z.boolean().optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

export const createResourceGroupSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string()
    .min(1, 'Name is required')
    .max(90, 'Name must be less than 90 characters')
    .regex(/^[a-zA-Z0-9._()-]+$/, 'Resource group name can only contain alphanumeric characters, periods, underscores, hyphens, and parentheses'),
  location: z.string().min(1, 'Location is required'),
  tags: z.record(z.string(), z.string()).optional(),
});

export const updateSubnetSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters').optional(),
  description: z.string().max(500).optional(),
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

export const updateSecurityGroupSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be less than 255 characters').optional(),
  description: z.string().min(1, 'Description is required').max(255, 'Description must be less than 255 characters').optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

// VM validations
export const createVMSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID').optional(),
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  provider: z.enum(['aws', 'gcp', 'azure', 'ncp']),
  instance_type: z.string().min(1, 'Instance type is required'),
  region: z.string().min(1, 'Region is required'),
  image_id: z.string().optional(),
  workspace_id: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
});

export const updateVMSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters').optional(),
  type: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
});

// Credential validations
export const createCredentialSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters').optional(),
  provider: z.string().min(1, 'Provider is required').refine((val) => ['aws', 'gcp', 'azure', 'ncp'].includes(val), {
    message: 'Provider must be one of: aws, gcp, azure, ncp',
  }),
  credentials: z.record(z.string(), z.unknown()),
});

export const updateCredentialSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters').optional(),
  credentials: z.record(z.string(), z.string()).optional(),
});

// Auth validations (backward compatibility - use createValidationSchemas for translated versions)
export const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

export const registerSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

export const profileSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  email: z.string().email('Invalid email address'),
});

export const passwordSchema = z.object({
  currentPassword: z.string().min(6, 'Current password is required'),
  newPassword: z.string().min(8, 'New password must be at least 8 characters'),
  confirmPassword: z.string().min(8, 'Confirm password is required'),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
});

export const notificationSchema = z.object({
  emailNotifications: z.boolean(),
  pushNotifications: z.boolean(),
  securityAlerts: z.boolean(),
  systemUpdates: z.boolean(),
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

export const createWorkspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters'),
});

export const updateWorkspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters').optional(),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters').optional(),
});

export const addMemberSchema = z.object({
  email: z.string().email('Invalid email address'),
  role: z.enum(['admin', 'member']),
});

export const regionSchema = z.string().min(1, 'Region is required');

export const uuidSchema = z.string().uuid('Invalid UUID format');

