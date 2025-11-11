/**
 * Base Repository
 * 모든 Repository의 공통 기능을 제공하는 베이스 클래스
 * BaseService를 상속하여 API 호출 기능 제공
 */

import { BaseService } from '@/lib/api';

export abstract class BaseRepository extends BaseService {
  // BaseService의 모든 기능을 상속받아 사용
  // 하위 클래스에서 인터페이스에 맞는 delete 메서드를 구현해야 함
}

