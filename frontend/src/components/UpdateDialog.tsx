import {Download, ExternalLink, RotateCw} from 'lucide-react';
import {Badge, Button, Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle, Notice} from '@claudepro/ui';
import type {UpdateModel} from '../useUpdate';

function sizeLabel(n?:number) {
  if(!n)return '';
  if(n<1024)return n+' B';
  if(n<1024*1024)return (n/1024).toFixed(1)+' KB';
  return (n/1024/1024).toFixed(1)+' MB';
}

export function UpdateDialog({u,current}:{u:UpdateModel;current?:string}) {
  const rel=u.release;
  const downloading=u.phase==='downloading';
  const ready=u.phase==='ready';
  const locked=u.busyEnv||downloading;
  return <Dialog open={u.open} onOpenChange={open=>{if(!downloading&&!ready) {if(!open)u.close();}}}>
    <DialogContent size="md" showCloseButton={!downloading&&!ready}>
      <DialogTitle>{ready?'更新已就绪':'发现新版本'}</DialogTitle>
      <DialogDescription>{current&&rel?.version?`当前 ${current} → ${rel.version}`:rel?.version?`新版本 ${rel.version}`:'GitHub Releases 有可用更新。'}</DialogDescription>
      {rel&&<div className="page-stack">
        <p><Badge variant="soft" tone="info">{rel.filename||'安装包'}</Badge>{rel.size?` ${sizeLabel(rel.size)}`:''}</p>
        {rel.notes&&<pre className="detection-note" style={{whiteSpace:'pre-wrap',maxHeight:160,overflow:'auto'}}>{rel.notes.slice(0,800)}</pre>}
      </div>}
      {downloading&&<Notice role="status">正在下载 {u.progress}%</Notice>}
      {ready&&<Notice tone="ok" role="status">已下载并校验，重启后替换当前程序。</Notice>}
      {u.busyEnv&&<Notice tone="warn">正在切换或恢复环境，请完成后再安装更新。</Notice>}
      {u.error&&<Notice tone="bad">{u.error}</Notice>}
      {u.allowInstall===false&&<Notice>当前是开发构建，不会替换正在运行的程序。请打开 GitHub Releases 下载正式版。</Notice>}
      <DialogFooter>
        {!ready&&!downloading&&<Button variant="outline" disabled={locked} onClick={()=>void u.skip()}>跳过此版本</Button>}
        {!ready&&!downloading&&<Button variant="outline" disabled={locked} onClick={u.close}>稍后提醒</Button>}
        {u.allowInstall&&!ready&&<Button disabled={locked} onClick={()=>void u.install()}><Download/>立即更新</Button>}
        {u.allowInstall===false&&<Button onClick={u.openReleases}><ExternalLink/>打开下载页</Button>}
        {ready&&<Button disabled={u.busyEnv} onClick={()=>void u.restart()}><RotateCw/>重启以完成更新</Button>}
      </DialogFooter>
    </DialogContent>
  </Dialog>;
}
