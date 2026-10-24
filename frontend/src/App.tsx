import {useCallback, useEffect, useRef} from 'react';
import {ArrowLeftRight, ArrowUpRight, Clock3, FolderOpen, Globe2, History, Languages, LoaderCircle, Monitor, RotateCcw, Save, ShieldCheck, Terminal, X} from 'lucide-react';
import {Badge, Button, Card, CardContent, CardHeader, Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle, Field, Input, Notice, PageHeader, PageState, SectionHeader, StatusLamp, Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from '@claudepro/ui';
import {TargetPicker} from './components/TargetPicker';
import {Diagnostics} from './components/Diagnostics';
import {Workspace} from './components/Workspace';
import {fields, labels, useEnvironment, type EnvironmentModel} from './useEnvironment';
import './style.css';

const icons={timezone:Clock3,culture:Languages,region:Globe2};

function Overview({m,restoreButton}:{m:EnvironmentModel;restoreButton:React.RefObject<HTMLButtonElement|null>}) {
  const {state,journal}=m;
  if(!state)return null;
  return <>
    <Card><CardContent className="settings-content"><TargetPicker m={m}/></CardContent></Card>
    <Card>
      <CardHeader><SectionHeader title="2. 查看差异" description="3 项 Windows 设置 · 修改前自动备份" flush actions={<ShieldCheck className="section-icon"/>}/></CardHeader>
      <Table className="environment-table" aria-label="当前环境与目标环境对照">
        <TableHeader><TableRow><TableHead>配置项</TableHead><TableHead>当前环境</TableHead><TableHead>目标环境</TableHead><TableHead>差异</TableHead></TableRow></TableHeader>
        <TableBody>{fields.map(f=>{
          const current=state.view.current?.[f]?.label;const target=m.draft.target[f];const Icon=icons[f];const same=current===target;
          return <TableRow key={f}>
            <TableCell><div className="setting-label"><span className="setting-icon"><Icon size={18}/></span><div><strong>{labels[f]}</strong><small>{f==='timezone'?'影响系统全局':'影响当前用户'}</small></div></div></TableCell>
            <TableCell><span className="setting-value">{m.label(f,current)}</span><small className="setting-id">{current}</small></TableCell>
            <TableCell><span className="setting-value">{m.label(f,target)}</span><small className="setting-id">{target||'请先完成切换设置'}</small></TableCell>
            <TableCell><Badge variant="soft" tone={!target?'neutral':same?'ok':'warn'}>{!target?'待配置':same?'相同':'不同'}</Badge></TableCell>
          </TableRow>;
        })}</TableBody>
      </Table>
    </Card>
      <div className="action-bar mg-panel"><small><ShieldCheck size={15}/>恢复真实原值，包括自定义区域格式。</small><div className="button-row">
        <Button ref={restoreButton} variant="outline" disabled={m.disabled||!m.active} onClick={m.restore}><RotateCcw/>恢复原环境</Button>
        <Button disabled={m.disabled||m.active||!m.configured} onClick={m.switchEnvironment}><ArrowLeftRight/>切换环境</Button>
      </div></div>
    {!m.configured&&<Notice>选择上方任一地区模板即可完成目标配置，也可展开自定义搜索选项。</Notice>}
    {m.active&&<Notice tone="warn">{journal?.status==='active'?'目标环境正在生效。退出应用不会自动恢复，使用结束后请点击“恢复原环境”。':'检测到未完成操作或配置变化。请先恢复原环境，再开始新的切换。'}</Notice>}
    <div className="launch-grid">{(['cli','desktop'] as const).map(client=>{
      const cli=client==='cli';const Icon=cli?Terminal:Monitor;const path=cli?state.config.cliPath||m.clients?.cli:state.config.desktopPath||m.clients?.desktop;
      return <Card key={client}><CardContent className="client-content"><div className="client-heading"><span className="client-icon mg-well"><Icon size={21}/></span><div><h3>{cli?'Claude Code':'Claude Desktop'}</h3><p>{cli?'在新终端中打开 CLI':'打开或激活桌面客户端'}</p></div></div>
        <code className="client-path" title={path||''}>{path||`未检测到 ${cli?'CLI':'Desktop'}`}</code>
        <Button variant={journal?.status==='active'?'default':'outline'} disabled={m.disabled||(m.active&&journal?.status!=='active')} onClick={()=>m.launch(client)}><ArrowUpRight/>{cli?'启动 CLI':'启动 Desktop'}</Button>
      </CardContent></Card>;
    })}</div>
    <Diagnostics m={m} summary/>
    <p className="scope-note"><ShieldCheck size={16}/><span>系统时区和区域设置会影响其他应用。切换、恢复前请保存并关闭 Claude；本工具不承诺避免账号封禁。</span></p>
  </>;
}

function Settings({m}:{m:EnvironmentModel}) {
  return <Card><CardContent className="settings-content">
    <TargetPicker m={m}/>
    <div className="settings-divider"/>
    <SectionHeader title="客户端启动" description="保留现有配置和凭据；路径留空时自动发现客户端。"/>
    <div className="path-fields">{(['workDir','cliPath','desktopPath'] as const).map(f=>{
      const name=f==='workDir'?'CLI 工作目录':f==='cliPath'?'CLI 路径（可选）':'Desktop 路径（可选）';
      const hint=f==='workDir'?'在此目录打开新的 PowerShell 窗口并运行 Claude。':f==='desktopPath'?'Microsoft Store 安装版建议留空，由应用自动发现。':undefined;
      return <Field key={f} label={name} htmlFor={f} hint={hint}><div className="path-input">
        <Input id={f} aria-describedby={hint?f+'-hint':undefined} value={m.draft[f]} disabled={m.disabled} onChange={e=>m.setDraft(d=>({...d,[f]:e.target.value}))} placeholder={f==='workDir'?'例如 C:\\develop':f==='cliPath'?m.clients?.cli||'留空自动检测 .exe / .cmd / .ps1':'留空自动检测，或选择 .exe 文件'}/>
        <Button variant="outline" disabled={m.disabled} aria-label={'浏览'+(f==='workDir'?'工作目录':f==='cliPath'?' CLI':' Desktop')} onClick={()=>m.browse(f)}><FolderOpen/>浏览</Button>
      </div></Field>;
    })}</div>
    <div className="form-footer"><small>仅管理三项区域设置。显示语言、输入法、网络及凭据保持原样。</small><Button disabled={m.disabled||!m.catalog||!m.draft.target.timezone||!m.draft.target.culture||!m.draft.target.region} onClick={m.save}><Save/>保存设置</Button></div>
  </CardContent></Card>;
}

function RecoveryHistory({m}:{m:EnvironmentModel}) {
  const journal=m.journal;
  return <Card><CardContent className="history-content"><SectionHeader title="当前 / 最近一次操作" description={journal?.id||'尚无操作记录'}/>
    {journal?<><div className="originals">{fields.map(f=><div key={f}><small>{labels[f]} · 切换前</small><strong>{m.label(f,journal.original?.[f]?.label)}</strong></div>)}</div>
      <ol className="timeline">{[...(journal.log||[])].reverse().map((entry,i)=><li key={i}><StatusLamp tone={i===0?'active':'off'} size="sm"/><div><strong>{entry.message}</strong><time>{entry.time} · UTC</time></div></li>)}</ol></>:<PageState kind="empty" title="尚无恢复记录" detail="首次切换后，这里会显示备份与恢复步骤。"/>}
    <div className="backup-path"><SectionHeader level="h3" size="sm" title={<><History size={15}/>本地备份目录</>}/><code>{m.state?.dataDir}</code><p>journal.json 保存当前恢复记录，history 保存之前的记录。存在未恢复操作时，请勿删除这些文件。</p></div>
  </CardContent></Card>;
}

export default function App() {
  const m=useEnvironment();
  const restoreButton=useRef<HTMLButtonElement>(null);
  const dismissExit=useRef<HTMLButtonElement>(null);
  // MiGate's footer is portalled after the popup mounts. Focus the safe action
  // when that actual button appears, and after an in-flight operation settles.
  const focusDismiss=useCallback((node:HTMLButtonElement|null)=>{dismissExit.current=node;node?.focus();},[]);
  useEffect(()=>{if(m.exit&&!m.busy)dismissExit.current?.focus();},[m.exit,m.busy]);
  const statusTone=m.loading?'neutral':m.busy?'info':m.active?'warn':'ok';
  return <Workspace tab={m.tab} onNavigate={m.setTab} busy={m.busy} disabled={m.disabled} onRefresh={m.reload}>
    <div className="page-stack">
      <PageHeader eyebrow="CLAUDEPRO / ENVIRONMENT" title={m.tab==='home'?'选好地区，一键切换。':m.tab==='settings'?'为下一次工作做好准备。':m.tab==='diagnostics'?'看清环境，再做选择。':'每一次修改，都有记录。'} description={m.tab==='home'?'保存当前配置，再进入你选定的工作环境。':m.tab==='settings'?'地区模板快速配置，也可搜索并自定义。':m.tab==='diagnostics'?'本机检查自动更新，网络检测由你触发。':'原始值、执行步骤与恢复结果均保存在本机。'} actions={<Badge className="environment-status" variant="soft" tone={statusTone}><StatusLamp tone={m.loading?'unknown':m.busy?'info':m.active?'warn':'ok'}/>{m.loading?'读取中':m.busy?'正在处理':m.status}</Badge>}/>
      {m.state?.demo&&<Notice>模拟模式 · 所有切换只写入独立演示文件，不修改 Windows，也不启动 Claude。</Notice>}
      {m.state?.warning&&<Notice tone="warn" role="alert">{m.state.warning}</Notice>}
      {m.error&&<Notice tone="bad"><div className="error-content"><div><strong>操作未完成</strong><p>{m.error}</p></div><Button variant="ghost" size="icon-sm" aria-label="收起错误" onClick={()=>m.setError('')}><X/></Button></div></Notice>}
      {m.busy&&<Notice role="status"><span className="operation-progress"><LoaderCircle className="animate-spin" size={16}/>{m.progress||'正在处理…'}</span></Notice>}
      {!m.state&&<Card><PageState kind={m.loading?'loading':'error'} title={m.loading?'正在读取本机环境…':'无法读取状态'} detail={m.loading?'检查 Windows 设置和恢复记录。':'为保护原配置，修改功能已停用。请查看错误或刷新重试。'}/></Card>}
      {m.tab==='diagnostics'?<Diagnostics m={m}/>:m.state&&(m.tab==='home'?<Overview m={m} restoreButton={restoreButton}/>:m.tab==='settings'?<Settings m={m}/>:<RecoveryHistory m={m}/>)}
    </div>
    <Dialog open={m.exit&&m.conflicts.length===0} onOpenChange={open=>{if(!m.busy)m.setExit(open);}}>
      <DialogContent size="md" showCloseButton={!m.busy} initialFocus={dismissExit}>
        <DialogTitle>环境尚未恢复</DialogTitle><DialogDescription>退出只关闭 ClaudePro，当前 Windows 设置会继续生效。恢复记录保存在本机，下次启动可继续恢复。</DialogDescription>
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
  </Workspace>;
}
