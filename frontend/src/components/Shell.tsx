import {type ReactNode} from 'react';
import {Activity, History, LayoutDashboard, RefreshCw, Settings2} from 'lucide-react';
import {Badge, Button, StatusLamp, Tooltip, TooltipProvider} from '@claudepro/ui';
import {ThemeControl} from '../theme';
import {TitleBar} from './TitleBar';
import type {Tab} from '../useEnvironment';

const navigation = [
  {key:'home', label:'环境概览', icon:LayoutDashboard},
  {key:'diagnostics', label:'环境检测', icon:Activity},
  {key:'history', label:'恢复记录', icon:History},
] as const;

export type ShellStatus = {tone:'ok'|'warn'|'info'|'unknown';text:string;demo?:boolean};

export function Shell({tab,onNavigate,status,busy,disabled,onRefresh,onSettings,children}: {
  tab:Tab; onNavigate:(tab:Tab)=>void; status:ShellStatus; busy:boolean; disabled:boolean;
  onRefresh:()=>void; onSettings:()=>void; children:ReactNode;
}) {
  const statusPill=<span className="titlebar-status">
    <Badge className="environment-status" variant="soft" tone={status.tone==='unknown'?'neutral':status.tone}><StatusLamp tone={status.tone} size="sm"/>{status.text}</Badge>
    {status.demo&&<Badge variant="soft" tone="info">模拟模式</Badge>}
  </span>;
  const tools=<>
    <Tooltip content="重新读取本机状态" side="bottom"><Button variant="ghost" size="icon-sm" aria-label="刷新状态" disabled={disabled} onClick={onRefresh}><RefreshCw/></Button></Tooltip>
    <ThemeControl/>
    <Tooltip content="客户端设置" side="bottom"><Button variant="ghost" size="icon-sm" aria-label="客户端设置" disabled={disabled} onClick={onSettings}><Settings2/></Button></Tooltip>
  </>;
  return <TooltipProvider delay={300}><div className="app-frame">
    <TitleBar status={statusPill} tools={tools}/>
    <div className="shell">
      <nav className="rail" aria-label="主导航"><div className="rail-group">
        <span className="rail-group-label mg-engraved">工作空间</span>
        {navigation.map(({key,label,icon:Icon})=>{
          const active=key===tab;
          return <Tooltip key={key} content={label} side="right"><Button variant="outline" className={'rail-item'+(active?' active is-pressed':'')} aria-label={label} aria-current={active?'page':undefined} disabled={busy} onClick={()=>onNavigate(key)}>
            <Icon/><span className="rail-label">{label}</span><StatusLamp className="rail-led" size="sm" tone={active?'active':'off'}/>
          </Button></Tooltip>;
        })}
      </div></nav>
      <main className="shell-main"><div className="shell-content">{children}</div></main>
    </div>
  </div></TooltipProvider>;
}
