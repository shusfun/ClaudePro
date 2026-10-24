import {useState} from 'react';
import {Check, Globe2, SlidersHorizontal} from 'lucide-react';
import {Badge, Button, Field, SectionHeader} from '@claudepro/ui';
import {SearchSelect} from './ui/search-select';
import {NetworkRecommendation} from './NetworkRecommendation';
import {fields, labels, type EnvironmentModel} from '../useEnvironment';

export function TargetPicker({m}:{m:EnvironmentModel}) {
  const [expanded,setExpanded]=useState(false);
  const selected=m.presets.find(p=>fields.every(f=>p.target[f]===m.draft.target[f]));
  const custom=expanded||(!selected&&fields.some(f=>!!m.draft.target[f]));
  return <div className="target-picker">
    <SectionHeader title="1. 选择地区" description="选一个模板，三项配置自动配齐。预览后点击「切换环境」生效。" actions={<Badge variant="soft" tone="neutral">{m.active?'恢复后可更换':selected?.name||'自定义 / 尚未选择'}</Badge>}/>
    <NetworkRecommendation m={m}/>
    <div className="preset-grid" role="group" aria-label="地区模板">{m.presets.map(p=><div key={p.id} className="preset-wrap">
      <Button variant="outline" className={'preset-card'+(selected?.id===p.id?' is-pressed':'')} aria-pressed={selected?.id===p.id} aria-label={p.name+'模板'} disabled={m.disabled||m.active||!!p.unavailable} onClick={()=>{m.setDraft(d=>({...d,target:{...p.target}}));setExpanded(false);}}>
        <span className="preset-country"><Globe2 size={16}/>{p.country}{selected?.id===p.id&&<Check size={14}/>}</span><strong>{p.name}{m.networkReport?.recommendedPreset===p.id&&<span className="preset-recommended"> · 出口推荐</span>}</strong><small>{p.target.culture}</small>
      </Button>{p.unavailable&&<small className="preset-unavailable">{p.unavailable}</small>}
    </div>)}</div>
    {!m.presets.length&&<p>正在读取地区模板；也可展开自定义选择本机支持的选项。</p>}
    <Button variant="ghost" size="sm" aria-expanded={custom} disabled={m.disabled||m.active} onClick={()=>setExpanded(v=>!v)}><SlidersHorizontal/>自定义区域设置</Button>
    {custom&&<div className="target-fields">{(['region','timezone','culture'] as const).map(f=>{
      const options=f==='timezone'?m.catalog?.timezones:f==='culture'?m.catalog?.cultures:m.catalog?.regions;
      return <Field key={f} label={labels[f]} htmlFor={'target-'+f}><SearchSelect id={'target-'+f} label={labels[f]} options={options||[]} value={m.draft.target[f]} disabled={m.disabled||m.active||!m.catalog} onChange={v=>m.setDraft(d=>({...d,target:{...d.target,[f]:v}}))}/></Field>;
    })}</div>}
  </div>;
}
