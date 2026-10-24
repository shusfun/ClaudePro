import {Globe2, Sparkles} from 'lucide-react';
import {Button, Notice} from '@claudepro/ui';
import {fields, type EnvironmentModel} from '../useEnvironment';

export function NetworkRecommendation({m,detect=true}:{m:EnvironmentModel;detect?:boolean}) {
  const n=m.networkReport;
  const preset=n?.status==='ok'?m.presets.find(p=>p.id===n.recommendedPreset):undefined;
  const selected=preset&&fields.every(f=>preset.target[f]===m.draft.target[f]);
  return <div className="recommendation">
    {detect&&<div className="recommendation-heading"><span>不确定选哪个地区？根据当前出口推荐。</span><Button variant="outline" size="sm" disabled={m.networkBusy||m.networkCooldown>0} onClick={()=>void m.checkNetwork()}><Globe2/>{m.networkBusy?'检测中…':m.networkCooldown>0?`${m.networkCooldown} 秒后可刷新`:n?'重新检测并推荐':'检测并推荐地区'}</Button></div>}
    {m.networkError&&<Notice tone="warn">未能检测：{m.networkError}</Notice>}
    {n&&(n.status==='ok'?<Notice tone={n.locationConflict?'warn':'info'}>
      <div className="recommendation-heading"><div><strong>{n.locationConflict?'检测来源存在分歧':preset?'推荐：'+preset.name:'暂不能推荐现有模板'}</strong><p className="detection-note">{n.ip} · {[n.city,n.region,n.country].filter(Boolean).join(' / ')} · {n.timezone||'时区未知'}</p><p>{n.recommendationReason}</p></div>
      {preset&&<Button size="sm" disabled={m.disabled||!m.state||m.active||!!preset.unavailable||!!selected} onClick={()=>m.setDraft(d=>({...d,target:{...preset.target}}))}><Sparkles/>{selected?'已采用推荐':'采用推荐'}</Button>}</div>
      {preset?.unavailable&&<p>{preset.unavailable}</p>}
      {m.active&&<p>当前环境尚未恢复，恢复后可采用其他模板。</p>}
      <p className="detection-note">{new Date(n.time).toLocaleString()} · {n.source} · {n.proxyMode}。IP 位置是近似结果，仅代表本次请求；更换代理后请重新检测。采用推荐只更新预览，点击「切换环境」才生效。</p>
    </Notice>:<Notice tone="warn">{n.message}；等待冷却后可重试，也可手动选择模板。</Notice>)}
    {n?.notes&&<details className="network-details"><summary>查看检测来源与状态</summary><p className="detection-note">{n.notes}</p></details>}
    {n?.cached&&<p className="detection-note">显示短时缓存结果；到期后可手动刷新。</p>}
    {detect&&!n&&!m.networkError&&<p className="detection-note">点击后使用 GeoJS、ipwho.is、ipinfo 等来源查询出口 IP、城市及地理时区。不会自动更改目标或系统设置。</p>}
  </div>;
}
