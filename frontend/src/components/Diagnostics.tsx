import {Browser} from '@wailsio/runtime';
import {Activity, ArrowUpRight, Globe2, RefreshCw} from 'lucide-react';
import {Badge, Button, Card, CardContent, Notice, SectionHeader} from '@claudepro/ui';
import {fields,labels,type EnvironmentModel} from '../useEnvironment';
import {NetworkRecommendation} from './NetworkRecommendation';

export function Diagnostics({m,summary=false}:{m:EnvironmentModel;summary?:boolean}) {
  const r=m.localReport, n=m.networkReport;
  const checks=r?.checks||[];
  const concerns=checks.filter(c=>['warn','error','unknown'].includes(c.status));
  const region=m.catalog?.regions?.find(o=>o.id===m.draft.target.region);
  const country=m.presets.find(p=>p.target.region===m.draft.target.region)?.country||region?.name.match(/\(([A-Z]{2})\)$/)?.[1];
  const action=(name:string)=>{
    if(name==='restore')void m.restore();else if(name==='refresh')void m.checkLocal();else m.setTab(name==='history'?'history':'settings');
  };
  const open=(url:string)=>{void Browser.OpenURL(url).catch(e=>m.setError(String(e)));};
  return <Card><CardContent className="diagnostics-content">
    <SectionHeader title={summary?'环境检测摘要':'本机环境检查'} description={r?'检测于 '+new Date(r.time).toLocaleString():'读取设置、客户端与恢复状态；只读检查'} actions={summary?<Button variant="ghost" size="sm" onClick={()=>m.setTab('diagnostics')}><Activity/>查看详情</Button>:<Button variant="outline" size="sm" disabled={m.busy||m.loading||m.localBusy} onClick={()=>void m.checkLocal()}><RefreshCw/>{m.localBusy?'检查中…':'重新检查'}</Button>}/>
    {m.localError&&<Notice tone="warn">本机检查未完成：{m.localError}</Notice>}
    {summary?<p>{!r?'正在检查本机环境…':concerns.length?`${concerns.length} 项需要查看，可在检测详情中处理。`:'本机检查完成，没有发现需要处理的项目。'} {n?.status==='ok'?`最近检测出口：${n.country}`:'网络尚未检测或未能完成。'}</p>:<>
      <div className="check-list">{checks.map(c=>{
        const field=fields.find(f=>f===c.id);
        const current=field?m.state?.view.current?.[field]?.label:undefined;
        const same=field&&current&&m.draft.target[field]===current;
        const status=field&&current?(same?'ok':'info'):c.status;
        const detail=field&&current?`当前：${m.label(field,current)}；目标：${m.label(field,m.draft.target[field])}${same?'（一致）':'（待切换，不代表故障）'}`:c.detail;
        return <div className="check-row" key={c.id}><div><strong>{field?labels[field]:c.name}</strong><p>{detail}</p></div><Badge variant="soft" tone={status==='ok'?'ok':status==='error'?'bad':['unknown','warn'].includes(status)?'warn':'neutral'}>{({ok:'已确认',info:'信息',warn:'需查看',unknown:'未能确认',error:'异常'} as Record<string,string>)[status]}</Badge><Button variant="ghost" size="sm" disabled={m.busy||m.loading||m.localBusy} onClick={()=>action(c.action)}>{({settings:'配置',history:'查看记录',restore:'恢复原环境',refresh:'重新检查'} as Record<string,string>)[c.action]}</Button></div>;
      })}</div>
      <div className="settings-divider"/>
      <SectionHeader title="网络出口检测" description="多源查询并自动回退；仅本次检测请求的出口，不代表 Claude 的实际路由。" actions={<Button variant="outline" disabled={m.networkBusy||m.networkCooldown>0} onClick={()=>void m.checkNetwork()}><Globe2/>{m.networkBusy?'检测中…':m.networkCooldown>0?`${m.networkCooldown} 秒后可刷新`:'检测网络出口'}</Button>}/>
      <p className="detection-note">使用 Go 环境变量代理规则，不自动模拟 Windows PAC 或 Claude 专属代理。检测服务会接收到此请求的出口 IP。</p>
      <NetworkRecommendation m={m} detect={false}/>
      {n&&<><p className="detection-note">{n.source} · {new Date(n.time).toLocaleString()} · {n.proxyMode}</p>{n.status==='ok'?<>
        <dl className="network-facts"><div><dt>出口 IP</dt><dd>{n.ip}</dd></div><div><dt>国家 / 地区</dt><dd>{n.country} ({n.countryCode})</dd></div><div><dt>ASN</dt><dd>{n.asn}</dd></div><div><dt>运营商</dt><dd>{n.organization}</dd></div></dl>
        <Notice tone={country&&country!==n.countryCode?'warn':'info'}>{n.locationConflict?'检测来源存在分歧，请先核对来源详情。':!country?'选择目标地区后可比较出口国家。':country===n.countryCode?'本次检测出口国家与所选目标一致。':'本次检测出口国家与所选目标不同；请自行核对代理分流。'} 该结果不判断账号安全或服务资格。</Notice>
      </>:null}</>}
      {!n&&!m.networkBusy&&!m.networkError&&<p>尚未联网检测。结果仅保留在本次应用会话。</p>}
      <div className="settings-divider"/>
      <SectionHeader title="此前调研的检测工具" description="在默认浏览器打开；浏览器语言、WebRTC、指纹等结果不代表 CLI 或 Desktop。"/>
      <div className="button-row"><Button variant="outline" onClick={()=>open('https://checkcc.org')}><ArrowUpRight/>check-cc</Button><Button variant="outline" onClick={()=>open('https://ipcheck.ing')}><ArrowUpRight/>MyIP</Button></div>
      <p className="detection-note">检测展示可观察事实，不生成封号风险分，也不承诺避免账号封禁。</p>
    </>}
  </CardContent></Card>;
}
