/**
 * Regions Module
 * 모든 리전 관련 기능을 중앙화하여 export
 */

export {
  AWS_REGIONS,
  GCP_REGIONS,
  AZURE_REGIONS,
  getAllRegions,
  getRegionsByProvider,
  supportsRegionSelection,
  getDefaultRegionForProvider,
} from './list';
export type { RegionOption } from './list';

