// Adapted from MiGate apps/console/src/components/workspace.tsx.
import {useEffect, useRef, useState, type ReactNode} from 'react';
import {Activity, ArrowLeftRight, History, LayoutDashboard, Menu, PanelLeftClose, PanelLeftOpen, RefreshCw, Settings2} from 'lucide-react';
import {Button, Sheet, SheetContent, StatusLamp, Tooltip, TooltipProvider} from '@claudepro/ui';
import {ThemeControl} from '../theme';
import type {Tab} from '../useEnvironment';

const navigation = [
  {key:'home', label:'环境概览', icon:LayoutDashboard},
  {key:'diagnostics', label:'环境检测', icon:Activity},
  {key:'settings', label:'切换设置', icon:Settings2},
  {key:'history', label:'恢复记录', icon:History},
] as const;
const storageKey = 'claudepro-rail-collapsed';

export function Workspace({tab,onNavigate,busy,disabled,onRefresh,children}: {
  tab:Tab; onNavigate:(tab:Tab)=>void; busy:boolean; disabled:boolean; onRefresh:()=>void; children:ReactNode;
}) {
  const [collapsed,setCollapsed]=useState(()=>{try{return localStorage.getItem(storageKey)==='1';}catch{return false;}});
  const [drawer,setDrawer]=useState(false);
  const trigger=useRef<HTMLButtonElement>(null);
  useEffect(()=>{try{localStorage.setItem(storageKey,collapsed?'1':'0');}catch{/* Keep this session's choice. */}},[collapsed]);
  const nav=(compact:boolean)=><nav className="rail-nav" aria-label="主导航"><div className="rail-group">
    <span className="rail-group-label mg-engraved">工作空间</span>
    {navigation.map(({key,label,icon:Icon})=>{
      const button=<Button key={key} variant="outline" className={'rail-item'+(key===tab?' active is-pressed':'')} aria-label={label} aria-current={key===tab?'page':undefined} disabled={busy} onClick={()=>{onNavigate(key);setDrawer(false);}}>
        <Icon/><span className="rail-label">{label}</span><StatusLamp className="rail-led" size="sm" tone={key===tab?'active':'off'}/>
      </Button>;
      return compact?<Tooltip key={key} content={label}>{button}</Tooltip>:button;
    })}
  </div></nav>;
  return <TooltipProvider delay={300}><div className="shell" data-collapsed={collapsed}>
    <aside className="rail">
      <div className="rail-brand" aria-label="ClaudePro"><span className="rail-mark"><ArrowLeftRight size={19}/></span><strong>ClaudePro</strong></div>
      {nav(collapsed)}
      <div className="rail-foot">
        <Button className="rail-collapse" variant="ghost" size="sm" aria-label={collapsed?'展开导航':'收起导航'} onClick={()=>setCollapsed(v=>!v)}>{collapsed?<PanelLeftOpen/>:<><PanelLeftClose/><span>收起导航</span></>}</Button>
        <div className="local-status"><StatusLamp tone="active"/><span>本地运行<small>Windows · v0.1.0</small></span></div>
      </div>
    </aside>
    <div className="shell-body">
      <header className="topbar">
        <Button ref={trigger} className="mobile-nav" variant="ghost" size="icon-sm" aria-label="打开导航" aria-expanded={drawer} onClick={()=>setDrawer(true)}><Menu/></Button>
        <div className="topbar-title"><h1>{navigation.find(n=>n.key===tab)?.label}</h1></div>
        <div className="topbar-tools"><Button variant="ghost" size="sm" disabled={disabled} onClick={onRefresh}><RefreshCw/>刷新状态</Button><ThemeControl/></div>
      </header>
      <main className="shell-main">{children}<footer className="app-footer"><span><StatusLamp tone="active" size="sm"/>本地备份 · 逐项验证 · 手动还原</span><span>ClaudePro / Windows</span></footer></main>
    </div>
    <Sheet open={drawer} onOpenChange={setDrawer}><SheetContent title="ClaudePro" description="工作空间导航" finalFocus={trigger}>{nav(false)}</SheetContent></Sheet>
  </div></TooltipProvider>;
}
