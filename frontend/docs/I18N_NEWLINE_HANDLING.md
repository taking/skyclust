# 다국어 번역에서 줄바꿈(\n) 처리 방법

## 개요

`ko.json`과 `en.json` 파일에 `\n`을 넣으면 JSON 파싱 시 이스케이프 문자로 처리되지만, React/HTML에서는 기본적으로 줄바꿈이 렌더링되지 않습니다.

## 문제

JSON 파일에 다음과 같이 작성하면:

```json
{
  "message": "첫 번째 줄\n두 번째 줄"
}
```

React 컴포넌트에서 `t('message')`를 사용하면:
- 문자열은 `"첫 번째 줄\n두 번째 줄"`로 파싱됩니다
- 하지만 HTML에서는 `\n`이 공백으로 처리되어 한 줄로 표시됩니다

## 해결 방법

### 방법 1: CSS `white-space` 사용 (권장)

```tsx
<div style={{ whiteSpace: 'pre-line' }}>
  {t('message')}
</div>
```

또는 Tailwind CSS 클래스 사용:

```tsx
<div className="whitespace-pre-line">
  {t('message')}
</div>
```

### 방법 2: `dangerouslySetInnerHTML` 사용 (비권장)

```tsx
<div dangerouslySetInnerHTML={{ __html: t('message').replace(/\n/g, '<br />') }} />
```

**주의**: XSS 공격 위험이 있으므로 신뢰할 수 있는 데이터에만 사용하세요.

### 방법 3: 컴포넌트로 분리

```tsx
function MultilineText({ text }: { text: string }) {
  return (
    <div className="whitespace-pre-line">
      {text}
    </div>
  );
}

// 사용
<MultilineText text={t('message')} />
```

## 예시

### JSON 파일

```json
{
  "description": "이것은 첫 번째 줄입니다.\n이것은 두 번째 줄입니다.\n이것은 세 번째 줄입니다."
}
```

### React 컴포넌트

```tsx
import { useTranslation } from '@/hooks/use-translation';

export function ExampleComponent() {
  const { t } = useTranslation();
  
  return (
    <div className="whitespace-pre-line">
      {t('description')}
    </div>
  );
}
```

### 결과

```
이것은 첫 번째 줄입니다.
이것은 두 번째 줄입니다.
이것은 세 번째 줄입니다.
```

## 권장 사항

1. **일반 텍스트**: `white-space: pre-line` CSS 사용
2. **긴 텍스트**: 여러 번역 키로 분리하거나 배열 사용
3. **동적 콘텐츠**: React 컴포넌트로 구조화

## 참고

- `white-space: pre-line`: 연속된 공백은 하나로 합치고, 줄바꿈(`\n`)은 유지
- `white-space: pre`: 모든 공백과 줄바꿈 유지
- `white-space: pre-wrap`: `pre`와 동일하지만 자동 줄바꿈도 허용

