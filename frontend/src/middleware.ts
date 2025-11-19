import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

/**
 * Next.js Middleware
 * 
 * 새로운 라우팅 구조 지원:
 * - /w/{workspaceId}/* -> /{workspaceId}/*
 * - /w/{workspaceId}/c/{credentialId}/* -> /{workspaceId}/{credentialId}/*
 */

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // /w/{workspaceId}/c/{credentialId}/* 패턴 매칭 (우선 처리)
  const wcPattern = /^\/w\/([^/]+)\/c\/([^/]+)(\/.*)?$/;
  const wcMatch = pathname.match(wcPattern);

  if (wcMatch) {
    const [, workspaceId, credentialId, restPath] = wcMatch;
    let newPath = `/${workspaceId}/${credentialId}${restPath || ''}`;
    
    // k8s -> kubernetes 변환 (실제 파일 구조와 일치시키기)
    newPath = newPath.replace(/\/k8s\//g, '/kubernetes/');
    
    // Query parameters 유지
    const url = request.nextUrl.clone();
    url.pathname = newPath;
    
    return NextResponse.rewrite(url);
  }

  // /w/{workspaceId}/* 패턴 매칭
  const wPattern = /^\/w\/([^/]+)(\/.*)?$/;
  const wMatch = pathname.match(wPattern);

  if (wMatch) {
    const [, workspaceId, restPath] = wMatch;
    const newPath = `/${workspaceId}${restPath || ''}`;
    
    // Query parameters 유지
    const url = request.nextUrl.clone();
    url.pathname = newPath;
    
    return NextResponse.rewrite(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (public folder)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};

