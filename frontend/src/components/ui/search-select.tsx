// ClaudePro extension using Base UI and the local MiGate input material.
import {Combobox} from '@base-ui/react/combobox';
import {Check, ChevronDown} from 'lucide-react';
import {fieldClass} from './input';

type Option = {id:string;name:string};
export function SearchSelect({id,label,options,value,disabled,onChange}:{id:string;label:string;options:Option[];value:string;disabled:boolean;onChange:(value:string)=>void}) {
  return <Combobox.Root items={options} value={options.find(o=>o.id===value)||null} disabled={disabled}
    itemToStringLabel={(o:Option)=>o.name} itemToStringValue={(o:Option)=>o.id}
    isItemEqualToValue={(a:Option,b:Option)=>a.id===b.id}
    filter={(o:Option,query:string)=>(o.name+' '+o.id).toLocaleLowerCase().includes(query.trim().toLocaleLowerCase())}
    onValueChange={o=>{if(o)onChange(o.id);}}>
    <Combobox.InputGroup className="relative">
      <Combobox.Input id={id} aria-label={label} placeholder={'搜索'+label+' / 代码'} className={fieldClass+' h-9.5 py-2 pr-9'}/>
      <Combobox.Trigger aria-label={'展开'+label} className="absolute right-1 top-1 flex h-7 w-7 items-center justify-center text-[var(--mg-ink-3)]"><ChevronDown size={16}/></Combobox.Trigger>
    </Combobox.InputGroup>
    <Combobox.Portal><Combobox.Positioner sideOffset={5} className="z-50 max-w-[calc(100vw-32px)]">
      <Combobox.Popup className="w-[var(--anchor-width)] max-w-[var(--available-width)] rounded-[var(--mg-radius)] border border-[var(--mg-line)] bg-popover p-1 text-popover-foreground shadow-lg">
        <Combobox.Empty className="p-4 text-sm text-[var(--mg-ink-3)]">没有匹配的选项，请尝试名称或代码</Combobox.Empty>
        <Combobox.List className="max-h-[min(360px,var(--available-height))] overflow-y-auto overscroll-contain">
          {(o:Option)=><Combobox.Item key={o.id} value={o} className="flex cursor-default items-center justify-between gap-2 rounded-[var(--mg-radius-sm)] px-3 py-2 text-[13px] break-words outline-none data-highlighted:bg-[var(--mg-accent-soft)] data-highlighted:text-[var(--mg-accent-soft-ink)]"><span>{o.name}</span><Combobox.ItemIndicator><Check size={15}/></Combobox.ItemIndicator></Combobox.Item>}
        </Combobox.List>
      </Combobox.Popup>
    </Combobox.Positioner></Combobox.Portal>
  </Combobox.Root>;
}
