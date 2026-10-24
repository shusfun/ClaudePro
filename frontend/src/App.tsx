import {useCallback, useEffect, useRef} from 'react';
import {X} from 'lucide-react';
import {Button, Card, Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle, Notice, PageState} from '@claudepro/ui';
import {Shell, type ShellStatus} from './components/Shell';
import {Home} from './components/Home';
import {Diagnostics} from './components/Diagnostics';
import {History} from './components/History';
import {SettingsDialog} from './components/SettingsDialog';
import {labels, useEnvironment} from './useEnvironment';
import {systemName} from './platform';
import './style.css';

export default function App() {
  const m=useEnvironment();
  const restoreButton=useRef<HTMLButtonElement>(null);
  const dismissExit=useRef<HTMLButtonElement>(null);
  // MiGate's footer is portalled after the popup mounts. Focus the safe action
  // when that actual button appears, and after an in-flight operation settles.
  const focusDismiss=useCallback((node:HTMLButtonElement|null)=>{dismissExit.current=node;node?.focus();},[]);
  useEffect(()=>{if(m.exit&&!m.busy)dismissExit.current?.focus();},[m.exit,m.busy]);
  const status:ShellStatus={
    tone:m.loading?'unknown':m.busy?'info':m.active?'warn':'ok',
    text:m.loading?'读取中':m.busy?'正在处理':m.status,
    demo:m.state?.demo,
  };
  return <Shell tab={m.tab} onNavigate={m.setTab} status={status} busy={m.busy} disabled={m.disabled} onRefresh={m.reload} onSettings={()=>m.setSettingsOpen(true)}>
    {(m.error||m.state?.warning)&&<div className="page-stack">
      {m.state?.warning&&<Notice tone="warn" role="alert">{m.state.warning}</Notice>}
      {m.error&&<Notice tone="bad"><div className="error-content"><div><strong>操作未完成</strong><p>{m.error}</p></div><Button variant="ghost" size="icon-sm" aria-label="收起错误" onClick={()=>m.setError('')}><X/></Button></div></Notice>}
    </div>}
    {!m.state&&m.tab!=='diagnostics'&&<Card><PageState kind={m.loading?'loading':'error'} title={m.loading?'正在读取本机环境…':'无法读取状态'} detail={m.loading?'检查系统设置和恢复记录。':'为保护原配置，修改功能已停用。请查看错误或刷新重试。'}/></Card>}
    {m.tab==='diagnostics'?<Diagnostics m={m}/>:m.state&&(m.tab==='home'?<Home m={m} restoreButton={restoreButton}/>:<History m={m}/>)}
    <SettingsDialog m={m}/>
    <Dialog open={m.exit&&m.conflicts.length===0} onOpenChange={open=>{if(!m.busy)m.setExit(open);}}>
      <DialogContent size="md" showCloseButton={!m.busy} initialFocus={dismissExit}>
        <DialogTitle>环境尚未恢复</DialogTitle><DialogDescription>退出只关闭 ClaudePro，当前 {systemName()} 设置会继续生效。恢复记录保存在本机，下次启动可继续恢复。</DialogDescription>
        {m.error&&<Notice tone="bad">{m.error}</Notice>}
        <DialogFooter><Button ref={focusDismiss} variant="outline" disabled={m.busy} onClick={()=>m.setExit(false)}>留在应用</Button><Button disabled={m.busy} onClick={m.confirmExit}>保留环境并退出</Button></DialogFooter>
      </DialogContent>
    </Dialog>
    <Dialog open={m.conflicts.length>0} onOpenChange={open=>{if(!open&&!m.busy)m.setConflicts([]);}}>
      <DialogContent size="lg" showCloseButton={!m.busy} finalFocus={restoreButton}>
        <DialogTitle>发现配置冲突</DialogTitle><DialogDescription>以下项目发生了外部变化，或上次写入尚未验证。请选择处理方式；保留当前值会记录为主动放弃该项还原。</DialogDescription>
        <div className="conflict-list">{m.conflicts.map(c=><fieldset className="conflict-item" disabled={m.busy} key={c.field}>
          <legend>{labels[c.field as keyof typeof labels]||c.field}</legend><p>当前：{c.current.label} → 原始：{c.original.label}</p>
          <details><summary>查看完整值（包括自定义格式）</summary><pre>{JSON.stringify({当前:c.current,原始:c.original,预期:c.expected},null,2)}</pre></details>
          <div className="radio-options">{[['restore','恢复原始值'],['keep','保留当前值']].map(([action,text])=><label key={action} className={m.decisions[c.field]===action?'selected':''}><input type="radio" name={'conflict-'+c.field} checked={m.decisions[c.field]===action} onChange={()=>m.setDecisions(d=>({...d,[c.field]:action}))}/>{text}</label>)}</div>
        </fieldset>)}</div>
        {m.error&&<Notice tone="bad">{m.error}</Notice>}
        {m.busy&&<Notice role="status">{m.progress||'正在恢复，请稍候…'}</Notice>}
        <DialogFooter><Button variant="outline" disabled={m.busy} onClick={()=>m.setConflicts([])}>取消</Button><Button disabled={m.busy||m.conflicts.some(c=>!m.decisions[c.field])} onClick={m.commitRestore}>确认并继续恢复</Button></DialogFooter>
      </DialogContent>
    </Dialog>
  </Shell>;
}
