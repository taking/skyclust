# SkyClust Frontend

SkyClust는 멀티 클라우드 관리 플랫폼의 프론트엔드 애플리케이션입니다. AWS, GCP, Azure, NCP 등 여러 클라우드 제공업체의 리소스를 통합 관리할 수 있는 웹 인터페이스를 제공합니다.

## 기술 스택

### 핵심 프레임워크
- **Next.js** 15.5.4 (App Router, Turbopack)
- **React** 19.1.0
- **TypeScript** 5.x
- **Tailwind CSS** 4.x

### 상태 관리 & 데이터 페칭
- **Zustand** 5.0.8 - 클라이언트 상태 관리
- **TanStack Query** 5.90.2 - 서버 상태 관리 및 캐싱
- **React Hook Form** 7.63.0 - 폼 관리
- **Zod** 4.1.11 - 스키마 검증

### UI 컴포넌트
- **Radix UI** - 접근성 우선 UI 프리미티브 (직접 사용)
  - `@radix-ui/react-dialog`, `@radix-ui/react-select` 등
  - 접근성(A11y)을 고려한 컴포넌트 라이브러리
- **Lucide React** - 아이콘
- **Recharts** 3.2.1 - 차트 및 데이터 시각화
- **react-hot-toast** 2.6.0 - 토스트 알림
- **class-variance-authority** - 컴포넌트 variant 관리
- **tailwind-merge** - Tailwind CSS 클래스 병합 유틸리티

### 기타 주요 라이브러리
- **next-intl** 4.4.0 - 국제화 (한국어/영어)
- **next-themes** 0.4.6 - 다크 모드 지원
- **@sentry/nextjs** 10 - 에러 모니터링 및 성능 추적
- **axios** 1.12.2 - HTTP 클라이언트
- **fuse.js** 7.1.0 - 퍼지 검색
- **@tanstack/react-virtual** 3.13.12 - 가상화 리스트
- **@dnd-kit** - 드래그 앤 드롭

## 프로젝트 구조

```
frontend/
├── src/
│   ├── app/                    # Next.js App Router 페이지
│   │   ├── compute/            # VM, 이미지, 스냅샷
│   │   ├── kubernetes/         # Kubernetes 클러스터 관리
│   │   ├── networks/           # VPC, 서브넷, 보안 그룹
│   │   ├── credentials/       # 클라우드 자격 증명
│   │   ├── workspaces/         # 작업 공간 관리
│   │   ├── dashboard/         # 대시보드
│   │   ├── cost-analysis/     # 비용 분석
│   │   └── ...
│   │
│   ├── components/            # 재사용 가능한 컴포넌트
│   │   ├── ui/               # 기본 UI 컴포넌트 (Radix UI 기반)
│   │   ├── common/           # 공통 컴포넌트
│   │   ├── layout/           # 레이아웃 컴포넌트
│   │   ├── features/         # 기능별 컴포넌트
│   │   ├── widgets/          # 대시보드 위젯
│   │   └── providers/        # Context Provider
│   │
│   ├── features/             # 기능별 모듈 (Feature-based)
│   │   ├── vms/             # VM 관리
│   │   ├── kubernetes/       # Kubernetes 관리
│   │   ├── networks/         # 네트워크 관리
│   │   ├── credentials/      # 자격 증명 관리
│   │   └── workspaces/       # 작업 공간 관리
│   │
│   ├── domain/               # 도메인 로직 (Clean Architecture)
│   │   └── use-cases/       # 비즈니스 로직
│   │
│   ├── infrastructure/       # 인프라스트럭처 계층
│   │   └── repositories/    # 데이터 접근 계층
│   │
│   ├── services/            # API 서비스 레이어
│   │   ├── auth.ts
│   │   ├── vm.ts
│   │   ├── kubernetes.ts
│   │   ├── network.ts
│   │   └── ...
│   │
│   ├── hooks/               # 커스텀 React Hooks
│   │   ├── use-form-with-validation.ts
│   │   ├── use-pagination.ts
│   │   ├── use-sse-events.ts
│   │   └── ...
│   │
│   ├── store/               # Zustand 스토어
│   │   ├── auth.ts
│   │   ├── workspace.ts
│   │   ├── credential-context.ts
│   │   └── ...
│   │
│   ├── lib/                 # 유틸리티 및 헬퍼
│   │   ├── types/          # TypeScript 타입 정의
│   │   ├── validations.ts  # Zod 스키마
│   │   ├── api.ts         # API 클라이언트
│   │   └── ...
│   │
│   └── i18n/               # 국제화 설정
│       ├── messages/      # 번역 파일 (ko.json, en.json)
│       └── config.ts
│
├── public/                 # 정적 파일
├── next.config.js         # Next.js 설정
├── tailwind.config.js     # Tailwind CSS 설정
├── tsconfig.json          # TypeScript 설정
└── package.json
```

## 아키텍처

### Clean Architecture 패턴
프로젝트는 Clean Architecture 원칙을 따릅니다:

- **Domain Layer** (`domain/`): 비즈니스 로직과 엔티티
- **Application Layer** (`features/`): 기능별 모듈 및 Use Cases
- **Infrastructure Layer** (`infrastructure/`): 데이터 접근 (Repository 패턴)
- **Presentation Layer** (`components/`, `app/`): UI 컴포넌트 및 페이지

### Feature-based 구조
기능별로 모듈화되어 있어 유지보수와 확장이 용이합니다:

```
features/
├── vms/
│   ├── components/         # VM 관련 컴포넌트
│   ├── hooks/             # VM 관련 hooks
│   └── types.ts           # VM 타입 정의
```

### Repository 패턴
데이터 접근 계층을 추상화하여 테스트와 유지보수를 용이하게 합니다:

```typescript
interface IVPCRepository {
  list(provider: string, credentialId: string, region?: string): Promise<VPC[]>;
  create(provider: string, data: CreateVPCForm): Promise<VPC>;
  // ...
}
```

## 시작하기

### 필수 요구사항
- Node.js 20.x 이상
- npm, yarn, pnpm, 또는 bun

### 설치

```bash
# 의존성 설치
npm install

# 또는
yarn install
pnpm install
bun install
```

### 환경 변수 설정

`.env.local` 파일을 생성하고 다음 변수를 설정하세요:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 개발 서버 실행

```bash
npm run dev

# 또는
yarn dev
pnpm dev
bun dev
```

브라우저에서 [http://localhost:3000](http://localhost:3000)을 열어 확인하세요.

### 프로덕션 빌드

```bash
npm run build
npm start
```

## 주요 기능

### 1. 멀티 클라우드 관리
- **VM 관리**: 여러 클라우드 제공업체의 가상 머신 통합 관리
- **Kubernetes**: 클러스터, 노드 풀, 노드 그룹 관리
- **네트워크**: VPC, 서브넷, 보안 그룹 관리
- **자격 증명**: 클라우드 제공업체 자격 증명 관리

### 2. 실시간 모니터링
- **SSE (Server-Sent Events)**: 실시간 리소스 상태 업데이트
- **대시보드**: 드래그 앤 드롭 가능한 위젯 기반 대시보드
- **알림**: 실시간 알림 시스템

### 3. 비용 분석
- 비용 추이 차트
- 예산 알림
- 비용 예측

### 4. 다국어 지원
- 한국어 (기본)
- 영어
- `next-intl` 기반 국제화

### 5. 접근성
- WCAG 2.1 준수
- 키보드 단축키 지원
- 스크린 리더 지원

### 6. 반응형 디자인
- 모바일, 태블릿, 데스크톱 지원
- 터치 제스처 지원

### 7. 오프라인 지원
- 오프라인 상태 감지
- 오프라인 작업 큐
- 네트워크 복구 시 자동 동기화

## 개발 가이드

### 코드 스타일

프로젝트는 다음 원칙을 따릅니다:
- **YAGNI** (You Aren't Gonna Need It)
- **KISS** (Keep It Simple, Stupid)
- **DRY** (Don't Repeat Yourself)
- **Clean Code** 원칙

### 폼 검증

Zod를 사용한 스키마 검증과 React Hook Form을 통합한 폼 관리:

```typescript
import { createValidationSchemas } from '@/lib/validations';
import { useTranslation } from '@/hooks/use-translation';

const { t } = useTranslation();
const schemas = createValidationSchemas(t);

// 검증된 스키마 사용
const form = useForm({
  resolver: zodResolver(schemas.createVMSchema),
});
```

### 상태 관리

#### Zustand (클라이언트 상태)
```typescript
import { useAuthStore } from '@/store/auth';

const { user, isAuthenticated } = useAuthStore();
```

#### TanStack Query (서버 상태)
```typescript
import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query-keys';

const { data, isLoading } = useQuery({
  queryKey: queryKeys.vms.all,
  queryFn: () => vmService.getAll(),
});
```

### 국제화

번역 키 사용:
```typescript
import { useTranslation } from '@/hooks/use-translation';

const { t } = useTranslation();
const message = t('common.home'); // "홈" 또는 "Home"
```

번역 파일 위치: `src/i18n/messages/`

### 커스텀 Hooks

프로젝트는 여러 재사용 가능한 커스텀 hooks를 제공합니다:

- `useFormWithValidation` - 폼 검증 통합
- `usePagination` - 페이지네이션
- `useSSEEvents` - 실시간 이벤트 구독
- `useErrorHandler` - 에러 처리
- `useTranslation` - 국제화
- `useKeyboardShortcuts` - 키보드 단축키

## 테스팅

```bash
# 린트 검사
npm run lint
```

## 에러 모니터링

프로젝트는 Sentry를 사용하여 에러 모니터링을 수행합니다:

- 클라이언트 사이드 에러 추적
- 성능 모니터링
- 소스맵 업로드

Sentry 설정은 `next.config.js`와 `sentry.client.config.ts`에 있습니다.

## 빌드 및 배포

### Docker 빌드

```bash
docker build -t skyclust-frontend .
docker run -p 3000:3000 skyclust-frontend
```

### Vercel 배포

프로젝트는 Vercel에 최적화되어 있습니다:

```bash
vercel deploy
```

## 주요 설정 파일

- `next.config.js` - Next.js 설정 (Sentry 포함)
- `tailwind.config.js` - Tailwind CSS 설정
- `tsconfig.json` - TypeScript 설정

## 주요 문서

- [Next.js 문서](https://nextjs.org/docs)
- [TanStack Query 문서](https://tanstack.com/query/latest)
- [React Hook Form 문서](https://react-hook-form.com/)
- [Zod 문서](https://zod.dev/)
- [next-intl 문서](https://next-intl-docs.vercel.app/)

## 기여하기

1. 이슈를 생성하거나 기존 이슈를 확인하세요
2. 기능 브랜치를 생성하세요 (`git checkout -b feature/amazing-feature`)
3. 변경사항을 커밋하세요 (`git commit -m 'feat: Add amazing feature'`)
4. 브랜치에 푸시하세요 (`git push origin feature/amazing-feature`)
5. Pull Request를 생성하세요

---

**SkyClust Frontend** - 멀티 클라우드 관리 플랫폼
