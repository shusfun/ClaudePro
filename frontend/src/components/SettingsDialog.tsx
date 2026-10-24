import {useEffect, useState} from 'react';
import {FolderOpen, Save} from 'lucide-react';
import {Button, Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle, Field, Input} from '@claudepro/ui';
import type {ClientPaths, EnvironmentModel} from '../useEnvironment';
import {isMac} from '../platform';

const keys=['workDir','cliPath','desktopPath'] as const;

export function SettingsDialog({m}:{m:EnvironmentModel}) {
  const mac=isMac();
  const saved=m.state?.config;
  const [paths,setPaths]=useState<ClientPaths>({workDir:'',cliPath:'',desktopPath:''});
  // 每次打开都从已保存的配置开始，取消即丢弃修改。
  useEffect(()=>{if(m.settingsOpen&&saved)setPaths({workDir:saved.workDir,cliPath:saved.cliPath,desktopPath:saved.desktopPath});},[m.settingsOpen]);
  const meta={
    workDir:{label:'CLI 工作目录',hint:mac?'在此目录打开新的“终端”窗口并运行 Claude。':'在此目录打开新的 PowerShell 窗口并运行 Claude。',placeholder:mac?'例如 ~/develop':'例如 C:\\develop',browse:'浏览工作目录'},
    cliPath:{label:'CLI 路径（可选）',hint:'留空时自动发现。',placeholder:m.clients?.cli||(mac?'留空自动检测 claude 可执行文件':'留空自动检测 .exe / .cmd / .ps1'),browse:'浏览 CLI'},
    desktopPath:{label:'Desktop 路径（可选）',hint:mac?'留空时在“应用程序”中查找 Claude.app。':'Microsoft Store 安装版建议留空，由应用自动发现。',placeholder:m.clients?.desktop||(mac?'留空自动检测 Claude.app':'留空自动检测，或选择 .exe 文件'),browse:'浏览 Desktop'},
  };
  return <Dialog open={m.settingsOpen} onOpenChange={open=>{if(!m.busy)m.setSettingsOpen(open);}}>
    <DialogContent size="md" showCloseButton={!m.busy}>
      <DialogTitle>客户端设置</DialogTitle>
      <DialogDescription>保留现有配置和凭据；路径留空时自动发现客户端。</DialogDescription>
      <div className="path-fields">{keys.map(f=><Field key={f} label={meta[f].label} htmlFor={f} hint={meta[f].hint}><div className="path-input">
        <Input id={f} aria-describedby={f+'-hint'} value={paths[f]} disabled={m.busy} placeholder={meta[f].placeholder} onChange={e=>setPaths(p=>({...p,[f]:e.target.value}))}/>
        <Button variant="outline" disabled={m.busy} aria-label={meta[f].browse} onClick={async()=>{const path=await m.browse(f);if(path)setPaths(p=>({...p,[f]:path}));}}><FolderOpen/>浏览</Button>
      </div></Field>)}</div>
      <DialogFooter>
        <Button variant="outline" disabled={m.busy} onClick={()=>m.setSettingsOpen(false)}>取消</Button>
        <Button disabled={m.busy||!paths.workDir} onClick={()=>m.save(paths)}><Save/>保存设置</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>;
}
