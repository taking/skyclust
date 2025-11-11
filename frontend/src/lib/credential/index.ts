/**
 * Credential Module
 * 모든 자격증명 관련 기능을 중앙화하여 export
 */

export {
  trackCredentialUsage,
  setDefaultCredential,
  getDefaultCredential,
  getRecentlyUsedCredentials,
  getRecommendedCredential,
  clearPreferences,
} from './preference';
export type { CredentialUsage, CredentialPreference } from './preference';

