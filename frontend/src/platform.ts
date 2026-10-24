import {System} from '@wailsio/runtime';

export type PlatformKind = 'windows'|'darwin'|'web';

// 运行时环境由 Wails 在页面加载前注入；浏览器预览（测试）里不存在，按 Windows 布局处理。
export function platform(): PlatformKind {
  try {
    if(System.IsMac())return 'darwin';
    if(System.IsWindows())return 'windows';
  }catch{/* 非 Wails 环境 */}
  const forced=new URLSearchParams(location.search).get('platform');
  return forced==='darwin'?'darwin':'web';
}

export const isMac=()=>platform()==='darwin';

export const systemName=()=>isMac()?'macOS':'Windows';

/** 在 <html> 上标记平台，供样式切换标题栏布局。 */
export function markPlatform() {
  document.documentElement.dataset.platform=platform();
}
