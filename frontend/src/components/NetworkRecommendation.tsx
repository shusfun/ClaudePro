import {Globe2, Sparkles} from 'lucide-react';
import {Button, Notice} from '@claudepro/ui';
import {fields, type EnvironmentModel} from '../useEnvironment';

export function detectLabel(m:EnvironmentModel,idle='检测并推荐地区') {
  return m.networkBusy?'检测中…':m.networkCooldown>0?`${m.networkCooldown} 秒后可刷新`:m.networkReport?'重新检测并推荐':idle;
}

export function DetectButton({m,idle,variant='outline'}:{m:EnvironmentModel;idle?:string;variant?:'outline'|'ghost'}) {
  return <Button variant={variant} size="sm" disabled={m.networkBusy||m.networkCooldown>0} onClick={()=>void m.checkNetwork()}><Globe2/>{detectLabel(m,idle)}</Button>;
}

/** 出口检测结果：一行结论 + 可选的采用按钮，细节折叠。 */
export function NetworkRecommendation({m}:{m:EnvironmentModel}) {
  const n=m.networkReport;
  const preset=n?.status==='ok'?m.presets.find(p=>p.id===n.recommendedPreset):undefined;
  const selected=preset&&fields.every(f=>preset.target[f]===m.draft.target[f]);
  if(m.networkError)return <Notice tone="warn">未能检测：{m.networkError}</Notice>;
  if(!n)return null;
  if(n.status!=='ok')return <Notice tone="warn">{n.message}；等待冷却后可重试，也可手动选择模板。</Notice>;
  return <Notice tone={n.locationConflict?'warn':'info'}>
    <div className="recommendation-heading">
      <div>
        <strong>{n.locationConflict?'检测来源存在分歧':preset?'推荐：'+preset.name:'暂不能推荐现有模板'}</strong>
        <p className="detection-note">{n.ip} · {[n.city,n.region,n.country].filter(Boolean).join(' / ')} · {n.timezone||'时区未知'}</p>
        <p>{n.recommendationReason}{preset?.unavailable?' '+preset.unavailable:''}</p>
      </div>
      {preset&&<Button size="sm" disabled={m.disabled||!m.state||m.active||!!preset.unavailable||!!selected} onClick={()=>m.setDraft(d=>({...d,target:{...preset.target}}))}><Sparkles/>{selected?'已采用推荐':'采用推荐'}</Button>}
    </div>
    <details className="network-details"><summary>查看检测来源与状态</summary>
      <p className="detection-note">{new Date(n.time).toLocaleString()} · {n.source} · {n.proxyMode}{n.cached?' · 短时缓存结果':''}</p>
      {n.notes&&<p className="detection-note">{n.notes}</p>}
      <p className="detection-note">IP 位置是近似结果，仅代表本次请求；更换代理后请重新检测。采用推荐只更新预览，点击切换按钮才生效。</p>
    </details>
  </Notice>;
}
