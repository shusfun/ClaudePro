import {useState} from 'react';
import {Check, Globe2, Lock, SlidersHorizontal} from 'lucide-react';
import {Badge, Button, Card, CardContent, Field, SectionHeader} from '@claudepro/ui';
import {SearchSelect} from './ui/search-select';
import {DetectButton, NetworkRecommendation} from './NetworkRecommendation';
import {fields, labels, type EnvironmentModel} from '../useEnvironment';

export function TargetPicker({m}:{m:EnvironmentModel}) {
  const [expanded,setExpanded]=useState(false);
  const selected=m.presets.find(p=>fields.every(f=>p.target[f]===m.draft.target[f]));
  const custom=expanded||(!selected&&fields.some(f=>!!m.draft.target[f]));
  const locked=m.disabled||m.active;
  return <Card><CardContent className="target-picker">
    <SectionHeader flush title="选择目标地区" description={m.active?'环境生效期间已锁定，恢复后可更换。':'选一个模板，三项设置自动配齐。'}
      actions={m.active?<Badge variant="soft" tone="neutral"><Lock size={12}/>已锁定</Badge>:<DetectButton m={m} variant="ghost"/>}/>
    {!m.active&&<NetworkRecommendation m={m}/>}
    <div className="preset-grid" role="group" aria-label="地区模板">{m.presets.map(p=>{
      const on=selected?.id===p.id;
      const recommended=m.networkReport?.status==='ok'&&m.networkReport.recommendedPreset===p.id&&!m.networkReport.locationConflict;
      return <div key={p.id} className="preset-wrap">
        <Button variant="outline" className={'preset-card'+(on?' is-pressed':'')} aria-pressed={on} aria-label={p.name+'模板'} disabled={locked||!!p.unavailable}
          onClick={()=>{m.setDraft(d=>({...d,target:{...p.target}}));setExpanded(false);}}>
          <span className="preset-country"><Globe2 size={14}/>{p.country}{on&&<Check size={14}/>}</span>
          <strong>{p.name}{recommended&&<Badge className="preset-tag" variant="soft" tone="info">推荐</Badge>}</strong>
          <small>{p.target.timezone}</small>
        </Button>
        {p.unavailable&&<small className="preset-unavailable">{p.unavailable}</small>}
      </div>;
    })}</div>
    {!m.presets.length&&<p>正在读取地区模板…</p>}
    <Button className="justify-self-start" variant="ghost" size="sm" aria-expanded={custom} disabled={locked} onClick={()=>setExpanded(v=>!v)}><SlidersHorizontal/>自定义区域设置</Button>
    {custom&&<div className="target-fields">{(['region','timezone','culture'] as const).map(f=>{
      const options=f==='timezone'?m.catalog?.timezones:f==='culture'?m.catalog?.cultures:m.catalog?.regions;
      return <Field key={f} label={labels[f]} htmlFor={'target-'+f}><SearchSelect id={'target-'+f} label={labels[f]} options={options||[]} value={m.draft.target[f]} disabled={locked||!m.catalog} onChange={v=>m.setDraft(d=>({...d,target:{...d.target,[f]:v}}))}/></Field>;
    })}</div>}
  </CardContent></Card>;
}
