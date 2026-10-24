import {Browser} from '@wailsio/runtime';
import {ArrowUpRight, RefreshCw} from 'lucide-react';
import {Badge, Button, Card, CardContent, Notice, PageHeader, PageState, SectionHeader} from '@claudepro/ui';
import {fields,labels,type EnvironmentModel} from '../useEnvironment';
import {DetectButton, NetworkRecommendation} from './NetworkRecommendation';

const statusText:Record<string,string>={ok:'已确认',info:'信息',warn:'需查看',unknown:'未能确认',error:'异常'};
const actionText:Record<string,string>={settings:'设置',history:'查看记录',restore:'恢复原环境',refresh:'重新检查'};

export function Diagnostics({m}:{m:EnvironmentModel}) {
  const r=m.localReport, n=m.networkReport;
  const checks=r?.checks||[];
  const region=m.catalog?.regions?.find(o=>o.id===m.draft.target.region);
  const country=m.presets.find(p=>p.target.region===m.draft.target.region)?.country||region?.name.match(/\(([A-Z]{2})\)$/)?.[1];
  const action=(name:string)=>{
    if(name==='restore')void m.restore();else if(name==='refresh')void m.checkLocal();else if(name==='history')m.setTab('history');
    else if(['cli','desktop'].includes(name))m.setSettingsOpen(true);else m.setTab('home');
  };
  const open=(url:string)=>{void Browser.OpenURL(url).catch(e=>m.setError(String(e)));};
  return <div className="page-stack">
    <PageHeader eyebrow="CLAUDEPRO / DIAGNOSTICS" title="环境检测" description="本机检查自动进行，网络出口检测需要手动触发。"/>
    <Card><CardContent>
      <SectionHeader title="本机环境" description={r?'检测于 '+new Date(r.time).toLocaleString():'读取设置、客户端与恢复状态；只读检查'}
        actions={<Button variant="outline" size="sm" disabled={m.busy||m.loading||m.localBusy} onClick={()=>void m.checkLocal()}><RefreshCw/>{m.localBusy?'检查中…':'重新检查'}</Button>}/>
      {m.localError&&<Notice tone="warn">本机检查未完成：{m.localError}</Notice>}
      <div className="check-list">{checks.map(c=>{
        const field=fields.find(f=>f===c.id);
        const current=field?m.state?.view.current?.[field]?.label:undefined;
        const same=field&&current&&m.draft.target[field]===current;
        const status=field&&current?(same?'ok':'info'):c.status;
        const detail=field&&current?`当前：${m.label(field,current)}；目标：${m.label(field,m.draft.target[field])}${same?'（一致）':'（待切换，不代表故障）'}`:c.detail;
        const act=['cli','desktop'].includes(c.id)?c.id:c.action;
        return <div className="check-row" key={c.id}>
          <div><strong>{field?labels[field]:c.name}</strong><p>{detail}</p></div>
          <Badge variant="soft" tone={status==='ok'?'ok':status==='error'?'bad':['unknown','warn'].includes(status)?'warn':'neutral'}>{statusText[status]}</Badge>
          <Button variant="ghost" size="sm" disabled={m.busy||m.loading||m.localBusy} onClick={()=>action(act)}>{actionText[c.action]}</Button>
        </div>;
      })}</div>
    </CardContent></Card>
    <Card><CardContent>
      <SectionHeader title="网络出口" description="多源查询并自动回退；仅代表本次请求的出口，不代表 Claude 的实际路由。" actions={<DetectButton m={m} idle="检测网络出口"/>}/>
      <NetworkRecommendation m={m}/>
      {n?.status==='ok'&&<>
        <dl className="network-facts"><div><dt>出口 IP</dt><dd>{n.ip}</dd></div><div><dt>国家 / 地区</dt><dd>{n.country} ({n.countryCode})</dd></div><div><dt>ASN</dt><dd>{n.asn}</dd></div><div><dt>运营商</dt><dd>{n.organization}</dd></div></dl>
        <p className="detection-note">{n.locationConflict?'检测来源存在分歧，请先核对来源详情。':!country?'选择目标地区后可比较出口国家。':country===n.countryCode?'本次检测出口国家与所选目标一致。':'本次检测出口国家与所选目标不同；请自行核对代理分流。'} 该结果不判断账号安全或服务资格。</p>
      </>}
      {!n&&!m.networkBusy&&!m.networkError&&<PageState size="sm" kind="empty" title="尚未联网检测。结果仅保留在本次应用会话。"/>}
      <p className="detection-note">使用 Go 环境变量代理规则，不模拟系统 PAC 或 Claude 专属代理。检测服务会接收到此请求的出口 IP。</p>
    </CardContent></Card>
    <Card><CardContent>
      <SectionHeader title="外部检测工具" description="在默认浏览器打开；浏览器语言、WebRTC、指纹等结果不代表 CLI 或 Desktop。"/>
      <div className="button-row"><Button variant="outline" size="sm" onClick={()=>open('https://checkcc.org')}><ArrowUpRight/>check-cc</Button><Button variant="outline" size="sm" onClick={()=>open('https://ipcheck.ing')}><ArrowUpRight/>MyIP</Button></div>
    </CardContent></Card>
  </div>;
}
