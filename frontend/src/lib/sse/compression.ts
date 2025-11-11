/**
 * SSE Message Compression Utilities
 * 
 * Backend에서 전송하는 압축된 SSE 메시지를 처리합니다.
 * - Base64 인코딩된 gzip 데이터를 디코딩하고 압축 해제
 * - Browser의 TextDecoder와 pako 라이브러리 사용
 */

import pako from 'pako';
import { log } from '../logging';

/**
 * SSE 압축 메시지를 압축 해제합니다.
 * Backend에서 전송하는 형식: Base64 인코딩된 gzip 압축 데이터
 * 
 * @param compressedData - Base64 인코딩된 gzip 압축 데이터
 * @returns 압축 해제된 원본 JSON 문자열
 */
export function decompressSSEMessage(compressedData: string): string {
  try {
    // Base64 디코딩
    const binaryString = atob(compressedData);
    
    // Binary string을 Uint8Array로 변환
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }
    
    // Gzip 압축 해제
    const decompressed = pako.ungzip(bytes, { to: 'string' });
    
    return decompressed;
  } catch (error) {
    log.error('Failed to decompress SSE message', error, { compressedDataLength: compressedData.length });
    throw new Error(`Failed to decompress SSE message: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * SSE 메시지가 압축되어 있는지 확인합니다.
 * Backend에서 전송하는 형식: "compressed: true" 플래그
 * 
 * @param event - SSE EventSource 이벤트
 * @returns 압축 여부
 */
export function isCompressedMessage(_event: MessageEvent): boolean {
  // EventSource의 기본 파싱은 "compressed: true" 플래그를 직접 처리하지 않으므로
  // 커스텀 파서를 사용하거나 메시지 데이터를 분석해야 함
  // 실제로는 SSE 메시지 파싱 시 플래그를 확인해야 함
  return false; // 기본값, 실제 파싱 로직에서 처리
}

/**
 * SSE 메시지를 파싱합니다.
 * Backend에서 전송하는 형식:
 * - event: <eventType>
 * - compressed: true (압축된 경우)
 * - data: <JSON 문자열 또는 Base64 인코딩된 압축 데이터>
 * 
 * @param rawData - SSE 원본 데이터 문자열
 * @param isCompressed - 압축 여부 플래그
 * @returns 파싱된 메시지 데이터
 */
export function parseSSEMessage(rawData: string, isCompressed: boolean = false): unknown {
  try {
    let jsonString: string;
    
    if (isCompressed) {
      // 압축 해제
      jsonString = decompressSSEMessage(rawData);
    } else {
      jsonString = rawData;
    }
    
    // JSON 파싱
    return JSON.parse(jsonString);
  } catch (error) {
    log.error('Failed to parse SSE message', error, { isCompressed, rawDataLength: rawData.length });
    throw new Error(`Failed to parse SSE message: ${error instanceof Error ? error.message : String(error)}`);
  }
}

