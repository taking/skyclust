/**
 * Kubeconfig Download Utility
 * kubeconfig 파일 다운로드 헬퍼 함수
 */

export function downloadKubeconfig(kubeconfig: string, clusterName: string): void {
  const blob = new Blob([kubeconfig], { type: 'application/yaml' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `kubeconfig-${clusterName}.yaml`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

