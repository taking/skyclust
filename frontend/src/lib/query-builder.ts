/**
 * Query Parameter Builder
 * URL 쿼리 파라미터를 안전하고 일관되게 생성하는 유틸리티
 */

/**
 * Query parameter 값 타입
 */
type QueryValue = string | number | boolean | null | undefined;

/**
 * Query parameter 객체 타입
 */
export type QueryParams = Record<string, QueryValue | QueryValue[]>;

/**
 * Query parameter를 URLSearchParams로 변환
 * undefined, null 값은 자동으로 제외됩니다.
 * 
 * @param params - Query parameter 객체
 * @returns URLSearchParams 인스턴스
 * 
 * @example
 * buildQueryParams({ name: 'test', age: 25, active: true })
 * // URLSearchParams: name=test&age=25&active=true
 * 
 * buildQueryParams({ name: 'test', age: undefined, active: null })
 * // URLSearchParams: name=test
 * 
 * buildQueryParams({ tags: ['tag1', 'tag2'] })
 * // URLSearchParams: tags=tag1&tags=tag2
 */
export function buildQueryParams(params: QueryParams): URLSearchParams {
  // 1. 새로운 URLSearchParams 인스턴스 생성
  const searchParams = new URLSearchParams();

  // 2. 파라미터 객체의 모든 키-값 쌍 순회
  for (const [key, value] of Object.entries(params)) {
    // 3. undefined, null 값은 제외 (유효한 값만 추가)
    if (value === undefined || value === null) {
      continue;
    }

    // 4. 배열인 경우 각 값에 대해 append (다중 값 지원)
    if (Array.isArray(value)) {
      for (const item of value) {
        // 배열 내부의 undefined, null 값도 제외
        if (item !== undefined && item !== null) {
          searchParams.append(key, String(item));
        }
      }
    } else {
      // 5. 단일 값인 경우 문자열로 변환하여 추가
      searchParams.append(key, String(value));
    }
  }

  // 6. 완성된 URLSearchParams 반환
  return searchParams;
}

/**
 * Query parameter를 문자열로 변환
 * 
 * @param params - Query parameter 객체
 * @returns Query string (예: "name=test&age=25")
 * 
 * @example
 * buildQueryString({ name: 'test', age: 25 })
 * // "name=test&age=25"
 */
export function buildQueryString(params: QueryParams): string {
  // 1. QueryParams를 URLSearchParams로 변환
  const searchParams = buildQueryParams(params);
  
  // 2. URLSearchParams를 문자열로 변환 (예: "name=test&age=25")
  return searchParams.toString();
}

/**
 * Query parameter를 포함한 엔드포인트 URL 생성
 * 
 * @param basePath - 기본 경로 (예: "vms", "/workspaces/123")
 * @param params - Query parameter 객체
 * @returns Query string이 포함된 엔드포인트 (예: "vms?workspace_id=123&status=running")
 * 
 * @example
 * buildEndpointWithQuery('vms', { workspace_id: '123', status: 'running' })
 * // "vms?workspace_id=123&status=running"
 * 
 * buildEndpointWithQuery('workspaces/123/members', { role: 'admin' })
 * // "workspaces/123/members?role=admin"
 */
export function buildEndpointWithQuery(basePath: string, params?: QueryParams): string {
  // 1. 파라미터가 없거나 빈 객체인 경우 기본 경로만 반환
  if (!params || Object.keys(params).length === 0) {
    return basePath;
  }

  // 2. Query parameter를 문자열로 변환
  const queryString = buildQueryString(params);
  
  // 3. Query string이 비어있으면 기본 경로만 반환
  if (!queryString) {
    return basePath;
  }

  // 4. 기본 경로와 query string을 결합하여 반환
  return `${basePath}?${queryString}`;
}

/**
 * 기존 URLSearchParams에 추가 파라미터를 병합
 * 
 * @param existing - 기존 URLSearchParams
 * @param additional - 추가할 파라미터 객체
 * @returns 병합된 URLSearchParams
 * 
 * @example
 * const existing = new URLSearchParams('name=test');
 * mergeQueryParams(existing, { age: 25, city: 'Seoul' })
 * // URLSearchParams: name=test&age=25&city=Seoul
 */
export function mergeQueryParams(
  existing: URLSearchParams,
  additional: QueryParams
): URLSearchParams {
  // 1. 기존 URLSearchParams를 복사하여 새 인스턴스 생성
  const merged = new URLSearchParams(existing);
  
  // 2. 추가할 파라미터를 URLSearchParams로 변환
  const additionalParams = buildQueryParams(additional);

  // 3. 추가 파라미터의 모든 키-값 쌍을 기존 파라미터에 추가
  // append를 사용하므로 같은 키가 있으면 값이 추가됨 (덮어쓰지 않음)
  for (const [key, value] of additionalParams.entries()) {
    merged.append(key, value);
  }

  // 4. 병합된 URLSearchParams 반환
  return merged;
}

/**
 * 여러 QueryParams 객체를 병합
 * 나중에 오는 객체의 값이 우선순위를 가집니다.
 * 
 * @param params - 병합할 QueryParams 객체 배열
 * @returns 병합된 QueryParams
 * 
 * @example
 * mergeMultipleQueryParams(
 *   { name: 'test', age: 25 },
 *   { age: 30, city: 'Seoul' }
 * )
 * // { name: 'test', age: 30, city: 'Seoul' }
 */
export function mergeMultipleQueryParams(...params: QueryParams[]): QueryParams {
  // 1. Object.assign을 사용하여 여러 객체를 병합
  // 나중에 오는 객체의 값이 우선순위를 가짐 (덮어쓰기)
  // 첫 번째 인자로 빈 객체를 전달하여 원본 객체를 변경하지 않음
  return Object.assign({}, ...params);
}

