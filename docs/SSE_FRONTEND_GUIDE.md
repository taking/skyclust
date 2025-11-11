# Frontend SSE 활용 가이드

이 문서는 SkyClust 프론트엔드에서 SSE(Server-Sent Events)를 활용하여 실시간 데이터 업데이트를 구현하는 방법을 설명합니다.

## 📋 목차

1. [개요](#개요)
2. [SSE 서비스 사용 방법](#sse-서비스-사용-방법)
3. [이벤트 구독/구독 해제](#이벤트-구독구독-해제)
4. [React Query와 통합](#react-query와-통합)
5. [실시간 업데이트 사용](#실시간-업데이트-사용)
6. [세밀한 쿼리 무효화](#세밀한-쿼리-무효화)
7. [대시보드 SSE 활용](#대시보드-sse-활용)
8. [예제 코드](#예제-코드)
9. [모범 사례](#모범-사례)

---

## 개요

SkyClust 프론트엔드는 SSE를 통해 백엔드에서 발생하는 리소스 변경 이벤트를 실시간으로 수신하고, React Query 캐시를 자동으로 업데이트합니다.

### 아키텍처 흐름

```
Backend (NATS 이벤트 발행)
    ↓
SSE Handler (이벤트 브로드캐스트)
    ↓
Frontend SSE Service (EventSource)
    ↓
useSSEEvents Hook (이벤트 리스너 등록)
    ↓
React Query (캐시 업데이트/무효화)
    ↓
UI (자동 리렌더링)
```

---

## SSE 서비스 사용 방법

### 1. SSE 서비스 초기화

SSE 서비스는 전역적으로 초기화되어 있습니다 (`src/services/sse.ts`):

```typescript
import { sseService } from '@/services/sse';

// SSE 연결 확인
if (sseService.isConnected()) {
  console.log('SSE connected');
}
```

### 2. SSE 연결

SSE 연결은 `Layout` 컴포넌트에서 자동으로 관리됩니다. 수동으로 연결하려면:

```typescript
import { sseService } from '@/services/sse';

// 토큰으로 연결
const token = localStorage.getItem('token');
if (token) {
  sseService.connect(token, {
    onConnected: () => {
      console.log('SSE connected');
    },
    onError: (error) => {
      console.error('SSE error', error);
    },
  });
}
```

### 3. SSE 연결 해제

```typescript
sseService.disconnect();
```

---

## 이벤트 구독/구독 해제

### 1. 단일 이벤트 구독

```typescript
import { sseService } from '@/services/sse';

// 이벤트 구독
await sseService.subscribeToEvent(
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.created',
  {
    credential_ids: ['cred-123'],
    regions: ['ap-northeast-2'],
    providers: ['aws'],
  }
);
```

### 2. 여러 이벤트 일괄 구독

```typescript
const eventTypes = [
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.created',
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.updated',
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.deleted',
];

await sseService.subscribeToEvents(eventTypes, {
  credential_ids: ['cred-123'],
  regions: ['ap-northeast-2'],
});
```

### 3. 이벤트 구독 해제

```typescript
// 단일 이벤트 구독 해제
await sseService.unsubscribeFromEvent(
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.created'
);

// 여러 이벤트 일괄 구독 해제
await sseService.unsubscribeFromEvents(eventTypes);
```

### 4. 구독 동기화 (권장)

위젯이나 필터가 변경될 때 필요한 이벤트만 구독하도록 동기화:

```typescript
const requiredEvents = new Set([
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.created',
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.updated',
]);

await sseService.syncSubscriptions(requiredEvents, {
  credential_ids: ['cred-123'],
  regions: ['ap-northeast-2'],
});
```

---

## React Query와 통합

### 1. useSSEEvents Hook 사용

`useSSEEvents` 훅은 자동으로 SSE 이벤트를 수신하고 React Query 캐시를 무효화합니다:

```typescript
import { useSSEEvents } from '@/hooks/use-sse-events';

function MyComponent() {
  // SSE 이벤트 자동 처리
  useSSEEvents();

  // React Query 사용
  const { data } = useQuery({
    queryKey: queryKeys.kubernetesClusters.list(undefined, 'aws', 'cred-123', 'ap-northeast-2'),
    queryFn: () => kubernetesService.getClusters('aws', 'cred-123', 'ap-northeast-2'),
  });

  return <div>{/* ... */}</div>;
}
```

### 2. 수동 이벤트 리스너 등록

특정 이벤트에 대한 커스텀 핸들러를 등록하려면:

```typescript
import { sseService } from '@/services/sse';
import { useQueryClient } from '@tanstack/react-query';

function MyComponent() {
  const queryClient = useQueryClient();

  useEffect(() => {
    const callbacks = {
      onKubernetesClusterCreated: (data) => {
        console.log('Cluster created', data);
        // 커스텀 로직
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesClusters.all,
        });
      },
    };

    sseService.updateCallbacks(callbacks);

    return () => {
      // cleanup (필요시)
    };
  }, [queryClient]);
}
```

---

## 실시간 업데이트 사용

실시간 업데이트는 이벤트 데이터에 리소스 객체가 포함된 경우, React Query 캐시를 즉시 업데이트하여 UI 반응성을 향상시킵니다.

### 자동 실시간 업데이트

`useSSEEvents` 훅은 자동으로 실시간 업데이트를 시도합니다:

```typescript
// useSSEEvents 내부에서 자동 처리
onKubernetesClusterCreated: (data) => {
  try {
    // 실시간 업데이트 시도
    applyKubernetesClusterCreatedUpdate(queryClient, eventData);
  } catch (error) {
    // 실패 시 무효화로 fallback
    queryClient.invalidateQueries({ ... });
  }
}
```

### 수동 실시간 업데이트

수동으로 실시간 업데이트를 적용하려면:

```typescript
import {
  applyKubernetesClusterCreatedUpdate,
  applyKubernetesClusterUpdatedUpdate,
  applyKubernetesClusterDeletedUpdate,
} from '@/lib/sse/query-updates';
import { useQueryClient } from '@tanstack/react-query';

function MyComponent() {
  const queryClient = useQueryClient();

  useEffect(() => {
    const callbacks = {
      onKubernetesClusterCreated: (data) => {
        applyKubernetesClusterCreatedUpdate(queryClient, data);
      },
      onKubernetesClusterUpdated: (data) => {
        applyKubernetesClusterUpdatedUpdate(queryClient, data);
      },
      onKubernetesClusterDeleted: (data) => {
        applyKubernetesClusterDeletedUpdate(queryClient, data);
      },
    };

    sseService.updateCallbacks(callbacks);
  }, [queryClient]);
}
```

### 사용 가능한 실시간 업데이트 함수

- `applyVMCreatedUpdate`, `applyVMUpdatedUpdate`, `applyVMDeletedUpdate`
- `applyKubernetesClusterCreatedUpdate`, `applyKubernetesClusterUpdatedUpdate`, `applyKubernetesClusterDeletedUpdate`
- `applyVPCCreatedUpdate`, `applyVPCUpdatedUpdate`, `applyVPCDeletedUpdate`
- `applySubnetCreatedUpdate`, `applySubnetUpdatedUpdate`, `applySubnetDeletedUpdate`
- `applySecurityGroupCreatedUpdate`, `applySecurityGroupUpdatedUpdate`, `applySecurityGroupDeletedUpdate`

---

## 세밀한 쿼리 무효화

실시간 업데이트가 실패하거나 이벤트 데이터에 리소스 객체가 없는 경우, 세밀한 쿼리 무효화를 사용하여 필요한 쿼리만 무효화합니다.

### 자동 세밀한 무효화

`useSSEEvents` 훅은 자동으로 세밀한 무효화를 수행합니다:

```typescript
// useSSEEvents 내부에서 자동 처리
onKubernetesClusterCreated: (data) => {
  try {
    applyKubernetesClusterCreatedUpdate(queryClient, eventData);
  } catch (error) {
    // Fallback: 세밀한 무효화
    invalidateKubernetesClusterQueries(queryClient, eventData, 'created');
  }
}
```

### 수동 세밀한 무효화

수동으로 세밀한 무효화를 적용하려면:

```typescript
import {
  invalidateVMQueries,
  invalidateKubernetesClusterQueries,
  invalidateVPCQueries,
  invalidateSubnetQueries,
  invalidateSecurityGroupQueries,
} from '@/lib/sse/query-invalidation';
import { useQueryClient } from '@tanstack/react-query';

function MyComponent() {
  const queryClient = useQueryClient();

  useEffect(() => {
    const callbacks = {
      onKubernetesClusterCreated: (data) => {
        invalidateKubernetesClusterQueries(queryClient, data, 'created');
      },
      onKubernetesClusterUpdated: (data) => {
        invalidateKubernetesClusterQueries(queryClient, data, 'updated');
      },
      onKubernetesClusterDeleted: (data) => {
        invalidateKubernetesClusterQueries(queryClient, data, 'deleted');
      },
    };

    sseService.updateCallbacks(callbacks);
  }, [queryClient]);
}
```

### 세밀한 무효화의 장점

- **성능 향상**: 필요한 쿼리만 무효화하여 불필요한 리페치 방지
- **정확한 범위**: provider, credentialId, region 등으로 정확한 쿼리만 무효화
- **하위 리소스 무효화**: 상위 리소스 변경 시 관련 하위 리소스도 자동 무효화

---

## 대시보드 SSE 활용

대시보드는 `useDashboardSSE` 훅을 사용하여 위젯별로 필요한 이벤트만 동적으로 구독합니다.

### useDashboardSSE Hook 사용

```typescript
import { useDashboardSSE } from '@/hooks/use-dashboard-sse';
import { useCredentialContext } from '@/hooks/use-credential-context';

function DashboardPage() {
  const { widgets } = useDashboard();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();

  // 대시보드 SSE 동적 구독
  useDashboardSSE({
    widgets,
    credentialId: selectedCredentialId || undefined,
    region: selectedRegion || undefined,
    includeSummary: true, // 대시보드 요약 정보 포함
    enabled: widgets.length > 0,
  });

  return <div>{/* ... */}</div>;
}
```

### 위젯별 이벤트 매핑

`useDashboardSSE`는 위젯 타입에 따라 필요한 이벤트를 자동으로 계산합니다:

```typescript
// 위젯 타입별 이벤트 매핑 (자동 처리)
const WIDGET_EVENT_MAPPING = {
  'vm-status': ['vm.created', 'vm.updated', 'vm.deleted'],
  'kubernetes-status': ['kubernetes.*.*.*.clusters.created', ...],
  'network-status': ['network.*.*.*.vpcs.created', ...],
  // ...
};
```

---

## 예제 코드

### 예제 1: Kubernetes 클러스터 목록 페이지

```typescript
import { useQuery } from '@tanstack/react-query';
import { useSSEEvents } from '@/hooks/use-sse-events';
import { queryKeys } from '@/lib/query';
import { kubernetesService } from '@/features/kubernetes';

function ClustersPage() {
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const provider = 'aws';

  // SSE 이벤트 자동 처리
  useSSEEvents();

  // 클러스터 목록 조회
  const { data: clusters, isLoading } = useQuery({
    queryKey: queryKeys.kubernetesClusters.list(
      undefined,
      provider,
      selectedCredentialId || '',
      selectedRegion || ''
    ),
    queryFn: () =>
      kubernetesService.getClusters(
        provider,
        selectedCredentialId || '',
        selectedRegion || ''
      ),
    enabled: !!selectedCredentialId && !!selectedRegion,
  });

  return (
    <div>
      {clusters?.map((cluster) => (
        <ClusterCard key={cluster.id} cluster={cluster} />
      ))}
    </div>
  );
}
```

### 예제 2: VPC 목록 페이지 (커스텀 이벤트 핸들러)

```typescript
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { sseService } from '@/services/sse';
import { queryKeys } from '@/lib/query';
import { networkService } from '@/features/networks';
import { applyVPCCreatedUpdate } from '@/lib/sse/query-updates';

function VPCsPage() {
  const queryClient = useQueryClient();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const provider = 'aws';

  // 커스텀 이벤트 핸들러
  useEffect(() => {
    const callbacks = {
      onNetworkVPCCreated: (data) => {
        // 실시간 업데이트
        applyVPCCreatedUpdate(queryClient, data);
        // 추가 로직 (예: 토스트 알림)
        toast.success('VPC created successfully');
      },
    };

    sseService.updateCallbacks(callbacks);
  }, [queryClient]);

  // VPC 목록 조회
  const { data: vpcs } = useQuery({
    queryKey: queryKeys.vpcs.list(
      provider,
      selectedCredentialId || '',
      selectedRegion || ''
    ),
    queryFn: () =>
      networkService.getVPCs(
        provider,
        selectedCredentialId || '',
        selectedRegion || ''
      ),
  });

  return <div>{/* ... */}</div>;
}
```

### 예제 3: 대시보드 페이지

```typescript
import { useDashboardSSE } from '@/hooks/use-dashboard-sse';
import { useCredentialContext } from '@/hooks/use-credential-context';

function DashboardPage() {
  const { widgets, setWidgets } = useDashboard();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();

  // 대시보드 SSE 동적 구독
  useDashboardSSE({
    widgets,
    credentialId: selectedCredentialId || undefined,
    region: selectedRegion || undefined,
    includeSummary: true,
    enabled: widgets.length > 0,
  });

  return (
    <div>
      {widgets.map((widget) => (
        <Widget key={widget.id} widget={widget} />
      ))}
    </div>
  );
}
```

### 예제 4: 특정 이벤트만 구독하는 컴포넌트

```typescript
import { useEffect } from 'react';
import { sseService } from '@/services/sse';
import { useQueryClient } from '@tanstack/react-query';

function ClusterDetailPage({ clusterId }: { clusterId: string }) {
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!sseService.isConnected()) {
      return;
    }

    // 특정 클러스터의 이벤트만 구독
    const eventTypes = [
      `kubernetes.*.*.*.clusters.${clusterId}.updated`,
      `kubernetes.*.*.*.clusters.${clusterId}.deleted`,
    ];

    sseService.subscribeToEvents(eventTypes);

    const callbacks = {
      onKubernetesClusterUpdated: (data) => {
        if (data.clusterId === clusterId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.kubernetesClusters.detail(clusterId),
          });
        }
      },
      onKubernetesClusterDeleted: (data) => {
        if (data.clusterId === clusterId) {
          // 클러스터 삭제 시 리다이렉트
          router.push('/kubernetes/clusters');
        }
      },
    };

    sseService.updateCallbacks(callbacks);

    return () => {
      // cleanup: 구독 해제
      sseService.unsubscribeFromEvents(eventTypes);
    };
  }, [clusterId, queryClient]);

  return <div>{/* ... */}</div>;
}
```

---

## 모범 사례

### 1. 폴링 제거

SSE를 사용하는 경우 `refetchInterval`을 제거합니다:

```typescript
// ❌ 나쁜 예: 폴링 사용
const { data } = useQuery({
  queryKey: queryKeys.vms.list(workspaceId),
  queryFn: () => vmService.getVMs(workspaceId),
  refetchInterval: 30000, // 폴링 제거!
});

// ✅ 좋은 예: SSE 사용
const { data } = useQuery({
  queryKey: queryKeys.vms.list(workspaceId),
  queryFn: () => vmService.getVMs(workspaceId),
  // refetchInterval 제거, SSE로 자동 업데이트
});
```

### 2. 이벤트 구독 최소화

필요한 이벤트만 구독하여 네트워크 트래픽을 최소화합니다:

```typescript
// ❌ 나쁜 예: 모든 이벤트 구독
sseService.subscribeToEvents(['*']);

// ✅ 좋은 예: 필요한 이벤트만 구독
const requiredEvents = new Set([
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.created',
  'kubernetes.aws.cred-123.ap-northeast-2.clusters.updated',
]);
sseService.syncSubscriptions(requiredEvents, {
  credential_ids: ['cred-123'],
  regions: ['ap-northeast-2'],
});
```

### 3. 실시간 업데이트 우선 사용

이벤트 데이터에 리소스 객체가 포함된 경우, 실시간 업데이트를 우선 사용합니다:

```typescript
// ✅ 좋은 예: 실시간 업데이트 시도 후 fallback
try {
  applyKubernetesClusterCreatedUpdate(queryClient, eventData);
} catch (error) {
  // Fallback: 무효화
  queryClient.invalidateQueries({ ... });
}
```

### 4. 대시보드에서는 useDashboardSSE 사용

대시보드 페이지에서는 `useDashboardSSE` 훅을 사용하여 위젯별로 필요한 이벤트만 동적으로 구독합니다:

```typescript
// ✅ 좋은 예
useDashboardSSE({
  widgets,
  credentialId: selectedCredentialId,
  region: selectedRegion,
  includeSummary: true,
});
```

### 5. 컴포넌트 언마운트 시 정리

커스텀 이벤트 핸들러를 등록한 경우, 컴포넌트 언마운트 시 정리합니다:

```typescript
useEffect(() => {
  const callbacks = { /* ... */ };
  sseService.updateCallbacks(callbacks);

  return () => {
    // cleanup (필요시)
    sseService.unsubscribeFromEvents(eventTypes);
  };
}, []);
```

### 6. 에러 처리

SSE 연결 실패 시 적절한 에러 처리를 수행합니다:

```typescript
sseService.connect(token, {
  onError: (error) => {
    console.error('SSE connection error', error);
    // 사용자에게 알림 또는 재연결 시도
    toast.error('Real-time updates unavailable');
  },
});
```

---

## 참고 자료

- [SSE Service 구현](../frontend/src/services/sse.ts)
- [useSSEEvents Hook 구현](../frontend/src/hooks/use-sse-events.ts)
- [useDashboardSSE Hook 구현](../frontend/src/hooks/use-dashboard-sse.ts)
- [실시간 업데이트 유틸리티](../frontend/src/lib/sse/query-updates.ts)
- [세밀한 무효화 유틸리티](../frontend/src/lib/sse/query-invalidation.ts)
- [Backend SSE 적용 가이드](./SSE_BACKEND_GUIDE.md)

