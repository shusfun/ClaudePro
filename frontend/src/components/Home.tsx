import {ArrowLeftRight, ArrowUpRight, Clock3, Globe2, Languages, LoaderCircle, Monitor, RotateCcw, ShieldCheck, Terminal} from 'lucide-react';
import {Badge, Button, Card, CardContent, CardHeader, Notice, SectionHeader, StatusLamp, Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from '@claudepro/ui';
import {TargetPicker} from './TargetPicker';
import {fields, labels, type EnvironmentModel} from '../useEnvironment';
import {isMac} from '../platform';

const icons={timezone:Clock3,culture:Languages,region:Globe2};

type Phase='idle'|'ready'|'active'|'needs_restore';

/** 状态主卡：说明当前处境，并只给出一个主操作。 */
function StatusCard({m,restoreButton}:{m:EnvironmentModel;restoreButton:React.RefObject<HTMLButtonElement|null>}) {
  const journal=m.journal;
  const current=m.state?.view.current;
  const target=m.active&&journal?journal.target:m.draft.target;
  const diff=fields.filter(f=>target[f]&&current?.[f]?.label!==target[f]).length;
  const phase:Phase=m.active?(journal?.status==='active'?'active':'needs_restore'):m.configured?'ready':'idle';
  const copy={
    idle:{lamp:'off',title:'日常环境',detail:'在下方选择目标地区，预览差异后一键切换。'},
    ready:{lamp:'info',title:diff?`将修改 ${diff} 项设置`:'当前已与目标一致',detail:diff?'修改前自动备份原值，逐项回读验证；失败会自动回滚。':'无需修改，也可以直接切换以记录本次会话。'},
    active:{lamp:'warn',title:'目标环境生效中',detail:'退出应用不会自动恢复。用完 Claude 后，请回来点击“恢复原环境”。'},
    needs_restore:{lamp:'bad',title:'需要恢复原环境',detail:'检测到未完成操作或外部修改。请先恢复，再开始新的切换。'},
  }[phase];
  const action=m.busy?<Button size="lg" disabled><LoaderCircle className="animate-spin"/>请稍候</Button>
    :m.active?<Button ref={restoreButton} size="lg" variant={phase==='needs_restore'?'default':'outline'} disabled={m.disabled} onClick={m.restore}><RotateCcw/>恢复原环境</Button>
    :<Button size="lg" ripple disabled={m.disabled||!m.configured} onClick={m.switchEnvironment}><ArrowLeftRight/>切换环境</Button>;
  return <Card aria-labelledby="status-title">
    <CardHeader>
      <SectionHeader flush title={<span id="status-title" className="status-title"><StatusLamp tone={m.busy?'info':copy.lamp as 'off'}/>{m.busy?'正在处理':copy.title}</span>} description={m.busy?m.progress||'正在处理…':copy.detail} actions={action}/>
    </CardHeader>
    <Table className="environment-table" aria-label="当前环境与目标环境对照">
      <TableHeader><TableRow><TableHead>配置项</TableHead><TableHead>{m.active?'切换前':'当前环境'}</TableHead><TableHead>目标环境</TableHead><TableHead>状态</TableHead></TableRow></TableHeader>
      <TableBody>{fields.map(f=>{
        const Icon=icons[f];const now=current?.[f]?.label;const next=target[f];const changed=!!next&&now!==next;
        const before=m.active&&journal?journal.original?.[f]?.label:now;
        return <TableRow key={f}>
          <TableCell><span className="setting-label"><Icon size={15}/>{labels[f]}</span></TableCell>
          <TableCell><span className="setting-value" title={before}>{m.label(f,before)}</span></TableCell>
          <TableCell><span className="setting-value" title={next}>{next?m.label(f,next):'待选择'}</span></TableCell>
          <TableCell><Badge variant="soft" tone={!next?'neutral':m.active?'warn':changed?'info':'ok'}>{!next?'待配置':m.active?'已生效':changed?'将修改':'不变'}</Badge></TableCell>
        </TableRow>;
      })}</TableBody>
    </Table>
  </Card>;
}

function Launchers({m}:{m:EnvironmentModel}) {
  const {state,journal}=m;
  const ready=journal?.status==='active';
  return <Card>
    <CardContent>
      <SectionHeader title="启动 Claude" description={ready?'环境已就绪，在新环境中打开客户端。':'切换环境后再启动，客户端才会读取新设置。'}/>
      <div className="launch-grid">{(['cli','desktop'] as const).map(client=>{
        const cli=client==='cli';const Icon=cli?Terminal:Monitor;
        const path=cli?state?.config.cliPath||m.clients?.cli:state?.config.desktopPath||m.clients?.desktop;
        return <Card key={client} variant="inset"><CardContent className="launch-item">
          <span className="client-icon mg-well"><Icon size={18}/></span>
          <div className="launch-text"><strong>{cli?'Claude Code':'Claude Desktop'}</strong><code title={path||''}>{path||`未检测到 ${cli?'CLI':'Desktop'}，可在设置中指定`}</code></div>
          <Button size="sm" variant={ready?'default':'outline'} disabled={m.disabled||(m.active&&!ready)} onClick={()=>m.launch(client)}><ArrowUpRight/>{cli?'启动 CLI':'启动 Desktop'}</Button>
        </CardContent></Card>;
      })}</div>
    </CardContent>
  </Card>;
}

export function Home({m,restoreButton}:{m:EnvironmentModel;restoreButton:React.RefObject<HTMLButtonElement|null>}) {
  if(!m.state)return null;
  return <div className="page-stack">
    {m.state.demo&&<Notice>模拟模式：所有切换只写入独立演示文件，不修改系统设置，也不启动 Claude。</Notice>}
    <StatusCard m={m} restoreButton={restoreButton}/>
    <TargetPicker m={m}/>
    <Launchers m={m}/>
    <p className="scope-note"><ShieldCheck size={14}/><span>仅修改系统时区、区域格式与国家 / 地区，会影响本机其他应用。切换和恢复前请保存并关闭 Claude；本工具不承诺避免账号封禁。{isMac()?'修改时区需要管理员授权。':'修改时区会请求 UAC 授权。'}</span></p>
  </div>;
}
